# Design: Checkers Solver Web Application

## Overview

Full-stack web application — Russian draughts (шашки) puzzle solver. User places pieces on a board, clicks "Solve", receives a forced solution with variation tree.

## Architecture

### Backend — Go Engine
- **Port**: `:8080`
- **Stack**: Go `net/http`, no frameworks
- **Core**: Bitboard representation (uint32 for 32 dark squares)
- **Search**: Negamax with alpha-beta pruning, iterative deepening, transposition table (Zobrist hashing), move ordering, quiescence search, null move pruning
- **API**: JSON REST with CORS for `localhost:5173`

### Frontend — React + TypeScript
- **Port**: `:5173` (Vite dev server)
- **Stack**: React + TypeScript, Vite, shadcn/ui, Tailwind CSS
- **Font**: TT Travels Next (fallback: Geist)
- **Theme**: Light/dark with toggle, CSS variables via shadcn/ui convention
- **Board**: SVG-based rendering with animations

## Key Components

### Backend
| File | Purpose |
|------|---------|
| `engine/board.go` | Bitboard representation, FEN parsing, Zobrist hashing |
| `engine/moves.go` | Move generation (captures, non-captures, king moves, multi-capture, Turkish strike) |
| `engine/eval.go` | Position evaluation with positional bonuses |
| `engine/search.go` | Negamax + alpha-beta, iterative deepening, TT, quiescence, null move |
| `engine/solver.go` | Solver entry point, goal types (WIN, DRAW, MATE_IN_N) |
| `api/handlers.go` | HTTP handlers, CORS, JSON serialization |

### Frontend
| Component | Purpose |
|-----------|---------|
| `Board.tsx` | SVG board rendering, piece placement, click handling, highlights |
| `Toolbar.tsx` | Piece placement tools (white/black piece/king, eraser), turn toggle |
| `SolutionPanel.tsx` | Solution display, notation, navigation, stats |
| `ThemeToggle.tsx` | Light/dark theme switcher |
| `PieceCounter.tsx` | Piece count display |

## API Contract

### POST /api/solve
Request: `{ fen, turn, goal, max_depth, time_limit }`
Response: `{ found, score, depth, moves, tree, nodes, time_ms }`

### GET /api/health
Returns: `{ status: "ok" }`

## Russian Draughts Rules (Critical)

1. Simple pieces move diagonally forward one square
2. Captures are mandatory and can be in any direction (including backward)
3. Multi-capture: piece removed after series completes
4. Kings move any distance diagonally, capture through any distance
5. If capture exists, non-capture moves are forbidden
6. Turkish strike: piece cannot capture the same piece twice in a series
7. Promotion: piece reaching king row mid-capture becomes king and continues as king

## Testing Strategy

Go unit tests for move generation:
- Mandatory capture
- Multi-capture series
- King promotion
- King moves
- Turkish strike
- Known puzzle positions

## Implementation Order

1. Correct move generation (most critical)
2. Basic negamax with alpha-beta
3. HTTP endpoint `/api/solve`
4. Frontend with board and API call
5. Optimizations (TT, move ordering, iterative deepening, quiescence, null move)
