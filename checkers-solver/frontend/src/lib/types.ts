// Piece types on the board
export type PieceType = 'white' | 'black' | 'whiteKing' | 'blackKing' | null;

// Tool for placement
export type Tool = 'whitePiece' | 'blackPiece' | 'whiteKing' | 'blackKing' | 'eraser';

// Board position: maps square index (0-31) to piece type
export type Position = Record<number, PieceType>;

// Move from the API
export interface MoveJSON {
  from: number;
  to: number;
  captures: number[];
  notation: string;
  promoted: boolean;
}

// Solution tree node
export interface NodeJSON {
  move: MoveJSON;
  score: number;
  children?: NodeJSON[];
  is_best: boolean;
}

// Solve API response
export interface SolveResponse {
  found: boolean;
  score: number;
  depth: number;
  moves: MoveJSON[];
  tree?: NodeJSON;
  nodes: number;
  time_ms: number;
}

// Square mapping helpers

/**
 * Convert square index (0-31) to (row, col) on 8x8 board.
 * row 0 = rank 8 (top), row 7 = rank 1 (bottom).
 */
export function squareToRowCol(sq: number): [number, number] {
  const row = Math.floor(sq / 4);
  let col = (sq % 4) * 2;
  if (row % 2 === 0) {
    col += 1; // even rows: dark squares at cols 1,3,5,7
  }
  return [row, col];
}

export function rowColToSquare(row: number, col: number): number | null {
  if (row < 0 || row > 7 || col < 0 || col > 7) return null;
  if ((row + col) % 2 === 0) return null; // light square
  return row * 4 + Math.floor(col / 2);
}

export function squareToNotation(sq: number): string {
  const [row, col] = squareToRowCol(sq);
  const file = String.fromCharCode('a'.charCodeAt(0) + col);
  const rank = 8 - row;
  return `${file}${rank}`;
}

/**
 * Convert a Position map to PDN FEN string.
 */
export function positionToFEN(position: Position, turn: 'white' | 'black'): string {
  const turnChar = turn === 'white' ? 'W' : 'B';

  const whiteParts: string[] = [];
  const blackParts: string[] = [];

  for (let sq = 0; sq < 32; sq++) {
    const piece = position[sq];
    if (!piece) continue;
    const num = sq + 1; // PDN uses 1-based
    switch (piece) {
      case 'white':
        whiteParts.push(String(num));
        break;
      case 'whiteKing':
        whiteParts.push(`K${num}`);
        break;
      case 'black':
        blackParts.push(String(num));
        break;
      case 'blackKing':
        blackParts.push(`K${num}`);
        break;
    }
  }

  return `${turnChar}:W${whiteParts.join(',')}:B${blackParts.join(',')}`;
}

/**
 * Parse a PDN FEN string into a Position and turn.
 */
export function parseFEN(fen: string): { position: Position; turn: 'white' | 'black' } {
  const position: Position = {};
  const parts = fen.split(':');
  const turn = parts[0].toUpperCase() === 'W' ? 'white' : 'black';

  for (let i = 1; i <= 2; i++) {
    const part = parts[i];
    const isBlack = part[0].toUpperCase() === 'B';
    const pieces = part.substring(1);
    if (!pieces) continue;

    for (const p of pieces.split(',')) {
      const trimmed = p.trim();
      if (!trimmed) continue;
      const isKing = trimmed.toUpperCase().startsWith('K');
      const numStr = isKing ? trimmed.substring(1) : trimmed;
      const num = parseInt(numStr, 10);
      if (isNaN(num) || num < 1 || num > 32) continue;
      const sq = num - 1;

      if (isBlack) {
        position[sq] = isKing ? 'blackKing' : 'black';
      } else {
        position[sq] = isKing ? 'whiteKing' : 'white';
      }
    }
  }

  return { position, turn: turn as 'white' | 'black' };
}
