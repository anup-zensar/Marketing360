import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { readChartTheme, useChartTheme } from './theme';

function styleOf(tokens: Record<string, string>) {
  return { getPropertyValue: (name: string) => tokens[name] ?? '' };
}

describe('051 FR-024: chart colours from the design tokens', () => {
  afterEach(() => {
    document.documentElement.removeAttribute('data-theme');
    document.documentElement.removeAttribute('style');
  });

  it('builds the four palettes from their tokens only', () => {
    const theme = readChartTheme(
      styleOf({
        '--ink': ' #292929 ',
        '--chart-cat-1': '#003c88',
        '--chart-budget': '#5b6b7a',
        '--chart-cat-5': '#a3195b',
        '--level-plan': '#003c88',
        '--chart-seq-5': '#003c88',
      }),
    );
    expect(theme.ink).toBe('#292929');
    expect(theme.palettes.categorical).toHaveLength(8);
    expect(theme.palettes.categorical[0]).toBe('#003c88');
    expect(theme.palettes.planning.slice(0, 5)).toEqual(['#5b6b7a', '', '', '', '#a3195b']);
    expect(theme.palettes.levels).toHaveLength(5);
    expect(theme.palettes.sequential[4]).toBe('#003c88');
  });

  it('reads the colours again when the theme changes', async () => {
    document.documentElement.style.setProperty('--ink', 'black');
    const { result } = renderHook(() => useChartTheme());
    expect(result.current.ink).toBe('black');

    document.documentElement.style.setProperty('--ink', 'white');
    await act(async () => {
      document.documentElement.dataset.theme = 'dark';
      await Promise.resolve();
    });
    expect(result.current.ink).toBe('white');
  });
});
