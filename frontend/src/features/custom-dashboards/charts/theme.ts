import { useSyncExternalStore } from 'react';
import type { ChartPalette } from '../model';

export interface ChartTheme {
  ink: string;
  ink2: string;
  hairline: string;
  surface: string;
  palettes: Record<ChartPalette, string[]>;
}

const CATEGORICAL = range('--chart-cat-', 8);
const PALETTE_TOKENS: Record<ChartPalette, string[]> = {
  categorical: CATEGORICAL,
  planning: [
    '--chart-budget',
    '--chart-planned',
    '--chart-committed',
    '--chart-actual',
    ...CATEGORICAL.slice(4),
  ],
  levels: [
    '--level-plan',
    '--level-campaign',
    '--level-initiative',
    '--level-activity',
    '--level-line-item',
  ],
  sequential: range('--chart-seq-', 5),
};

/** The chart colours of the theme shown, read again whenever the theme changes. */
export function useChartTheme(): ChartTheme {
  return useSyncExternalStore(watchTheme, currentChartTheme);
}

export function readChartTheme(style: Pick<CSSStyleDeclaration, 'getPropertyValue'>): ChartTheme {
  const token = (name: string) => style.getPropertyValue(name).trim();
  const palettes: Record<ChartPalette, string[]> = {
    categorical: PALETTE_TOKENS.categorical.map(token),
    planning: PALETTE_TOKENS.planning.map(token),
    levels: PALETTE_TOKENS.levels.map(token),
    sequential: PALETTE_TOKENS.sequential.map(token),
  };
  return {
    ink: token('--ink'),
    ink2: token('--ink-2'),
    hairline: token('--hairline'),
    surface: token('--surface'),
    palettes,
  };
}

let cached: { theme: string; chartTheme: ChartTheme } | null = null;

function currentChartTheme(): ChartTheme {
  const theme = document.documentElement.dataset.theme ?? 'light';
  if (cached?.theme !== theme) {
    cached = { theme, chartTheme: readChartTheme(getComputedStyle(document.documentElement)) };
  }
  return cached.chartTheme;
}

function watchTheme(onChange: () => void): () => void {
  const observer = new MutationObserver(onChange);
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
  return () => observer.disconnect();
}

function range(prefix: string, count: number): string[] {
  return Array.from({ length: count }, (_, index) => `${prefix}${index + 1}`);
}
