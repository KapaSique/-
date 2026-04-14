package api

import (
	"checkers-solver/engine"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// --- Request/Response types ---

type SolveRequestJSON struct {
	FEN       string `json:"fen"`
	Turn      string `json:"turn"`
	Goal      string `json:"goal"`
	MaxDepth  int    `json:"max_depth"`
	TimeLimit int    `json:"time_limit"` // seconds
}

type MoveJSON struct {
	From     int    `json:"from"`
	To       int    `json:"to"`
	Captures []int  `json:"captures"`
	Notation string `json:"notation"`
	Promoted bool   `json:"promoted"`
}

type NodeJSON struct {
	Move     MoveJSON    `json:"move"`
	Score    int         `json:"score"`
	Children []*NodeJSON `json:"children,omitempty"`
	IsBest   bool        `json:"is_best"`
}

type SolveResponseJSON struct {
	Found  bool       `json:"found"`
	Score  int        `json:"score"`
	Depth  int        `json:"depth"`
	Moves  []MoveJSON `json:"moves"`
	Tree   *NodeJSON  `json:"tree,omitempty"`
	Nodes  int64      `json:"nodes"`
	TimeMs int64      `json:"time_ms"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// --- Handlers ---

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func HandleSolve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "method not allowed"})
		return
	}

	var req SolveRequestJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid JSON: " + err.Error()})
		return
	}

	// Parse FEN
	board, err := engine.ParseFEN(req.FEN)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "invalid FEN: " + err.Error()})
		return
	}

	// Override turn if specified
	if req.Turn == "black" || req.Turn == "B" {
		board.Turn = engine.Black
	} else if req.Turn == "white" || req.Turn == "W" {
		board.Turn = engine.White
	}

	// Parse goal
	goalType := engine.GoalWin
	switch req.Goal {
	case "draw":
		goalType = engine.GoalDraw
	case "mate_in":
		goalType = engine.GoalMateIn
	}

	// Defaults
	maxDepth := req.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 15
	}
	timeLimit := req.TimeLimit
	if timeLimit <= 0 {
		timeLimit = 5
	}

	solveReq := engine.SolveRequest{
		Board:     board,
		GoalType:  goalType,
		MaxDepth:  maxDepth,
		TimeLimit: time.Duration(timeLimit) * time.Second,
	}

	// Run solver with request context timeout
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(timeLimit+1)*time.Second)
	defer cancel()

	log.Printf("Solving: FEN=%s depth=%d time=%ds", req.FEN, maxDepth, timeLimit)

	result := engine.Solve(ctx, solveReq)

	log.Printf("Result: found=%v score=%d depth=%d nodes=%d time=%dms",
		result.Found, result.Score, result.Depth, result.NodesCount, result.TimeMs)

	// Convert to JSON response
	resp := SolveResponseJSON{
		Found:  result.Found,
		Score:  result.Score,
		Depth:  result.Depth,
		Moves:  convertMoves(result.Moves),
		Tree:   convertNode(result.Tree),
		Nodes:  result.NodesCount,
		TimeMs: result.TimeMs,
	}

	writeJSON(w, http.StatusOK, resp)
}

// --- Conversion helpers ---

func convertMoves(moves []engine.Move) []MoveJSON {
	result := make([]MoveJSON, len(moves))
	for i, m := range moves {
		result[i] = MoveJSON{
			From:     m.From,
			To:       m.To,
			Captures: m.Captures,
			Notation: m.Notation(),
			Promoted: m.Promoted,
		}
	}
	return result
}

func convertNode(n *engine.Node) *NodeJSON {
	if n == nil {
		return nil
	}
	node := &NodeJSON{
		Move: MoveJSON{
			From:     n.Move.From,
			To:       n.Move.To,
			Captures: n.Move.Captures,
			Notation: n.Move.Notation(),
			Promoted: n.Move.Promoted,
		},
		Score:  n.Score,
		IsBest: n.IsBest,
	}
	for _, child := range n.Children {
		node.Children = append(node.Children, convertNode(child))
	}
	return node
}

// --- CORS middleware ---

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Utility ---

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
