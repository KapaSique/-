import type { Tool } from '../lib/types';
import { Eraser, RotateCcw } from 'lucide-react';

interface ToolbarProps {
  selectedTool: Tool;
  onSelectTool: (tool: Tool) => void;
  turn: 'white' | 'black';
  onToggleTurn: () => void;
  onClear: () => void;
  onSolve: () => void;
  solving: boolean;
}

const tools: { tool: Tool; label: string }[] = [
  { tool: 'whitePiece', label: 'Белая' },
  { tool: 'blackPiece', label: 'Чёрная' },
  { tool: 'whiteKing', label: 'Белая дамка' },
  { tool: 'blackKing', label: 'Чёрная дамка' },
  { tool: 'eraser', label: 'Ластик' },
];

export default function Toolbar({
  selectedTool,
  onSelectTool,
  turn,
  onToggleTurn,
  onClear,
  onSolve,
  solving,
}: ToolbarProps) {
  return (
    <div className="flex flex-col gap-3">
      {/* Piece tools */}
      <div className="flex flex-wrap gap-2">
        {tools.map(({ tool, label }) => (
          <button
            key={tool}
            onClick={() => onSelectTool(tool)}
            className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all"
            style={{
              background: selectedTool === tool ? 'var(--accent)' : 'var(--muted-bg)',
              color: selectedTool === tool ? 'var(--accent-fg)' : 'var(--fg)',
              border: '1px solid var(--border)',
            }}
          >
            <ToolIcon tool={tool} active={selectedTool === tool} />
            {label}
          </button>
        ))}
      </div>

      {/* Turn toggle */}
      <div className="flex items-center gap-3">
        <button
          onClick={onToggleTurn}
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all"
          style={{
            background: 'var(--muted-bg)',
            color: 'var(--fg)',
            border: '1px solid var(--border)',
          }}
        >
          <div
            className="w-3 h-3 rounded-full"
            style={{
              background: turn === 'white' ? '#f5f5f5' : '#333',
              border: '1px solid #999',
            }}
          />
          Ход: {turn === 'white' ? 'белых' : 'чёрных'}
        </button>

        <button
          onClick={onClear}
          className="flex items-center gap-2 px-3 py-2 rounded-lg text-sm font-medium transition-all"
          style={{
            background: 'var(--muted-bg)',
            color: 'var(--fg)',
            border: '1px solid var(--border)',
          }}
        >
          <RotateCcw size={14} />
          Очистить
        </button>
      </div>

      {/* Solve button */}
      <button
        onClick={onSolve}
        disabled={solving}
        className="px-6 py-3 rounded-lg text-base font-semibold transition-all disabled:opacity-60"
        style={{
          background: 'var(--accent)',
          color: 'var(--accent-fg)',
          border: 'none',
          cursor: solving ? 'wait' : 'pointer',
        }}
      >
        {solving ? 'Решаю...' : 'Решить'}
      </button>
    </div>
  );
}

function ToolIcon({ tool, active }: { tool: Tool; active: boolean }) {
  if (tool === 'eraser') {
    return <Eraser size={16} />;
  }

  const isWhite = tool === 'whitePiece' || tool === 'whiteKing';
  const isKing = tool === 'whiteKing' || tool === 'blackKing';

  return (
    <svg width="20" height="20" viewBox="0 0 20 20">
      <circle
        cx="10"
        cy="10"
        r="8"
        fill={isWhite ? '#f5f5f5' : '#333'}
        stroke={active ? 'var(--accent-fg)' : '#999'}
        strokeWidth="1.5"
      />
      {isKing && (
        <text
          x="10"
          y="14"
          textAnchor="middle"
          fontSize="10"
          fill="#ffd700"
          fontWeight="bold"
        >
          ♛
        </text>
      )}
    </svg>
  );
}
