import { ChevronLeft, ChevronRight, Trophy, Minus, X } from 'lucide-react';
import type { SolveResponse, MoveJSON } from '../lib/types';

interface SolutionPanelProps {
  solution: SolveResponse | null;
  error: string | null;
  currentMoveIndex: number;
  onNavigate: (index: number) => void;
  onMoveClick: (move: MoveJSON) => void;
}

export default function SolutionPanel({
  solution,
  error,
  currentMoveIndex,
  onNavigate,
  onMoveClick,
}: SolutionPanelProps) {
  if (error) {
    return (
      <div
        className="rounded-lg p-4"
        style={{ background: 'var(--muted-bg)', border: '1px solid var(--destructive)' }}
      >
        <div className="flex items-center gap-2 text-sm font-medium" style={{ color: 'var(--destructive)' }}>
          <X size={16} />
          Ошибка
        </div>
        <p className="mt-2 text-sm" style={{ color: 'var(--muted)' }}>
          {error}
        </p>
      </div>
    );
  }

  if (!solution) {
    return (
      <div
        className="rounded-lg p-4"
        style={{ background: 'var(--muted-bg)', border: '1px solid var(--border)' }}
      >
        <p className="text-sm" style={{ color: 'var(--muted)' }}>
          Расставьте позицию и нажмите «Решить»
        </p>
      </div>
    );
  }

  const scoreText = getScoreText(solution.score);
  const scoreIcon = getScoreIcon(solution.score);

  return (
    <div
      className="rounded-lg p-4 space-y-3"
      style={{ background: 'var(--muted-bg)', border: '1px solid var(--border)' }}
    >
      {/* Result header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {scoreIcon}
          <span className="font-semibold text-sm">{scoreText}</span>
        </div>
        <span className="text-xs" style={{ color: 'var(--muted)' }}>
          Глубина: {solution.depth}
        </span>
      </div>

      {/* Stats */}
      <div className="flex gap-4 text-xs" style={{ color: 'var(--muted)' }}>
        <span>Узлов: {formatNumber(solution.nodes)}</span>
        <span>Время: {solution.time_ms}мс</span>
      </div>

      {/* Move list */}
      {solution.moves.length > 0 && (
        <div>
          <div className="text-xs font-medium mb-2" style={{ color: 'var(--muted)' }}>
            Главный вариант:
          </div>
          <div className="flex flex-wrap gap-1">
            {solution.moves.map((move, i) => (
              <button
                key={i}
                onClick={() => onMoveClick(move)}
                className="px-2 py-1 rounded text-xs font-mono transition-all"
                style={{
                  background: i === currentMoveIndex ? 'var(--accent)' : 'var(--card-bg)',
                  color: i === currentMoveIndex ? 'var(--accent-fg)' : 'var(--fg)',
                  border: '1px solid var(--border)',
                  cursor: 'pointer',
                }}
              >
                {i % 2 === 0 ? `${Math.floor(i / 2) + 1}. ` : ''}
                {move.notation}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Navigation */}
      {solution.moves.length > 0 && (
        <div className="flex items-center gap-2">
          <button
            onClick={() => onNavigate(Math.max(0, currentMoveIndex - 1))}
            disabled={currentMoveIndex <= 0}
            className="p-1.5 rounded disabled:opacity-30 transition-all"
            style={{ background: 'var(--card-bg)', border: '1px solid var(--border)' }}
          >
            <ChevronLeft size={16} />
          </button>
          <span className="text-xs" style={{ color: 'var(--muted)' }}>
            {currentMoveIndex + 1} / {solution.moves.length}
          </span>
          <button
            onClick={() => onNavigate(Math.min(solution.moves.length - 1, currentMoveIndex + 1))}
            disabled={currentMoveIndex >= solution.moves.length - 1}
            className="p-1.5 rounded disabled:opacity-30 transition-all"
            style={{ background: 'var(--card-bg)', border: '1px solid var(--border)' }}
          >
            <ChevronRight size={16} />
          </button>
        </div>
      )}
    </div>
  );
}

function getScoreText(score: number): string {
  if (score >= 99900) return 'Выигрыш найден!';
  if (score <= -99900) return 'Проигрыш';
  if (Math.abs(score) < 20) return 'Ничья';
  return `Оценка: ${score > 0 ? '+' : ''}${(score / 100).toFixed(1)}`;
}

function getScoreIcon(score: number) {
  if (score >= 99900) {
    return <Trophy size={18} style={{ color: 'var(--success)' }} />;
  }
  if (score <= -99900) {
    return <X size={18} style={{ color: 'var(--destructive)' }} />;
  }
  return <Minus size={18} style={{ color: 'var(--muted)' }} />;
}

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}
