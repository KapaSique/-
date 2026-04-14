package engine

import (
	"context"
	"time"
)

// GoalType defines what the solver is trying to find.
type GoalType int

const (
	GoalWin    GoalType = iota // find forced win
	GoalDraw                   // find forced draw
	GoalMateIn                 // find win in N moves
)

// SolveRequest contains parameters for the solver.
type SolveRequest struct {
	Board     Board
	GoalType  GoalType
	MaxDepth  int
	TimeLimit time.Duration
}

// SolveResult contains the solver output.
type SolveResult struct {
	Found      bool
	Score      int
	Depth      int
	Moves      []Move  // principal variation (forced solution)
	Tree       *Node   // variation tree
	NodesCount int64
	TimeMs     int64
}

// Node represents a node in the solution tree.
type Node struct {
	Move     Move
	Score    int
	Children []*Node
	IsBest   bool
}

// Solve runs the solver to find a forced solution.
func Solve(ctx context.Context, req SolveRequest) SolveResult {
	start := time.Now()

	cfg := DefaultConfig()
	cfg.MaxDepth = req.MaxDepth
	cfg.TimeLimit = req.TimeLimit

	searcher := NewSearcher(cfg)

	// Create a context with timeout
	searchCtx, cancel := context.WithTimeout(ctx, req.TimeLimit)
	defer cancel()

	result := searcher.SearchWithContext(searchCtx, req.Board, req.MaxDepth)

	elapsed := time.Since(start)

	// Build solution tree
	tree := buildSolutionTree(req.Board, result.PV)

	found := false
	switch req.GoalType {
	case GoalWin:
		found = result.Score >= WinScore-100
	case GoalDraw:
		found = result.Score >= -10 && result.Score <= 10
	case GoalMateIn:
		found = result.Score >= WinScore-100
	}

	return SolveResult{
		Found:      found,
		Score:      result.Score,
		Depth:      result.Depth,
		Moves:      result.PV,
		Tree:       tree,
		NodesCount: result.Nodes,
		TimeMs:     elapsed.Milliseconds(),
	}
}

// buildSolutionTree builds a tree from the principal variation.
func buildSolutionTree(b Board, pv []Move) *Node {
	if len(pv) == 0 {
		return nil
	}

	root := &Node{
		Move:   pv[0],
		IsBest: true,
	}

	current := root
	board := b
	for i, m := range pv {
		if i == 0 {
			board = ApplyMove(board, m)
			continue
		}
		child := &Node{
			Move:   m,
			IsBest: true,
		}
		current.Children = append(current.Children, child)

		// Add alternative moves as non-best children
		if i%2 == 1 { // Opponent's moves — show all alternatives
			altMoves := GenerateMoves(board)
			for _, alt := range altMoves {
				if alt.From == m.From && alt.To == m.To {
					continue
				}
				altNode := &Node{
					Move:   alt,
					IsBest: false,
				}
				current.Children = append(current.Children, altNode)
			}
		}

		board = ApplyMove(board, m)
		current = child
	}

	return root
}

// SolveWithFullTree performs a shallow solve that builds a complete variation tree.
func SolveWithFullTree(ctx context.Context, req SolveRequest) SolveResult {
	start := time.Now()

	cfg := DefaultConfig()
	cfg.MaxDepth = req.MaxDepth
	cfg.TimeLimit = req.TimeLimit

	searcher := NewSearcher(cfg)

	searchCtx, cancel := context.WithTimeout(ctx, req.TimeLimit)
	defer cancel()

	result := searcher.SearchWithContext(searchCtx, req.Board, req.MaxDepth)
	elapsed := time.Since(start)

	// Build a deeper tree by exploring alternatives
	tree := buildFullTree(searchCtx, req.Board, req.MaxDepth, searcher)

	found := false
	if req.GoalType == GoalWin {
		found = result.Score >= WinScore-100
	}

	return SolveResult{
		Found:      found,
		Score:      result.Score,
		Depth:      result.Depth,
		Moves:      result.PV,
		Tree:       tree,
		NodesCount: result.Nodes,
		TimeMs:     elapsed.Milliseconds(),
	}
}

// buildFullTree builds a tree by searching each move at the root.
func buildFullTree(ctx context.Context, b Board, maxDepth int, searcher *Searcher) *Node {
	moves := GenerateMoves(b)
	if len(moves) == 0 {
		return nil
	}

	// Find best move
	var bestIdx int
	bestScore := -Infinity
	children := make([]*Node, len(moves))

	for i, m := range moves {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		nb := ApplyMove(b, m)
		result := searcher.SearchWithContext(ctx, nb, maxDepth-1)
		score := -result.Score

		child := &Node{
			Move:  m,
			Score: score,
		}

		// Recursively build children for the best response
		if len(result.PV) > 0 && maxDepth > 2 {
			child.Children = buildPVChildren(nb, result.PV)
		}

		children[i] = child

		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	children[bestIdx].IsBest = true

	// Create a virtual root
	root := &Node{
		Move:     moves[bestIdx],
		Score:    bestScore,
		Children: children,
		IsBest:   true,
	}

	return root
}

func buildPVChildren(b Board, pv []Move) []*Node {
	if len(pv) == 0 {
		return nil
	}

	node := &Node{
		Move:   pv[0],
		IsBest: true,
	}

	if len(pv) > 1 {
		nb := ApplyMove(b, pv[0])
		node.Children = buildPVChildren(nb, pv[1:])
	}

	return []*Node{node}
}
