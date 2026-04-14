package engine

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
)

// Color represents the side to move.
type Color int

const (
	White Color = iota
	Black
)

func (c Color) Opponent() Color {
	if c == White {
		return Black
	}
	return White
}

func (c Color) String() string {
	if c == White {
		return "W"
	}
	return "B"
}

// Board represents a Russian draughts position using bitboards.
// 32 dark squares numbered 0-31, left-to-right, top-to-bottom.
//
// Numbering (dark squares only):
//
//	  a  b  c  d  e  f  g  h
//	8 [  ][ 0][  ][ 1][  ][ 2][  ][ 3]
//	7 [ 4][  ][ 5][  ][ 6][  ][ 7][  ]
//	6 [  ][ 8][  ][ 9][  ][10][  ][11]
//	5 [12][  ][13][  ][14][  ][15][  ]
//	4 [  ][16][  ][17][  ][18][  ][19]
//	3 [20][  ][21][  ][22][  ][23][  ]
//	2 [  ][24][  ][25][  ][26][  ][27]
//	1 [28][  ][29][  ][30][  ][31][  ]
type Board struct {
	White     uint32 // bitmask of white men
	Black     uint32 // bitmask of black men
	WhiteKing uint32 // bitmask of white kings
	BlackKing uint32 // bitmask of black kings
	Turn      Color  // side to move
}

// Occupied returns a bitmask of all occupied squares.
func (b Board) Occupied() uint32 {
	return b.White | b.Black | b.WhiteKing | b.BlackKing
}

// WhiteAll returns all white pieces (men + kings).
func (b Board) WhiteAll() uint32 {
	return b.White | b.WhiteKing
}

// BlackAll returns all black pieces (men + kings).
func (b Board) BlackAll() uint32 {
	return b.Black | b.BlackKing
}

// PiecesOf returns all pieces for the given color.
func (b Board) PiecesOf(c Color) uint32 {
	if c == White {
		return b.WhiteAll()
	}
	return b.BlackAll()
}

// MenOf returns men (non-kings) for the given color.
func (b Board) MenOf(c Color) uint32 {
	if c == White {
		return b.White
	}
	return b.Black
}

// KingsOf returns kings for the given color.
func (b Board) KingsOf(c Color) uint32 {
	if c == White {
		return b.WhiteKing
	}
	return b.BlackKing
}

// Empty returns a bitmask of unoccupied dark squares.
func (b Board) Empty() uint32 {
	return ^b.Occupied() & 0xFFFFFFFF
}

// --- Square ↔ (row, col) conversion ---

// SquareToRowCol converts a square index (0-31) to (row, col) on the 8x8 board.
// row 0 = rank 8 (top), row 7 = rank 1 (bottom).
func SquareToRowCol(sq int) (int, int) {
	// Each row has 4 dark squares.
	row := sq / 4
	col := (sq % 4) * 2
	if row%2 == 0 {
		col++ // even rows: dark squares at cols 1,3,5,7
	}
	// odd rows: dark squares at cols 0,2,4,6
	return row, col
}

// RowColToSquare converts (row, col) to square index. Returns -1 if not a dark square.
func RowColToSquare(row, col int) int {
	if row < 0 || row > 7 || col < 0 || col > 7 {
		return -1
	}
	// Dark squares: (row+col) is odd
	if (row+col)%2 == 0 {
		return -1
	}
	return row*4 + col/2
}

// SquareToNotation converts square index to algebraic notation (e.g., 0 → "b8").
func SquareToNotation(sq int) string {
	row, col := SquareToRowCol(sq)
	file := string(rune('a' + col))
	rank := 8 - row
	return fmt.Sprintf("%s%d", file, rank)
}

// NotationToSquare converts algebraic notation to square index.
func NotationToSquare(notation string) (int, error) {
	if len(notation) != 2 {
		return -1, fmt.Errorf("invalid notation: %s", notation)
	}
	col := int(notation[0] - 'a')
	rank := int(notation[1] - '0')
	row := 8 - rank
	sq := RowColToSquare(row, col)
	if sq < 0 {
		return -1, fmt.Errorf("not a dark square: %s", notation)
	}
	return sq, nil
}

// --- Promotion rows ---

// WhitePromotionRow is the bitmask of squares where white men promote (row 0 = rank 8).
var WhitePromotionRow uint32 = 0xF // squares 0-3

// BlackPromotionRow is the bitmask of squares where black men promote (row 7 = rank 1).
var BlackPromotionRow uint32 = 0xF0000000 // squares 28-31

// PromotionRow returns the promotion row bitmask for the given color.
func PromotionRow(c Color) uint32 {
	if c == White {
		return WhitePromotionRow
	}
	return BlackPromotionRow
}

// --- FEN (PDN notation) ---

// ToFEN converts a board to PDN FEN string.
// Format: W:Wp1,p2,Kp3:Bp4,p5,Kp6
// W/B = side to move; positions are 1-based (1-32).
func (b Board) ToFEN() string {
	var sb strings.Builder
	sb.WriteString(b.Turn.String())

	sb.WriteString(":W")
	first := true
	for sq := 0; sq < 32; sq++ {
		mask := uint32(1) << sq
		if b.WhiteKing&mask != 0 {
			if !first {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("K%d", sq+1))
			first = false
		} else if b.White&mask != 0 {
			if !first {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("%d", sq+1))
			first = false
		}
	}

	sb.WriteString(":B")
	first = true
	for sq := 0; sq < 32; sq++ {
		mask := uint32(1) << sq
		if b.BlackKing&mask != 0 {
			if !first {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("K%d", sq+1))
			first = false
		} else if b.Black&mask != 0 {
			if !first {
				sb.WriteString(",")
			}
			sb.WriteString(fmt.Sprintf("%d", sq+1))
			first = false
		}
	}

	return sb.String()
}

// ParseFEN parses a PDN FEN string into a Board.
func ParseFEN(fen string) (Board, error) {
	var b Board

	parts := strings.Split(fen, ":")
	if len(parts) != 3 {
		return b, fmt.Errorf("invalid FEN: expected 3 parts separated by ':', got %d", len(parts))
	}

	// Turn
	switch strings.ToUpper(parts[0]) {
	case "W":
		b.Turn = White
	case "B":
		b.Turn = Black
	default:
		return b, fmt.Errorf("invalid FEN: turn must be W or B, got %s", parts[0])
	}

	// Parse white and black pieces
	for i := 1; i <= 2; i++ {
		part := parts[i]
		if len(part) < 1 {
			return b, fmt.Errorf("invalid FEN part: %s", part)
		}

		var isBlack bool
		switch strings.ToUpper(string(part[0])) {
		case "W":
			isBlack = false
		case "B":
			isBlack = true
		default:
			return b, fmt.Errorf("invalid FEN part prefix: %s", string(part[0]))
		}

		pieces := part[1:]
		if pieces == "" {
			continue
		}

		for _, p := range strings.Split(pieces, ",") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}

			isKing := false
			if strings.HasPrefix(strings.ToUpper(p), "K") {
				isKing = true
				p = p[1:]
			}

			num, err := strconv.Atoi(p)
			if err != nil {
				return b, fmt.Errorf("invalid square number: %s", p)
			}
			if num < 1 || num > 32 {
				return b, fmt.Errorf("square number out of range: %d", num)
			}

			sq := num - 1 // convert to 0-based
			mask := uint32(1) << sq

			if isBlack {
				if isKing {
					b.BlackKing |= mask
				} else {
					b.Black |= mask
				}
			} else {
				if isKing {
					b.WhiteKing |= mask
				} else {
					b.White |= mask
				}
			}
		}
	}

	return b, nil
}

// --- Zobrist Hashing ---

// zobristTable stores random values for each (square, piece-type) combination.
// piece-type: 0=white man, 1=black man, 2=white king, 3=black king
var zobristTable [32][4]uint64
var zobristTurn uint64

func init() {
	r := rand.New(rand.NewSource(0xDEADBEEF))
	for sq := 0; sq < 32; sq++ {
		for pt := 0; pt < 4; pt++ {
			zobristTable[sq][pt] = r.Uint64()
		}
	}
	zobristTurn = r.Uint64()
}

// Hash computes the Zobrist hash of the board.
func (b Board) Hash() uint64 {
	var h uint64
	for sq := 0; sq < 32; sq++ {
		mask := uint32(1) << sq
		if b.White&mask != 0 {
			h ^= zobristTable[sq][0]
		}
		if b.Black&mask != 0 {
			h ^= zobristTable[sq][1]
		}
		if b.WhiteKing&mask != 0 {
			h ^= zobristTable[sq][2]
		}
		if b.BlackKing&mask != 0 {
			h ^= zobristTable[sq][3]
		}
	}
	if b.Turn == Black {
		h ^= zobristTurn
	}
	return h
}

// NewBoard creates a standard starting position.
func NewBoard() Board {
	return Board{
		// Black occupies rows 0-2 (squares 0-11)
		Black: 0x00000FFF,
		// White occupies rows 5-7 (squares 20-31)
		White: 0xFFF00000,
		Turn:  White,
	}
}

// Clone returns a copy of the board.
func (b Board) Clone() Board {
	return b // structs are copied by value
}

// String returns a text representation of the board for debugging.
func (b Board) String() string {
	var sb strings.Builder
	sb.WriteString("  a b c d e f g h\n")
	for row := 0; row < 8; row++ {
		rank := 8 - row
		sb.WriteString(fmt.Sprintf("%d ", rank))
		for col := 0; col < 8; col++ {
			sq := RowColToSquare(row, col)
			if sq < 0 {
				sb.WriteString(". ")
				continue
			}
			mask := uint32(1) << sq
			switch {
			case b.WhiteKing&mask != 0:
				sb.WriteString("W ")
			case b.White&mask != 0:
				sb.WriteString("w ")
			case b.BlackKing&mask != 0:
				sb.WriteString("B ")
			case b.Black&mask != 0:
				sb.WriteString("b ")
			default:
				sb.WriteString("_ ")
			}
		}
		sb.WriteString(fmt.Sprintf("%d\n", rank))
	}
	sb.WriteString("  a b c d e f g h\n")
	sb.WriteString(fmt.Sprintf("Turn: %s\n", b.Turn))
	return sb.String()
}
