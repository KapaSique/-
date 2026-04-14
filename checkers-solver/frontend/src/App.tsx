import { useState, useCallback } from 'react';
import Board from './components/Board';
import Toolbar from './components/Toolbar';
import SolutionPanel from './components/SolutionPanel';
import ThemeToggle from './components/ThemeToggle';
import PieceCounter from './components/PieceCounter';
import { solvePosition } from './lib/api';
import type { Position, Tool, SolveResponse, MoveJSON } from './lib/types';
import { positionToFEN } from './lib/types';

export default function App() {
  const [position, setPosition] = useState<Position>({});
  const [selectedTool, setSelectedTool] = useState<Tool>('whitePiece');
  const [turn, setTurn] = useState<'white' | 'black'>('white');
  const [solving, setSolving] = useState(false);
  const [solution, setSolution] = useState<SolveResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [currentMoveIndex, setCurrentMoveIndex] = useState(0);
  const [highlightedSquares, setHighlightedSquares] = useState<number[]>([]);
  const [bestMoveSquares, setBestMoveSquares] = useState<number[]>([]);

  const handleSquareClick = useCallback(
    (sq: number) => {
      setPosition((prev) => {
        const next = { ...prev };
        switch (selectedTool) {
          case 'whitePiece':
            next[sq] = prev[sq] === 'white' ? null : 'white';
            break;
          case 'blackPiece':
            next[sq] = prev[sq] === 'black' ? null : 'black';
            break;
          case 'whiteKing':
            next[sq] = prev[sq] === 'whiteKing' ? null : 'whiteKing';
            break;
          case 'blackKing':
            next[sq] = prev[sq] === 'blackKing' ? null : 'blackKing';
            break;
          case 'eraser':
            next[sq] = null;
            break;
        }
        if (next[sq] === null) delete next[sq];
        return next;
      });
      setSolution(null);
      setError(null);
      setHighlightedSquares([]);
      setBestMoveSquares([]);
    },
    [selectedTool]
  );

  const handleClear = useCallback(() => {
    setPosition({});
    setSolution(null);
    setError(null);
    setHighlightedSquares([]);
    setBestMoveSquares([]);
    setCurrentMoveIndex(0);
  }, []);

  const handleSolve = useCallback(async () => {
    const hasPieces = Object.keys(position).length > 0;
    if (!hasPieces) {
      setError('Расставьте фигуры на доске');
      return;
    }

    setSolving(true);
    setError(null);
    setSolution(null);
    setHighlightedSquares([]);
    setBestMoveSquares([]);
    setCurrentMoveIndex(0);

    try {
      const fen = positionToFEN(position, turn);
      const result = await solvePosition({
        fen,
        turn: turn === 'white' ? 'W' : 'B',
        maxDepth: 15,
        timeLimit: 5,
      });

      setSolution(result);

      if (result.moves.length > 0) {
        const firstMove = result.moves[0];
        setBestMoveSquares([firstMove.from, firstMove.to]);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Неизвестная ошибка');
    } finally {
      setSolving(false);
    }
  }, [position, turn]);

  const handleNavigate = useCallback(
    (index: number) => {
      if (!solution) return;
      setCurrentMoveIndex(index);
      const move = solution.moves[index];
      if (move) {
        setHighlightedSquares([move.from, move.to]);
        setBestMoveSquares(index === 0 ? [move.from, move.to] : []);
      }
    },
    [solution]
  );

  const handleMoveClick = useCallback(
    (move: MoveJSON) => {
      if (!solution) return;
      const index = solution.moves.findIndex(
        (m) => m.from === move.from && m.to === move.to
      );
      if (index >= 0) {
        handleNavigate(index);
      }
    },
    [solution, handleNavigate]
  );

  return (
    <div
      className="min-h-screen"
      style={{ background: 'var(--bg)', color: 'var(--fg)' }}
    >
      {/* Header */}
      <header
        className="flex items-center justify-between px-6 py-4"
        style={{ borderBottom: '1px solid var(--border)' }}
      >
        <h1 className="text-xl font-bold tracking-tight">
          Солвер шашечных задач
        </h1>
        <ThemeToggle />
      </header>

      {/* Main content */}
      <main className="flex flex-col lg:flex-row gap-6 p-6 max-w-6xl mx-auto">
        {/* Left: Board */}
        <div className="flex flex-col items-center gap-4">
          <Board
            position={position}
            onSquareClick={handleSquareClick}
            selectedTool={selectedTool}
            highlightedSquares={highlightedSquares}
            bestMoveSquares={bestMoveSquares}
          />
          <PieceCounter position={position} />
        </div>

        {/* Right: Controls + Solution */}
        <div className="flex flex-col gap-4 lg:w-80">
          <Toolbar
            selectedTool={selectedTool}
            onSelectTool={setSelectedTool}
            turn={turn}
            onToggleTurn={() =>
              setTurn((t) => (t === 'white' ? 'black' : 'white'))
            }
            onClear={handleClear}
            onSolve={handleSolve}
            solving={solving}
          />

          <SolutionPanel
            solution={solution}
            error={error}
            currentMoveIndex={currentMoveIndex}
            onNavigate={handleNavigate}
            onMoveClick={handleMoveClick}
          />
        </div>
      </main>
    </div>
  );
}
