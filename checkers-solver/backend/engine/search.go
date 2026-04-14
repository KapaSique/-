package engine

import (
	"context"
	"math/bits"
	"time"
)

// --- Transposition Table ---

type TTFlag int

const (
	TTExact      TTFlag = iota
	TTLowerBound        // score >= beta (beta cutoff)
	TTUpperBound        // score <= alpha (failed low)
)

type TTEntry struct {
	Hash  uint64
	Depth int
	Score int
	Flag  TTFlag
	Move  Move
}

type TranspositionTable struct {
	entries []TTEntry
	size    int
	mask    int
}

func NewTranspositionTable(size int) *TranspositionTable {
	// Round to power of 2
	actual := 1
	for actual < size {
		actual <<= 1
	}
	return &TranspositionTable{
		entries: make([]TTEntry, actual),
		size:    actual,
		mask:    actual - 1,
	}
}

func (tt *TranspositionTable) Probe(hash uint64) (TTEntry, bool) {
	idx := int(hash) & tt.mask
	entry := tt.entries[idx]
	if entry.Hash == hash {
		return entry, true
	}
	return TTEntry{}, false
}

func (tt *TranspositionTable) Store(hash uint64, depth, score int, flag TTFlag, move Move) {
	idx := int(hash) & tt.mask
	// Replace if deeper or same position
	existing := tt.entries[idx]
	if existing.Hash == 0 || existing.Depth <= depth || existing.Hash == hash {
		tt.entries[idx] = TTEntry{
			Hash:  hash,
			Depth: depth,
			Score: score,
			Flag:  flag,
			Move:  move,
		}
	}
}

func (tt *TranspositionTable) Clear() {
	for i := range tt.entries {
		tt.entries[i] = TTEntry{}
	}
}

// --- Killer Moves ---

type KillerTable struct {
	killers [64][2]Move // [ply][slot]
}

func (kt *KillerTable) Store(ply int, m Move) {
	if ply >= 64 {
		return
	}
	if kt.killers[ply][0].From != m.From || kt.killers[ply][0].To != m.To {
		kt.killers[ply][1] = kt.killers[ply][0]
		kt.killers[ply][0] = m
	}
}

func (kt *KillerTable) IsKiller(ply int, m Move) bool {
	if ply >= 64 {
		return false
	}
	return (kt.killers[ply][0].From == m.From && kt.killers[ply][0].To == m.To) ||
		(kt.killers[ply][1].From == m.From && kt.killers[ply][1].To == m.To)
}

// --- History Heuristic ---

type HistoryTable struct {
	table [2][32][32]int // [color][from][to]
}

func (ht *HistoryTable) Update(color Color, m Move, depth int) {
	c := 0
	if color == Black {
		c = 1
	}
	ht.table[c][m.From][m.To] += depth * depth
}

func (ht *HistoryTable) Score(color Color, m Move) int {
	c := 0
	if color == Black {
		c = 1
	}
	return ht.table[c][m.From][m.To]
}

func (ht *HistoryTable) Clear() {
	for c := 0; c < 2; c++ {
		for i := 0; i < 32; i++ {
			for j := 0; j < 32; j++ {
				ht.table[c][i][j] = 0
			}
		}
	}
}

// --- Engine Config ---

type EngineConfig struct {
	MaxDepth     int
	TimeLimit    time.Duration
	TTSize       int
	UseNullMove  bool
	UseKillers   bool
	UseHistory   bool
	QuiesceDepth int
}

func DefaultConfig() EngineConfig {
	return EngineConfig{
		MaxDepth:     20,
		TimeLimit:    5 * time.Second,
		TTSize:       1 << 22,
		UseNullMove:  true,
		UseKillers:   true,
		UseHistory:   true,
		QuiesceDepth: 10,
	}
}

// --- Search Result ---

type SearchResult struct {
	Score    int
	BestMove Move
	PV       []Move // principal variation
	Nodes    int64
	Depth    int
}

// --- Searcher ---

type Searcher struct {
	Config  EngineConfig
	TT      *TranspositionTable
	Killers KillerTable
	History HistoryTable
	Nodes   int64
	ctx     context.Context
}

func NewSearcher(cfg EngineConfig) *Searcher {
	return &Searcher{
		Config: cfg,
		TT:     NewTranspositionTable(cfg.TTSize),
	}
}

// Search performs iterative deepening search.
func (s *Searcher) Search(b Board, maxDepth int, timeLimit time.Duration) SearchResult {
	ctx, cancel := context.WithTimeout(context.Background(), timeLimit)
	defer cancel()
	return s.SearchWithContext(ctx, b, maxDepth)
}

// SearchWithContext performs iterative deepening search with a context for cancellation.
func (s *Searcher) SearchWithContext(ctx context.Context, b Board, maxDepth int) SearchResult {
	s.ctx = ctx
	s.Nodes = 0
	s.History.Clear()

	var bestResult SearchResult

	// Iterative deepening
	for depth := 1; depth <= maxDepth; depth++ {
		result := s.searchRoot(b, depth)

		// Check if search was cancelled
		select {
		case <-ctx.Done():
			if bestResult.BestMove.From != 0 || bestResult.BestMove.To != 0 || len(bestResult.PV) > 0 {
				return bestResult
			}
			return result
		default:
		}

		result.Depth = depth
		result.Nodes = s.Nodes
		bestResult = result

		// If we found a winning score, no need to search deeper
		if result.Score >= WinScore-100 || result.Score <= -(WinScore-100) {
			break
		}
	}

	return bestResult
}

func (s *Searcher) searchRoot(b Board, depth int) SearchResult {
	moves := GenerateMoves(b)
	if len(moves) == 0 {
		return SearchResult{Score: -WinScore}
	}

	// Order moves
	s.orderMoves(b, moves, 0)

	var bestMove Move
	var bestPV []Move
	alpha := -Infinity
	beta := Infinity

	for _, m := range moves {
		nb := ApplyMove(b, m)
		s.Nodes++

		var childPV []Move
		score := -s.negamax(nb, depth-1, -beta, -alpha, 1, &childPV)

		select {
		case <-s.ctx.Done():
			return SearchResult{Score: alpha, BestMove: bestMove, PV: bestPV}
		default:
		}

		if score > alpha {
			alpha = score
			bestMove = m
			bestPV = append([]Move{m}, childPV...)
		}
	}

	return SearchResult{
		Score:    alpha,
		BestMove: bestMove,
		PV:       bestPV,
	}
}

func (s *Searcher) negamax(b Board, depth, alpha, beta, ply int, pv *[]Move) int {
	select {
	case <-s.ctx.Done():
		return 0
	default:
	}

	alphaOrig := alpha

	// Transposition table lookup
	hash := b.Hash()
	if entry, found := s.TT.Probe(hash); found && entry.Depth >= depth {
		switch entry.Flag {
		case TTExact:
			*pv = []Move{entry.Move}
			return entry.Score
		case TTLowerBound:
			if entry.Score > alpha {
				alpha = entry.Score
			}
		case TTUpperBound:
			if entry.Score < beta {
				beta = entry.Score
			}
		}
		if alpha >= beta {
			*pv = []Move{entry.Move}
			return entry.Score
		}
	}

	// Terminal check
	moves := GenerateMoves(b)
	if len(moves) == 0 {
		return -WinScore + ply // losing (prefer longer losses)
	}

	// Leaf node — quiescence or static eval
	if depth <= 0 {
		return s.quiescence(b, alpha, beta, s.Config.QuiesceDepth, ply)
	}

	// Null move pruning
	if s.Config.UseNullMove && depth >= 3 && !inCheck(b) && hasSufficientMaterial(b) {
		nullBoard := b.Clone()
		nullBoard.Turn = b.Turn.Opponent()
		var nullPV []Move
		score := -s.negamax(nullBoard, depth-3, -beta, -beta+1, ply+1, &nullPV)
		if score >= beta {
			return beta
		}
	}

	// Order moves
	s.orderMoves(b, moves, ply)

	var bestMove Move
	var bestPV []Move
	bestScore := -Infinity

	for _, m := range moves {
		nb := ApplyMove(b, m)
		s.Nodes++

		var childPV []Move
		score := -s.negamax(nb, depth-1, -beta, -alpha, ply+1, &childPV)

		if score > bestScore {
			bestScore = score
			bestMove = m
			bestPV = append([]Move{m}, childPV...)
		}

		if score > alpha {
			alpha = score
		}

		if alpha >= beta {
			// Beta cutoff
			if !m.IsCapture() {
				if s.Config.UseKillers {
					s.Killers.Store(ply, m)
				}
				if s.Config.UseHistory {
					s.History.Update(b.Turn, m, depth)
				}
			}
			break
		}
	}

	// Store in TT
	var flag TTFlag
	if bestScore <= alphaOrig {
		flag = TTUpperBound
	} else if bestScore >= beta {
		flag = TTLowerBound
	} else {
		flag = TTExact
	}
	s.TT.Store(hash, depth, bestScore, flag, bestMove)

	*pv = bestPV
	return bestScore
}

// quiescence searches only captures to avoid horizon effect.
func (s *Searcher) quiescence(b Board, alpha, beta, depth, ply int) int {
	select {
	case <-s.ctx.Done():
		return 0
	default:
	}

	standPat := Evaluate(b, b.Turn)

	if standPat >= beta {
		return beta
	}
	if standPat > alpha {
		alpha = standPat
	}

	if depth <= 0 {
		return standPat
	}

	captures := GenerateCaptures(b)
	if len(captures) == 0 {
		return standPat
	}

	for _, m := range captures {
		nb := ApplyMove(b, m)
		s.Nodes++

		score := -s.quiescence(nb, -beta, -alpha, depth-1, ply+1)

		if score >= beta {
			return beta
		}
		if score > alpha {
			alpha = score
		}
	}

	return alpha
}

// --- Move ordering ---

func (s *Searcher) orderMoves(b Board, moves []Move, ply int) {
	scores := make([]int, len(moves))
	hash := b.Hash()

	for i, m := range moves {
		score := 0

		// TT move gets highest priority
		if entry, found := s.TT.Probe(hash); found {
			if entry.Move.From == m.From && entry.Move.To == m.To {
				score += 10000000
			}
		}

		// Captures first
		if m.IsCapture() {
			score += 1000000 + len(m.Captures)*100000
		}

		// Killer moves
		if s.Config.UseKillers && s.Killers.IsKiller(ply, m) {
			score += 500000
		}

		// History heuristic
		if s.Config.UseHistory {
			score += s.History.Score(b.Turn, m)
		}

		// Promotion bonus
		if m.Promoted {
			score += 200000
		}

		scores[i] = score
	}

	// Sort by score descending (selection sort for small arrays)
	for i := 0; i < len(moves)-1; i++ {
		best := i
		for j := i + 1; j < len(moves); j++ {
			if scores[j] > scores[best] {
				best = j
			}
		}
		if best != i {
			moves[i], moves[best] = moves[best], moves[i]
			scores[i], scores[best] = scores[best], scores[i]
		}
	}
}

// --- Helper functions ---

// inCheck is a simplified check — in draughts there's no "check",
// but we use it to determine if captures are available (mandatory capture).
func inCheck(b Board) bool {
	return len(GenerateCaptures(b)) > 0
}

// hasSufficientMaterial returns true if the side to move has more than just a king.
func hasSufficientMaterial(b Board) bool {
	pieces := b.PiecesOf(b.Turn)
	return bits.OnesCount32(pieces) > 1
}
