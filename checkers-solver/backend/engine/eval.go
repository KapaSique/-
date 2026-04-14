package engine

import "math/bits"

// Evaluation constants
const (
	PawnValue = 100
	KingValue = 350

	CenterBonus       = 10
	BackRankBonus     = 15
	AdvancedPawnBonus = 5
	EdgePenalty       = -8
	TempoBonus        = 3
	KingMobility      = 12

	WinScore  = 100000
	DrawScore = 0
	Infinity  = 999999
)

// Center squares: d4(17), e5(14), d5(13), e4(18) — adjusted to bitboard indices
// In our numbering: row 3 cols d,e = squares 13,14; row 4 cols d,e = squares 17,18
var centerMask uint32 = (1 << 13) | (1 << 14) | (1 << 17) | (1 << 18)

// Edge columns (a and h files)
// a-file dark squares: 4, 12, 20, 28
// h-file dark squares: 3, 11, 19, 27 (wait, let me recalculate)
// Actually, from SquareToRowCol:
// sq 4: row=1, col=0 (a7) — edge
// sq 12: row=3, col=0 (a5) — edge
// sq 20: row=5, col=0 (a3) — edge
// sq 28: row=7, col=0 (a1) — edge
// sq 3: row=0, col=7 (h8) — edge
// sq 7: row=1, col=6 (g7) — not h-file, but near edge
// Actually h-file: col=7
// sq 3: row=0, col=7 — yes h8
// sq 11: row=2, col=7 — h6
// sq 19: row=4, col=7 — h4
// sq 27: row=6, col=7 — h2
var edgeMask uint32 = (1 << 4) | (1 << 12) | (1 << 20) | (1 << 28) |
	(1 << 3) | (1 << 11) | (1 << 19) | (1 << 27)

// Back rank masks
// White back rank: row 7 = squares 28,29,30,31
var whiteBackRank uint32 = 0xF0000000
// Black back rank: row 0 = squares 0,1,2,3
var blackBackRank uint32 = 0x0000000F

// Evaluate returns a score for the given board from the perspective of `perspective`.
// Positive = good for perspective, negative = bad.
func Evaluate(b Board, perspective Color) int {
	score := evaluateInternal(b)
	if perspective == Black {
		score = -score
	}
	return score
}

// evaluateInternal returns score from White's perspective.
func evaluateInternal(b Board) int {
	whiteMen := b.White
	blackMen := b.Black
	whiteKings := b.WhiteKing
	blackKings := b.BlackKing

	wMenCount := bits.OnesCount32(whiteMen)
	bMenCount := bits.OnesCount32(blackMen)
	wKingCount := bits.OnesCount32(whiteKings)
	bKingCount := bits.OnesCount32(blackKings)

	// No pieces left = loss
	wTotal := wMenCount + wKingCount
	bTotal := bMenCount + bKingCount
	if wTotal == 0 {
		return -WinScore
	}
	if bTotal == 0 {
		return WinScore
	}

	score := 0

	// Material
	score += (wMenCount - bMenCount) * PawnValue
	score += (wKingCount - bKingCount) * KingValue

	// Center control
	wCenter := bits.OnesCount32((whiteMen | whiteKings) & centerMask)
	bCenter := bits.OnesCount32((blackMen | blackKings) & centerMask)
	score += (wCenter - bCenter) * CenterBonus

	// Edge penalty
	wEdge := bits.OnesCount32((whiteMen | whiteKings) & edgeMask)
	bEdge := bits.OnesCount32((blackMen | blackKings) & edgeMask)
	score += (wEdge - bEdge) * EdgePenalty

	// Back rank bonus
	wBack := bits.OnesCount32(whiteMen & whiteBackRank)
	bBack := bits.OnesCount32(blackMen & blackBackRank)
	score += (wBack - bBack) * BackRankBonus

	// Advanced pawn bonus — reward pawns closer to promotion
	score += advancedPawnScore(whiteMen, White) - advancedPawnScore(blackMen, Black)

	// King mobility
	if wKingCount > 0 || bKingCount > 0 {
		score += kingMobilityScore(b, White) - kingMobilityScore(b, Black)
	}

	// Endgame adjustments: when few pieces, kings become even more valuable
	totalPieces := wTotal + bTotal
	if totalPieces <= 8 {
		score += (wKingCount - bKingCount) * 50 // extra king bonus in endgame
	}

	return score
}

// advancedPawnScore rewards pawns that are closer to the promotion row.
func advancedPawnScore(men uint32, color Color) int {
	score := 0
	m := men
	for m != 0 {
		sq := popLSB(&m)
		row := sq / 4
		var advance int
		if color == White {
			advance = 7 - row // white advances toward row 0
		} else {
			advance = row // black advances toward row 7
		}
		score += advance * AdvancedPawnBonus
	}
	return score
}

// kingMobilityScore counts the number of squares a side's kings can reach.
func kingMobilityScore(b Board, color Color) int {
	kings := b.KingsOf(color)
	occupied := b.Occupied()
	score := 0
	for kings != 0 {
		sq := popLSB(&kings)
		for d := 0; d < 4; d++ {
			for _, next := range diagRays[sq][d] {
				if occupied&(uint32(1)<<next) != 0 {
					break
				}
				score += KingMobility
			}
		}
	}
	return score
}
