import type { Position } from '../lib/types';

interface PieceCounterProps {
  position: Position;
}

export default function PieceCounter({ position }: PieceCounterProps) {
  let whiteCount = 0;
  let blackCount = 0;
  let whiteKings = 0;
  let blackKings = 0;

  for (let sq = 0; sq < 32; sq++) {
    const p = position[sq];
    if (p === 'white') whiteCount++;
    if (p === 'black') blackCount++;
    if (p === 'whiteKing') whiteKings++;
    if (p === 'blackKing') blackKings++;
  }

  return (
    <div className="flex gap-4 text-sm" style={{ color: 'var(--muted)' }}>
      <div className="flex items-center gap-1.5">
        <div
          className="w-3 h-3 rounded-full"
          style={{ background: '#f5f5f5', border: '1px solid #ccc' }}
        />
        <span>
          {whiteCount + whiteKings}
          {whiteKings > 0 && (
            <span className="text-xs ml-0.5">({whiteKings}D)</span>
          )}
        </span>
      </div>
      <div className="flex items-center gap-1.5">
        <div
          className="w-3 h-3 rounded-full"
          style={{ background: '#333', border: '1px solid #111' }}
        />
        <span>
          {blackCount + blackKings}
          {blackKings > 0 && (
            <span className="text-xs ml-0.5">({blackKings}D)</span>
          )}
        </span>
      </div>
    </div>
  );
}
