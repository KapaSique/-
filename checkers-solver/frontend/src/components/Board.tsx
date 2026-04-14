import { type Position, type Tool, squareToRowCol, rowColToSquare } from '../lib/types';

interface BoardProps {
  position: Position;
  onSquareClick: (sq: number) => void;
  selectedTool: Tool;
  highlightedSquares?: number[];
  bestMoveSquares?: number[];
}

const BOARD_SIZE = 480;
const CELL_SIZE = BOARD_SIZE / 8;
const MARGIN = 24;

export default function Board({
  position,
  onSquareClick,
  highlightedSquares = [],
  bestMoveSquares = [],
}: BoardProps) {
  const totalSize = BOARD_SIZE + MARGIN * 2;

  const files = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'];
  const ranks = ['8', '7', '6', '5', '4', '3', '2', '1'];

  return (
    <svg
      width={totalSize}
      height={totalSize}
      viewBox={`0 0 ${totalSize} ${totalSize}`}
      className="select-none"
    >
      {/* File labels (top + bottom) */}
      {files.map((f, i) => (
        <g key={`file-${f}`}>
          <text
            x={MARGIN + i * CELL_SIZE + CELL_SIZE / 2}
            y={MARGIN - 8}
            textAnchor="middle"
            fontSize="12"
            fill="var(--muted)"
            fontFamily="var(--font-sans)"
          >
            {f}
          </text>
          <text
            x={MARGIN + i * CELL_SIZE + CELL_SIZE / 2}
            y={MARGIN + BOARD_SIZE + 16}
            textAnchor="middle"
            fontSize="12"
            fill="var(--muted)"
            fontFamily="var(--font-sans)"
          >
            {f}
          </text>
        </g>
      ))}

      {/* Rank labels (left + right) */}
      {ranks.map((r, i) => (
        <g key={`rank-${r}`}>
          <text
            x={MARGIN - 12}
            y={MARGIN + i * CELL_SIZE + CELL_SIZE / 2 + 4}
            textAnchor="middle"
            fontSize="12"
            fill="var(--muted)"
            fontFamily="var(--font-sans)"
          >
            {r}
          </text>
          <text
            x={MARGIN + BOARD_SIZE + 12}
            y={MARGIN + i * CELL_SIZE + CELL_SIZE / 2 + 4}
            textAnchor="middle"
            fontSize="12"
            fill="var(--muted)"
            fontFamily="var(--font-sans)"
          >
            {r}
          </text>
        </g>
      ))}

      {/* Board squares */}
      {Array.from({ length: 8 }, (_, row) =>
        Array.from({ length: 8 }, (_, col) => {
          const isDark = (row + col) % 2 === 1;
          const sq = rowColToSquare(row, col);
          const x = MARGIN + col * CELL_SIZE;
          const y = MARGIN + row * CELL_SIZE;

          const isHighlighted = sq !== null && highlightedSquares.includes(sq);
          const isBest = sq !== null && bestMoveSquares.includes(sq);

          return (
            <g key={`${row}-${col}`}>
              <rect
                x={x}
                y={y}
                width={CELL_SIZE}
                height={CELL_SIZE}
                fill={isDark ? 'var(--color-board-dark)' : 'var(--color-board-light)'}
                className={isDark ? 'cursor-pointer' : ''}
                onClick={() => {
                  if (sq !== null) onSquareClick(sq);
                }}
              />
              {isDark && (
                <rect
                  x={x}
                  y={y}
                  width={CELL_SIZE}
                  height={CELL_SIZE}
                  fill="transparent"
                  className="cursor-pointer hover:opacity-80"
                  onClick={() => {
                    if (sq !== null) onSquareClick(sq);
                  }}
                >
                  <title>{sq !== null ? squareLabel(sq) : ''}</title>
                </rect>
              )}
              {/* Highlight overlay */}
              {isHighlighted && (
                <rect
                  x={x}
                  y={y}
                  width={CELL_SIZE}
                  height={CELL_SIZE}
                  fill="var(--color-board-highlight)"
                  pointerEvents="none"
                />
              )}
              {isBest && (
                <rect
                  x={x}
                  y={y}
                  width={CELL_SIZE}
                  height={CELL_SIZE}
                  fill="var(--color-board-best)"
                  pointerEvents="none"
                />
              )}
              {/* Piece */}
              {sq !== null && position[sq] && (
                <Piece
                  type={position[sq]!}
                  cx={x + CELL_SIZE / 2}
                  cy={y + CELL_SIZE / 2}
                  r={CELL_SIZE * 0.38}
                />
              )}
            </g>
          );
        })
      )}

      {/* Board border */}
      <rect
        x={MARGIN}
        y={MARGIN}
        width={BOARD_SIZE}
        height={BOARD_SIZE}
        fill="none"
        stroke="var(--border)"
        strokeWidth="2"
      />
    </svg>
  );
}

function squareLabel(sq: number): string {
  const [row, col] = squareToRowCol(sq);
  const file = String.fromCharCode('a'.charCodeAt(0) + col);
  const rank = 8 - row;
  return `${file}${rank} (${sq})`;
}

interface PieceProps {
  type: 'white' | 'black' | 'whiteKing' | 'blackKing';
  cx: number;
  cy: number;
  r: number;
}

function Piece({ type, cx, cy, r }: PieceProps) {
  const isWhite = type === 'white' || type === 'whiteKing';
  const isKing = type === 'whiteKing' || type === 'blackKing';

  const fill = isWhite ? 'var(--color-piece-white)' : 'var(--color-piece-black)';
  const stroke = isWhite ? 'var(--color-piece-white-stroke)' : 'var(--color-piece-black-stroke)';

  return (
    <g pointerEvents="none">
      {/* Shadow */}
      <circle cx={cx} cy={cy + 2} r={r} fill="rgba(0,0,0,0.2)" />
      {/* Main body */}
      <circle cx={cx} cy={cy} r={r} fill={fill} stroke={stroke} strokeWidth="2" />
      {/* Inner ring for depth */}
      <circle
        cx={cx}
        cy={cy}
        r={r * 0.75}
        fill="none"
        stroke={isWhite ? 'rgba(0,0,0,0.1)' : 'rgba(255,255,255,0.15)'}
        strokeWidth="1.5"
      />
      {/* King crown */}
      {isKing && (
        <g>
          <text
            x={cx}
            y={cy + r * 0.2}
            textAnchor="middle"
            fontSize={r * 0.9}
            fill="var(--color-king-crown)"
            fontWeight="bold"
            style={{ textShadow: '0 1px 2px rgba(0,0,0,0.3)' }}
          >
            ♛
          </text>
        </g>
      )}
    </g>
  );
}
