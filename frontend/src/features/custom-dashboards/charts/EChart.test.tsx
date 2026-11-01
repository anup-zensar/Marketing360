import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { EChart } from './EChart';

const echarts = vi.hoisted(() => ({
  chart: {
    setOption: vi.fn(),
    resize: vi.fn(),
    dispose: vi.fn(),
    getDataURL: vi.fn(() => 'data:image/svg+xml,'),
    getWidth: () => 400,
    getHeight: () => 240,
  },
  init: vi.fn(),
}));

vi.mock('echarts/core', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  init: echarts.init.mockImplementation(() => echarts.chart),
}));

const OPTION = { series: [{ type: 'bar', data: [1, 2] }] };

describe('051 FR-038: the chart component (ADR-0020)', () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it('draws with the SVG renderer in English, named by its description', () => {
    render(<EChart option={OPTION} description="Bar chart: Planned cost by type" height={240} />);

    const chart = screen.getByRole('img', { name: 'Bar chart: Planned cost by type' });
    expect(echarts.init).toHaveBeenCalledWith(chart, null, { renderer: 'svg', locale: 'EN' });
    expect(echarts.chart.setOption).toHaveBeenCalledWith(
      expect.objectContaining({
        series: OPTION.series,
        aria: {
          enabled: true,
          label: { enabled: true, description: 'Bar chart: Planned cost by type' },
          decal: { show: true },
        },
      }),
      { notMerge: true },
    );
  });

  it('replaces the option when it changes, and disposes of the chart when it goes', () => {
    const { rerender, unmount } = render(<EChart option={OPTION} description="A" height={240} />);
    const next = { series: [{ type: 'line', data: [3] }] };
    rerender(<EChart option={next} description="A" height={240} />);
    expect(echarts.chart.setOption).toHaveBeenLastCalledWith(
      expect.objectContaining({ series: next.series }),
      { notMerge: true },
    );

    unmount();
    expect(echarts.chart.dispose).toHaveBeenCalledTimes(1);
  });

  it('follows the size of its place on the dashboard', () => {
    const observed: (() => void)[] = [];
    vi.stubGlobal(
      'ResizeObserver',
      class {
        constructor(callback: () => void) {
          observed.push(callback);
        }
        observe() {}
        disconnect() {}
      },
    );
    render(<EChart option={OPTION} description="A" height={240} />);
    observed.forEach((callback) => callback());
    expect(echarts.chart.resize).toHaveBeenCalledTimes(1);
  });

  it('stops moving when people prefer less motion', () => {
    vi.stubGlobal('matchMedia', (query: string) => ({ matches: query.includes('reduce') }));
    render(<EChart option={OPTION} description="A" height={240} />);
    expect(echarts.chart.setOption).toHaveBeenCalledWith(
      expect.objectContaining({ animation: false }),
      { notMerge: true },
    );
  });
});
