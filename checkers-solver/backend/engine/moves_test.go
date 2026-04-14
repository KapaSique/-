package engine

import (
	"testing"
)

// Helper: set a single piece on the board
func boardWith(pieces map[int]byte, turn Color) Board {
	var b Board
	b.Turn = turn
	for sq, p := range pieces {
		mask := uint32(1) << sq
		switch p {
		case 'w':
			b.White |= mask
		case 'b':
			b.Black |= mask
		case 'W':
			b.WhiteKing |= mask
		case 'B':
			b.BlackKing |= mask
		}
	}
	return b
}

func TestSquareConversion(t *testing.T) {
	// sq 0 should be b8
	if n := SquareToNotation(0); n != "b8" {
		t.Errorf("sq 0: expected b8, got %s", n)
	}
	// sq 4 should be a7
	if n := SquareToNotation(4); n != "a7" {
		t.Errorf("sq 4: expected a7, got %s", n)
	}
	// sq 31 should be g1 (h1 is a light square)
	if n := SquareToNotation(31); n != "g1" {
		t.Errorf("sq 31: expected g1, got %s", n)
	}
	// sq 28 should be a1
	if n := SquareToNotation(28); n != "a1" {
		t.Errorf("sq 28: expected a1, got %s", n)
	}
	// sq 27 should be h2
	if n := SquareToNotation(27); n != "h2" {
		t.Errorf("sq 27: expected h2, got %s", n)
	}

	// Round-trip
	for sq := 0; sq < 32; sq++ {
		notation := SquareToNotation(sq)
		back, err := NotationToSquare(notation)
		if err != nil {
			t.Errorf("failed to convert notation %s back: %v", notation, err)
		}
		if back != sq {
			t.Errorf("round trip failed: sq=%d, notation=%s, back=%d", sq, notation, back)
		}
	}
}

func TestSimpleManMoves(t *testing.T) {
	// White man on e3 (sq 22), should move to d4 (sq 17) or f4 (sq 18)
	b := boardWith(map[int]byte{22: 'w'}, White)
	moves := GenerateMoves(b)
	if len(moves) != 2 {
		t.Fatalf("expected 2 moves for white man on e3, got %d", len(moves))
	}

	destinations := map[int]bool{}
	for _, m := range moves {
		destinations[m.To] = true
	}
	if !destinations[17] || !destinations[18] {
		t.Errorf("expected destinations 17 and 18, got %v", destinations)
	}
}

func TestBlackManMoves(t *testing.T) {
	// Black man on d6 (sq 9), should move to c5 (sq 12) or e5 (sq 14)
	// Wait — let me verify: sq 9 = row 2, col 3 = d6
	// Forward for black = SW (row+1) and SE (row+1)
	// SW: row=3, col=2 = c5 -> RowColToSquare(3,2) = 3*4+2/2=13
	// SE: row=3, col=4 = e5 -> RowColToSquare(3,4) = 3*4+4/2=14
	b := boardWith(map[int]byte{9: 'b'}, Black)
	moves := GenerateMoves(b)
	if len(moves) != 2 {
		t.Fatalf("expected 2 moves for black man on d6 (sq 9), got %d", len(moves))
	}

	destinations := map[int]bool{}
	for _, m := range moves {
		destinations[m.To] = true
	}
	if !destinations[13] || !destinations[14] {
		t.Errorf("expected destinations 13 and 14, got %v", destinations)
	}
}

func TestMandatoryCapture(t *testing.T) {
	// White man on e3 (sq 22), black man on d4 (sq 17)
	// White must capture: e3:c5 (sq 22 -> sq 13, capturing sq 17)
	b := boardWith(map[int]byte{22: 'w', 17: 'b'}, White)
	moves := GenerateMoves(b)

	// All moves should be captures (mandatory capture rule)
	for _, m := range moves {
		if !m.IsCapture() {
			t.Errorf("expected only capture moves due to mandatory capture, got quiet move %s", m.Notation())
		}
	}

	if len(moves) == 0 {
		t.Fatal("expected at least one capture move")
	}

	// Should capture d4 and land on c5
	found := false
	for _, m := range moves {
		if m.From == 22 && m.To == 13 {
			found = true
			if len(m.Captures) != 1 || m.Captures[0] != 17 {
				t.Errorf("expected capture of sq 17, got %v", m.Captures)
			}
		}
	}
	if !found {
		t.Error("expected capture e3:c5 (22->13)")
	}
}

func TestBackwardCapture(t *testing.T) {
	// White man on c5 (sq 13), black man on d4 (sq 17)
	// Wait — d4 is below c5 for white. White man can capture backwards in Russian draughts.
	// c5 (sq 13) capturing d4 (sq 17) and landing on e3 (sq 22)
	// Direction: SW from sq 13 -> neighbor is sq 17 (d4), then sq 22 (e3)
	b := boardWith(map[int]byte{13: 'w', 17: 'b'}, White)
	moves := GenerateMoves(b)

	captureFound := false
	for _, m := range moves {
		if m.IsCapture() && m.From == 13 {
			captureFound = true
		}
	}
	if !captureFound {
		t.Error("white man should be able to capture backwards")
	}
}

func TestMultipleCapture(t *testing.T) {
	// White man on a1 (sq 28), black men on b2 (sq 24) and d4 (sq 17)
	// Wait, let me set up a proper multi-capture:
	// White man on g1 (sq 31), black on f2 (sq 26), black on d4 (sq 17)
	// g1 -> captures f2 -> lands on e3 (sq 22) -> captures d4 (sq 17) -> lands on c5 (sq 13)
	// Check: sq 31 = row 7, col 6 = g1
	// NW neighbor of 31: row 6, col 5 = f2 = sq 26... wait
	// row 6, col 5 -> RowColToSquare(6,5) = 6*4+5/2 = 24+2 = 26. Yes, f2 = sq 26
	// NW neighbor of 26: row 5, col 4 = e3 = RowColToSquare(5,4) = 5*4+4/2=22. Yes.
	// NW neighbor of 22: row 4, col 3 = d4 = RowColToSquare(4,3) = 4*4+3/2=17. Yes.
	// NW neighbor of 17: row 3, col 2 = c5 = RowColToSquare(3,2) = 3*4+2/2=13. Yes.
	b := boardWith(map[int]byte{31: 'w', 26: 'b', 17: 'b'}, White)
	moves := GenerateMoves(b)

	// Should find a multi-capture: 31 -> 22 -> 13 capturing 26 and 17
	found := false
	for _, m := range moves {
		if m.From == 31 && m.To == 13 && len(m.Captures) == 2 {
			found = true
		}
	}
	if !found {
		t.Errorf("expected multi-capture g1:e3:c5 (31->22->13), got moves: ")
		for _, m := range moves {
			t.Logf("  %s (from=%d to=%d captures=%v)", m.Notation(), m.From, m.To, m.Captures)
		}
	}
}

func TestKingPromotion(t *testing.T) {
	// White man on c7 (sq 5), one square from promotion
	// c7 -> b8 (sq 0) or d8 (sq 1) — both are promotion squares
	b := boardWith(map[int]byte{5: 'w'}, White)
	moves := GenerateMoves(b)

	promoted := 0
	for _, m := range moves {
		if m.Promoted {
			promoted++
		}
	}
	if promoted == 0 {
		t.Error("expected at least one promotion move")
	}
}

func TestKingMoves(t *testing.T) {
	// White king on e5 (sq 14) — should be able to move in all 4 diagonal directions
	b := boardWith(map[int]byte{14: 'W'}, White)
	moves := GenerateMoves(b)

	if len(moves) < 4 {
		t.Errorf("king on e5 should have many moves, got %d", len(moves))
	}

	// King should be able to reach far squares
	destinations := map[int]bool{}
	for _, m := range moves {
		destinations[m.To] = true
	}

	// NW diagonal from e5: d6(9), c7(5), b8(0) — 3 squares
	// NE diagonal from e5: f6(11), g7(7), h8(3) — 3 squares
	// SW diagonal from e5: d4(17), c3(21), b2(24), a1(28) — 4 squares
	// SE diagonal from e5: f4(19), g3(23), h2(27) — 3 squares
	// Total: 13 moves
	if len(moves) != 13 {
		t.Errorf("king on e5 (empty board) should have 13 moves, got %d", len(moves))
		for _, m := range moves {
			t.Logf("  %s (from=%d to=%d)", m.Notation(), m.From, m.To)
		}
	}
}

func TestKingCapture(t *testing.T) {
	// White king on a1 (sq 28), black man on c3 (sq 21)
	// King should be able to fly over and capture: a1 -> captures c3 -> lands on d4/e5/f6/g7/h8
	b := boardWith(map[int]byte{28: 'W', 21: 'b'}, White)
	moves := GenerateMoves(b)

	// All should be captures
	for _, m := range moves {
		if !m.IsCapture() {
			t.Errorf("expected only captures, got quiet move %s", m.Notation())
		}
	}

	if len(moves) == 0 {
		t.Fatal("expected king capture moves")
	}

	// King captures c3 (sq 21) and can land on any empty square beyond:
	// d4(17), e5(14), f6(11)... wait let me check the diagonal
	// a1(28) NE -> b2(24) -> c3(21) is the opponent piece -> d4(17), e5(14), f6(10)?
	// Let me trace: sq 28 = row 7, col 0
	// NE: row-1, col+1 = (6,1) = sq 24 (b2). empty? no it's the path before opponent
	// Wait, b2 is empty, c3(21) is opponent, so king flies: 28 -> past 24(empty) wait...
	// Actually diagRays[28] direction NE gives: 24, 21, 17, 14, 10, 7, 3
	// But 24 is between 28 and 21, and there's nothing at 24, so we continue.
	// At 21 we find the opponent. Beyond 21: 17, 14, 10, 7, 3 are landing options.
	// Wait, sq 10 = row 2, col 5 = f6. sq 7 = row 1, col 6 = g7. sq 3 = row 0, col 7 = h8.
	// So 5 possible landing squares: 17, 14, 10, 7, 3
	if len(moves) != 5 {
		t.Errorf("expected 5 king capture landing squares, got %d", len(moves))
		for _, m := range moves {
			t.Logf("  %s (from=%d to=%d captures=%v)", m.Notation(), m.From, m.To, m.Captures)
		}
	}
}

func TestTurkishStrike(t *testing.T) {
	// Turkish strike: during a multi-capture, a piece cannot jump over the same captured piece twice.
	// Set up: White king on a1 (28), black men at c3 (21) and e3 (22)
	// The king should NOT be able to capture c3, go to some square, turn around and capture c3 again.
	// This is inherently handled by our `captured` bitmask tracking.
	b := boardWith(map[int]byte{28: 'W', 21: 'b', 22: 'b'}, White)
	moves := GenerateMoves(b)

	for _, m := range moves {
		// Check no duplicate captures
		seen := map[int]bool{}
		for _, c := range m.Captures {
			if seen[c] {
				t.Errorf("Turkish strike violation: piece captured twice at sq %d in move %s", c, m.Notation())
			}
			seen[c] = true
		}
	}
}

func TestFENRoundTrip(t *testing.T) {
	b := NewBoard()
	fen := b.ToFEN()
	parsed, err := ParseFEN(fen)
	if err != nil {
		t.Fatalf("failed to parse FEN: %v", err)
	}

	if parsed.White != b.White || parsed.Black != b.Black ||
		parsed.WhiteKing != b.WhiteKing || parsed.BlackKing != b.BlackKing ||
		parsed.Turn != b.Turn {
		t.Errorf("FEN round trip failed.\nOriginal: %+v\nParsed:   %+v\nFEN: %s", b, parsed, fen)
	}
}

func TestNoMovesIsLoss(t *testing.T) {
	// A player with no pieces loses
	b := boardWith(map[int]byte{14: 'w'}, Black)
	moves := GenerateMoves(b)
	if len(moves) != 0 {
		t.Errorf("black has no pieces, should have 0 moves, got %d", len(moves))
	}
}

func TestInitialPosition(t *testing.T) {
	b := NewBoard()
	moves := GenerateMoves(b)
	// Standard starting position should have 7 moves for white
	if len(moves) != 7 {
		t.Errorf("initial position should have 7 moves for white, got %d", len(moves))
		for _, m := range moves {
			t.Logf("  %s", m.Notation())
		}
	}
}

func TestZobristConsistency(t *testing.T) {
	b1 := NewBoard()
	b2 := NewBoard()
	if b1.Hash() != b2.Hash() {
		t.Error("identical boards should have identical hashes")
	}

	// Different boards should (very likely) have different hashes
	b3 := boardWith(map[int]byte{14: 'w', 17: 'b'}, White)
	if b1.Hash() == b3.Hash() {
		t.Error("different boards should (very likely) have different hashes")
	}
}
