import { baseOption, labelOption } from './build-common';
import { cellOf, keyLabel, measureColumn, niceMax } from './chart-data';
import type { ChartContext } from './chart-data';
import { formatShare, formatTick } from './format';
import { labelText } from './tooltip';

const PIE_RADIUS = '70%';
const DONUT_HOLE = '45%';

/** Pie and donut: one slice per group, labelled with SP&A One's share of the total. */
export function pieOption(ctx: ChartContext) {
  const isDonut = ctx.widget.type === 'donut';
  return {
    ...baseOption(ctx, 'item', 'categorical'),
    series: [
      {
        type: 'pie',
        name: measureColumn(ctx.data, 0).label,
        radius: [isDonut ? DONUT_HOLE : '0%', PIE_RADIUS],
        center: centreOf(ctx),
        data: slices(ctx, true),
        label: { ...labelOption(ctx, true), formatter: sliceLabel },
        itemStyle: { borderColor: ctx.theme.surface, borderWidth: 1 },
      },
    ],
  };
}

/** Funnel: the groups in the API's order (largest first), vertical or horizontal. */
export function funnelOption(ctx: ChartContext) {
  const isHorizontal = ctx.display.orientation === 'horizontal';
  return {
    ...baseOption(ctx, 'item', 'categorical'),
    series: [
      {
        type: 'funnel',
        name: measureColumn(ctx.data, 0).label,
        sort: 'none',
        orient: isHorizontal ? 'horizontal' : 'vertical',
        left: '10%',
        top: 24,
        bottom: 48,
        width: '70%',
        data: slices(ctx, false),
        label: { ...labelOption(ctx, true), position: isHorizontal ? 'top' : 'right' },
        itemStyle: { borderColor: ctx.theme.surface, borderWidth: 1 },
      },
    ],
  };
}

/** Gauge: the one value between the chosen minimum and maximum (a round number by default). */
export function gaugeOption(ctx: ChartContext) {
  const { theme, display } = ctx;
  const value = ctx.data.totals[0] ?? ctx.data.rows[0]?.values[0];
  const cell = cellOf(ctx, value, 0);
  const min = Number(display.gauge?.min ?? 0);
  const chosenMax = display.gauge?.max;
  const max = chosenMax ? Number(chosenMax) : niceMax(cell?.value ?? 0);
  const column = measureColumn(ctx.data, 0);
  const colour = theme.palettes[display.palette ?? 'categorical'][0] ?? theme.ink;
  return {
    ...baseOption(ctx, 'item', 'categorical'),
    legend: { show: false },
    series: [
      {
        type: 'gauge',
        name: column.label,
        min,
        max,
        progress: { show: true, width: 14, itemStyle: { color: colour } },
        axisLine: { lineStyle: { width: 14, color: [[1, theme.hairline]] } },
        pointer: { show: false },
        axisTick: { show: false },
        splitLine: { length: 8, lineStyle: { color: theme.ink2 } },
        axisLabel: {
          color: theme.ink2,
          formatter: (tick: number) => formatTick(tick, column, display),
        },
        title: { color: theme.ink2, offsetCenter: [0, '70%'] },
        detail: { color: theme.ink, formatter: () => cell?.text ?? '—', offsetCenter: [0, '30%'] },
        data: [{ value: cell?.value ?? 0, name: column.label, text: cell?.text ?? '—' }],
      },
    ],
  };
}

function slices(ctx: ChartContext, withShare: boolean) {
  return ctx.data.rows.flatMap((row) => {
    const cell = cellOf(ctx, row.values[0], 0);
    if (cell === null) {
      return [];
    }
    const share = withShare && cell.share !== null ? formatShare(cell.share) : null;
    return [
      {
        name: keyLabel(ctx, row.dimension, 'dimension'),
        value: cell.value,
        text: share ? `${cell.text} (${share})` : cell.text,
        share,
      },
    ];
  });
}

function sliceLabel(params: unknown): string {
  if (typeof params === 'object' && params !== null && 'name' in params && 'data' in params) {
    const { data, name } = params;
    const share =
      typeof data === 'object' && data !== null && 'share' in data && typeof data.share === 'string'
        ? data.share
        : null;
    return `${String(name)}: ${share ?? labelText(params)}`;
  }
  return labelText(params);
}

function centreOf(ctx: ChartContext): [string, string] {
  const legend = ctx.display.legend ?? { show: true, position: 'bottom' };
  if (!legend.show) {
    return ['50%', '50%'];
  }
  const shift: Record<typeof legend.position, [string, string]> = {
    top: ['50%', '55%'],
    bottom: ['50%', '45%'],
    left: ['60%', '50%'],
    right: ['40%', '50%'],
  };
  return shift[legend.position];
}
