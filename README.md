# Russian Checkers Solver

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=for-the-badge&logo=react&logoColor=black" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-6.0-3178C6?style=for-the-badge&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Vite-8.0-646CFF?style=for-the-badge&logo=vite&logoColor=white" alt="Vite" />
  <img src="https://img.shields.io/badge/Tailwind-4.2-06B6D4?style=for-the-badge&logo=tailwindcss&logoColor=white" alt="Tailwind" />
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Bitboard_Engine-FF6B35?style=flat-square" alt="Bitboard Engine" />
  <img src="https://img.shields.io/badge/Alpha%E2%80%91Beta_Search-4CAF50?style=flat-square" alt="Alpha-Beta" />
  <img src="https://img.shields.io/badge/Zero_External_Deps-9C27B0?style=flat-square" alt="Zero Deps" />
  <img src="https://img.shields.io/badge/Russian_Draughts_Rules-2196F3?style=flat-square" alt="Russian Draughts" />
</p>

<p align="center">
  <strong>Full-stack web application for solving Russian draughts (шашки) puzzles</strong>
</p>

<p align="center">
  Place pieces on the board, click <em>Solve</em>, and get a forced solution with a variation tree — like a chess tactics trainer, but for Russian draughts.
</p>

---

## Screenshots

<p align="center">
  <img src="https://placehold.co/800x450/1a1a2e/e0e0e0?text=Russian+Checkers+Solver+%F0%9F%8F%81&font=roboto" alt="App Screenshot" />
</p>

## Features

- **Bitboard engine** — entire position fits in 4×`uint32`, enabling lightning-fast move generation via bitwise operations
- **Alpha-beta search** — negamax with iterative deepening, transposition table, killer moves, history heuristic, null move pruning, and quiescence search
- **Full Russian draughts rules** — mandatory captures, multi-capture chains, flying kings, Turkish strike, mid-chain promotion
- **Zero external dependencies** (backend) — pure Go standard library, portable and easy to build
- **Modern React UI** — SVG board, light/dark theme, solution tree navigation, Russian draughts notation
- **Real-time solving** — Vite dev server proxies API calls to the Go backend seamlessly

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Browser                               │
│  ┌───────────┐  ┌──────────────┐  ┌───────────────────┐     │
│  │  SVG Board │  │   Toolbar    │  │ Solution Panel    │     │
│  │  (click)   │  │  (tools)     │  │ (moves + nav)     │     │
│  └─────┬─────┘  └──────┬───────┘  └─────────┬─────────┘     │
│        └───────────────┼────────────────────┘                │
│                    Position → FEN                            │
└────────────────────────┬────────────────────────────────────┘
                         │ POST /api/solve
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     Go Backend (:8080)                       │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌────────────┐  │
│  │  Parse   │→ │  Search  │→ │  Negamax │→ │  Build     │  │
│  │  FEN     │  │  Engine  │  │  + α-β   │  │  Solution  │  │
│  └──────────┘  └──────────┘  └──────────┘  └────────────┘  │
│                                                             │
│  Bitboards:  white_men | black_men | white_kings | black_kings
│  Search:     TT + Killer + History + Null Move + Quiescence │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### Backend

```bash
cd checkers-solver/backend
go run main.go
# Server starts on http://localhost:8080
```

### Frontend

```bash
cd checkers-solver/frontend
npm install
npm run dev
# App available at http://localhost:5173
```

The Vite dev server automatically proxies `/api` requests to the Go backend.

### Build for Production

```bash
cd checkers-solver/frontend
npm run build
# Output in dist/
```

## Project Structure

```
checkers-solver/
├── backend/
│   ├── main.go                  # HTTP server entry point
│   ├── api/
│   │   └── handlers.go          # /api/solve, /api/health, CORS
│   └── engine/
│       ├── board.go             # Bitboard representation, FEN, Zobrist hashing
│       ├── moves.go             # Move generation (all Russian draughts rules)
│       ├── moves_test.go        # 12 unit tests for move generation
│       ├── eval.go              # Position evaluation function
│       ├── search.go            # Negamax + alpha-beta, TT, quiescence
│       └── solver.go            # Solver entry point, goal types, tree building
└── frontend/
    ├── src/
    │   ├── App.tsx              # Root component (state management)
    │   ├── components/
    │   │   ├── Board.tsx        # SVG board rendering
    │   │   ├── Toolbar.tsx      # Piece placement tools
    │   │   ├── SolutionPanel.tsx# Solution display + navigation
    │   │   ├── ThemeToggle.tsx  # Light/dark theme switcher
    │   │   └── PieceCounter.tsx # Piece count display
    │   └── lib/
    │       ├── types.ts         # TypeScript types, FEN conversion
    │       └── api.ts           # API client
    └── ...
```

## Russian Draughts Rules Implemented

| Rule | Description |
|------|-------------|
| **Mandatory captures** | If a capture exists, quiet moves are forbidden |
| **Multi-capture chains** | A piece can capture multiple times in one move |
| **Flying kings** | Kings move any distance diagonally |
| **Turkish strike** | A piece cannot capture the same enemy twice in one sequence |
| **Mid-chain promotion** | A man reaching the promotion row becomes a king mid-chain |
| **Captured pieces** | Stay on board until the entire capture sequence ends |

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.26 (stdlib only, zero external deps) |
| **Frontend** | React 19 + TypeScript + Vite 8 |
| **Styling** | Tailwind CSS 4 with CSS variable theming |
| **Icons** | lucide-react |
| **Board** | SVG with precise coordinate math |

## License

MIT
