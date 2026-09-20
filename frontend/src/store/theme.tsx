import { createSignal, createEffect } from 'solid-js';

export type Theme = 'dark' | 'light';

const initialTheme: Theme = (() => {
  if (typeof window === 'undefined') return 'dark';
  const saved = localStorage.getItem('theme_preference') as Theme;
  if (saved === 'dark' || saved === 'light') return saved;
  return window.matchMedia('(prefers-color-scheme: light)').matches ? 'light' : 'dark';
})();

const [theme, setTheme] = createSignal<Theme>(initialTheme);

createEffect(() => {
  const current = theme();
  if (typeof document !== 'undefined') {
    document.documentElement.setAttribute('data-theme', current);
    localStorage.setItem('theme_preference', current);
  }
});

export function useTheme() {
  const toggleTheme = () => {
    setTheme((prev) => (prev === 'dark' ? 'light' : 'dark'));
  };

  return {
    theme,
    toggleTheme,
    isDark: () => theme() === 'dark',
  };
}