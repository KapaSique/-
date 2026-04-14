package engine

import "fmt"

// Move represents a single move or capture sequence.
type Move struct {
	From     int   // starting square
	To       int   // ending square
	Captures []int // squares of captured pieces (in order)
	Path     []int // full path of squares visited during capture sequence
	Promoted bool  // piece promoted to king during this move
}

// IsCapture returns true if this is a capture move.
func (m Move) IsCapture() bool {
	return len(m.Captures) > 0
}

// Notation returns the move in Russian draughts notation.
func (m Move) Notation() string {
	if m.IsCapture() {
		if len(m.Path) > 0 {
			parts := make([]string, len(m.Path))
			for i, sq := range m.Path {
				parts[i] = SquareToNotation(sq)
			}
			return joinWith(parts, ":")
		}
		return fmt.Sprintf("%s:%s", SquareToNotation(m.From), SquareToNotation(m.To))
	}
	return fmt.Sprintf("%s-%s", SquareToNotation(m.From), SquareToNotation(m.To))
}

func joinWith(parts []string, sep string) string {
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}

// --- Direction helpers ---

// For square sq on the board, get neighbors in each diagonal direction.
// Returns -1 if off-board.

// diagNeighbors[sq][dir] gives the adjacent square in direction dir.
// dir: 0=NW, 1=NE, 2=SW, 3=SE
var diagNeighbors [32][4]int

// diagRays[sq][dir] gives all squares along diagonal direction from sq (exclusive).
var diagRays [32][4][]int

func init() {
	// Initialize neighbor table
	for sq := 0; sq < 32; sq++ {
		row, col := SquareToRowCol(sq)
		dirs := [4][2]int{
			{-1, -1}, // NW
			{-1, 1},  // NE
			{1, -1},  // SW
			{1, 1},   // SE
		}
		for d, delta := range dirs {
			nr, nc := row+delta[0], col+delta[1]
			n := RowColToSquare(nr, nc)
			diagNeighbors[sq][d] = n
		}
	}

	// Initialize ray table (for king moves)
	for sq := 0; sq < 32; sq++ {
		row, col := SquareToRowCol(sq)
		dirs := [4][2]int{
			{-1, -1}, // NW
			{-1, 1},  // NE
			{1, -1},  // SW
			{1, 1},   // SE
		}
		for d, delta := range dirs {
			var ray []int
			r, c := row+delta[0], col+delta[1]
			for r >= 0 && r <= 7 && c >= 0 && c <= 7 {
				s := RowColToSquare(r, c)
				if s >= 0 {
					ray = append(ray, s)
				}
				r += delta[0]
				c += delta[1]
			}
			diagRays[sq][d] = ray
		}
	}
}

// --- Move generation ---

// GenerateMoves generates all legal moves for the side to move.
// If captures are available, only captures are returned (mandatory capture rule).
func GenerateMoves(b Board) []Move {
	captures := GenerateCaptures(b)
	if len(captures) > 0 {
		return captures
	}
	return generateQuietMoves(b)
}

// GenerateCaptures generates all capture moves for the side to move.
func GenerateCaptures(b Board) []Move {
	var moves []Move
	color := b.Turn
	men := b.MenOf(color)
	kings := b.KingsOf(color)

	// Men captures
	for men != 0 {
		sq := popLSB(&men)
		manCaptures := generateManCaptures(b, sq, color)
		moves = append(moves, manCaptures...)
	}

	// King captures
	for kings != 0 {
		sq := popLSB(&kings)
		kingCaptures := generateKingCaptures(b, sq, color)
		moves = append(moves, kingCaptures...)
	}

	return moves
}

// generateQuietMoves generates all non-capture moves.
func generateQuietMoves(b Board) []Move {
	var moves []Move
	color := b.Turn
	men := b.MenOf(color)
	kings := b.KingsOf(color)

	// Men quiet moves: forward diagonals only
	for men != 0 {
		sq := popLSB(&men)
		var forwardDirs []int
		if color == White {
			forwardDirs = []int{0, 1} // NW, NE (white moves up)
		} else {
			forwardDirs = []int{2, 3} // SW, SE (black moves down)
		}
		for _, d := range forwardDirs {
			to := diagNeighbors[sq][d]
			if to >= 0 && !isOccupied(b, to) {
				m := Move{From: sq, To: to}
				// Check promotion
				if uint32(1)<<to&PromotionRow(color) != 0 {
					m.Promoted = true
				}
				moves = append(moves, m)
			}
		}
	}

	// King quiet moves: any direction, any distance
	for kings != 0 {
		sq := popLSB(&kings)
		for d := 0; d < 4; d++ {
			for _, to := range diagRays[sq][d] {
				if isOccupied(b, to) {
					break // blocked
				}
				moves = append(moves, Move{From: sq, To: to})
			}
		}
	}

	return moves
}

// --- Man captures (with multi-capture / Turkish strike / mid-capture promotion) ---

func generateManCaptures(b Board, sq int, color Color) []Move {
	var results []Move
	captured := uint32(0)
	path := []int{sq}
	generateManCapturesRec(b, sq, color, captured, path, false, &results)
	return results
}

// generateManCapturesRec recursively builds capture sequences for a man.
// promoted tracks if the piece was promoted mid-capture (then it continues as king).
func generateManCapturesRec(b Board, sq int, color Color, captured uint32, path []int, promoted bool, results *[]Move) {
	opponent := color.Opponent()
	opponentPieces := b.PiecesOf(opponent)
	occupied := b.Occupied()
	found := false

	if promoted {
		// Continue as king
		generateKingCapturesRecFromMan(b, sq, color, captured, path, results)
		return
	}

	// Try all 4 directions (men can capture backwards in Russian draughts)
	for d := 0; d < 4; d++ {
		mid := diagNeighbors[sq][d]
		if mid < 0 {
			continue
		}
		midMask := uint32(1) << mid

		// Must be an opponent piece that hasn't been captured yet
		if opponentPieces&midMask == 0 || captured&midMask != 0 {
			continue
		}

		// The square beyond must be empty (or our starting position in the series)
		land := diagNeighbors[mid][d]
		if land < 0 {
			continue
		}
		landMask := uint32(1) << land
		// Landing square must be empty (pieces captured in this series are still on board
		// but we treat them as passable — NO, in Russian draughts captured pieces remain
		// until end but the capturing piece cannot land on an occupied square)
		if occupied&landMask != 0 && captured&landMask == 0 {
			// Square is occupied by a real piece — can't land
			continue
		}
		if occupied&landMask != 0 && captured&landMask != 0 {
			// Square has a captured piece — can still land? No, in Russian draughts
			// the capturing piece cannot land on a square occupied by a captured piece either.
			// Actually, captured pieces are NOT removed until the end of the sequence,
			// but the piece CAN pass over them... Let me reconsider.
			// In Russian draughts, captured pieces stay on the board during the capture sequence
			// and the capturing piece CANNOT land on them or pass through them for men.
			// For kings, the piece can fly over empty squares but not over captured pieces.
			// Actually, for the purpose of landing: the square must be truly empty.
			continue
		}

		found = true
		newCaptured := captured | midMask
		newPath := append(append([]int{}, path...), land)

		// Check if promotion happens at landing
		isPromoted := uint32(1)<<land&PromotionRow(color) != 0

		// Continue capturing from new position
		generateManCapturesRec(b, land, color, newCaptured, newPath, isPromoted, results)
	}

	if !found && len(path) > 1 {
		// End of capture sequence — record the move
		captures := bitmaskToSquares(captured)
		m := Move{
			From:     path[0],
			To:       sq,
			Captures: captures,
			Path:     append([]int{}, path...),
			Promoted: uint32(1)<<sq&PromotionRow(color) != 0,
		}
		*results = append(*results, m)
	}
}

// generateKingCapturesRecFromMan handles the case where a man promoted mid-capture
// and now continues capturing as a king.
func generateKingCapturesRecFromMan(b Board, sq int, color Color, captured uint32, path []int, results *[]Move) {
	opponent := color.Opponent()
	opponentPieces := b.PiecesOf(opponent)
	occupied := b.Occupied()
	found := false

	for d := 0; d < 4; d++ {
		ray := diagRays[sq][d]
		capturedInDir := -1

		for _, next := range ray {
			nextMask := uint32(1) << next

			if captured&nextMask != 0 {
				// Can't pass over a piece captured in this sequence
				break
			}

			if occupied&nextMask != 0 {
				if opponentPieces&nextMask != 0 && capturedInDir < 0 {
					// Found an opponent piece to capture
					capturedInDir = next
					continue
				}
				// Blocked by own piece or second opponent piece
				break
			}

			// Empty square
			if capturedInDir >= 0 {
				// Can land here after capturing
				found = true
				newCaptured := captured | (uint32(1) << capturedInDir)
				newPath := append(append([]int{}, path...), next)
				generateKingCapturesRecFromMan(b, next, color, newCaptured, newPath, results)
			}
		}
	}

	if !found && len(path) > 1 {
		captures := bitmaskToSquares(captured)
		m := Move{
			From:     path[0],
			To:       sq,
			Captures: captures,
			Path:     append([]int{}, path...),
			Promoted: true,
		}
		*results = append(*results, m)
	}
}

// --- King captures ---

func generateKingCaptures(b Board, sq int, color Color) []Move {
	var results []Move
	captured := uint32(0)
	path := []int{sq}
	generateKingCapturesRec(b, sq, color, captured, path, &results)
	return results
}

func generateKingCapturesRec(b Board, sq int, color Color, captured uint32, path []int, results *[]Move) {
	opponent := color.Opponent()
	opponentPieces := b.PiecesOf(opponent)
	occupied := b.Occupied()
	found := false

	for d := 0; d < 4; d++ {
		ray := diagRays[sq][d]
		capturedInDir := -1

		for _, next := range ray {
			nextMask := uint32(1) << next

			if captured&nextMask != 0 {
				// Can't pass through a piece already captured in this series
				break
			}

			if occupied&nextMask != 0 {
				if opponentPieces&nextMask != 0 && capturedInDir < 0 {
					// Found an opponent piece to potentially capture
					capturedInDir = next
					continue
				}
				// Blocked (own piece, or second opponent piece in this direction)
				break
			}

			// Empty square
			if capturedInDir >= 0 {
				// Can land here after capturing the piece
				found = true
				capMask := uint32(1) << capturedInDir
				newCaptured := captured | capMask
				newPath := append(append([]int{}, path...), next)

				// Try to continue capturing
				generateKingCapturesRec(b, next, color, newCaptured, newPath, results)
			}
		}
	}

	if !found && len(path) > 1 {
		// End of capture sequence
		captures := bitmaskToSquares(captured)
		m := Move{
			From:     path[0],
			To:       sq,
			Captures: captures,
			Path:     append([]int{}, path...),
		}
		*results = append(*results, m)
	}
}

// --- Apply move ---

// ApplyMove applies a move to the board and returns the new board state.
func ApplyMove(b Board, m Move) Board {
	nb := b.Clone()
	fromMask := uint32(1) << m.From
	toMask := uint32(1) << m.To
	color := b.Turn

	// Determine if the moving piece is a king
	isKing := false
	if color == White {
		if nb.WhiteKing&fromMask != 0 {
			isKing = true
			nb.WhiteKing &^= fromMask
			nb.WhiteKing |= toMask
		} else {
			nb.White &^= fromMask
			// Check promotion
			if toMask&PromotionRow(color) != 0 || m.Promoted {
				nb.WhiteKing |= toMask
			} else {
				nb.White |= toMask
			}
		}
	} else {
		if nb.BlackKing&fromMask != 0 {
			isKing = true
			nb.BlackKing &^= fromMask
			nb.BlackKing |= toMask
		} else {
			nb.Black &^= fromMask
			if toMask&PromotionRow(color) != 0 || m.Promoted {
				nb.BlackKing |= toMask
			} else {
				nb.Black |= toMask
			}
		}
	}
	_ = isKing

	// Remove captured pieces
	opponent := color.Opponent()
	for _, capSq := range m.Captures {
		capMask := uint32(1) << capSq
		if opponent == White {
			nb.White &^= capMask
			nb.WhiteKing &^= capMask
		} else {
			nb.Black &^= capMask
			nb.BlackKing &^= capMask
		}
	}

	nb.Turn = color.Opponent()
	return nb
}

// --- Utility ---

func isOccupied(b Board, sq int) bool {
	return b.Occupied()&(uint32(1)<<sq) != 0
}

// popLSB pops the least significant set bit and returns its index.
func popLSB(bits *uint32) int {
	b := *bits
	sq := 0
	if b == 0 {
		return -1
	}
	// Find index of LSB
	t := b & (-b)
	for t > 1 {
		t >>= 1
		sq++
	}
	*bits &^= uint32(1) << sq
	return sq
}

// bitmaskToSquares returns a list of set bit indices.
func bitmaskToSquares(mask uint32) []int {
	var squares []int
	m := mask
	for m != 0 {
		sq := popLSB(&m)
		squares = append(squares, sq)
	}
	return squares
}

// HasLegalMoves returns true if the current side has any legal moves.
func HasLegalMoves(b Board) bool {
	return len(GenerateMoves(b)) > 0
}
