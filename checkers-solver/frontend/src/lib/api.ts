import type { SolveResponse } from './types';

const API_BASE = '/api';

export async function solvePosition(params: {
  fen: string;
  turn: string;
  goal?: string;
  maxDepth?: number;
  timeLimit?: number;
}): Promise<SolveResponse> {
  const response = await fetch(`${API_BASE}/solve`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      fen: params.fen,
      turn: params.turn,
      goal: params.goal || 'win',
      max_depth: params.maxDepth || 15,
      time_limit: params.timeLimit || 5,
    }),
  });

  if (!response.ok) {
    const err = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(err.error || `HTTP ${response.status}`);
  }

  return response.json();
}

export async function healthCheck(): Promise<boolean> {
  try {
    const res = await fetch(`${API_BASE}/health`);
    return res.ok;
  } catch {
    return false;
  }
}
