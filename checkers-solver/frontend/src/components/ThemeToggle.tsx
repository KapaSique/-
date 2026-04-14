import { Moon, Sun } from 'lucide-react';
import { useEffect, useState } from 'react';

export default function ThemeToggle() {
  const [dark, setDark] = useState(() => {
    if (typeof window === 'undefined') return false;
    const stored = localStorage.getItem('theme');
    if (stored) return stored === 'dark';
    return window.matchMedia('(prefers-color-scheme: dark)').matches;
  });

  useEffect(() => {
    document.documentElement.classList.toggle('dark', dark);
    localStorage.setItem('theme', dark ? 'dark' : 'light');
  }, [dark]);

  return (
    <button
      onClick={() => setDark(!dark)}
      className="p-2 rounded-lg transition-all"
      style={{
        background: 'var(--muted-bg)',
        border: '1px solid var(--border)',
        color: 'var(--fg)',
      }}
      title={dark ? 'Светлая тема' : 'Тёмная тема'}
    >
      {dark ? <Sun size={18} /> : <Moon size={18} />}
    </button>
  );
}
