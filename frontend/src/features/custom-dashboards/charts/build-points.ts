import { axisTitles, baseOption, gridOption, labelOption } from './build-common';
import { categoryAxis, valueAxis } from './build-axis';
import { cellOf, columnOf, keyId, keyLabel, measureColumn, uniqueKeys } from './chart-data';
import type { Cell, ChartContext } from './chart-data';
import { formatTick } from './format';

type Row = ChartContext['data']['rows'][number];

const SMALLEST_POINT = 6;
const POINT_GROWTH = 30;
const BINS = 5;

/** Scatter: each group is a point at (first measure, second measure), sized by a third. */
export function scatterOption(ctx: ChartContext) {
  const titles = axisTitles(ctx);
  const groups = uniqueKeys(ctx.data.rows.map((row) => row.series));
  const largest = Math.max(
    0,
    ...ctx.data.rows.map((row) => cellOf(ctx, row.values[2], 2)?.value ?? 0),
  );
  const series = groups.map((group) => ({
    type: 'scatter',
    name: group === null ? measureColumn(ctx.data, 1).label : keyLabel(ctx, group, 'series'),
    data: ctx.data.rows
      .filter((row) => keyId(row.series) === keyId(group))
      .map((row) => pointOf(ctx, row, largest))
      .filter((point) => point !== null),
    label: { ...labelOption(ctx, false), formatter: '{b}' },
    emphasis: { focus: 'series' },
  }));
  return {
    ...baseOption(ctx, 'item', 'categorical'),
    grid: gridOption(ctx, ctx.display.zoom ? 'bottom' : null),
    xAxis: { ...valueAxis(ctx, 0, titles.x || measureColumn(ctx.data, 0).label), scale: true },
    yAxis: { ...valueAxis(ctx, 1, titles.y || measureColumn(ctx.data, 1).label), scale: true },
    series,
    dataZoom: ctx.display.zoom ? [{ type: 'slider', xAxisIndex: 0 }] : undefined,
  };
}

/** Heatmap: groups across, breakdown down, five bins of the Sequential palette. */
export function heatmapOption(ctx: ChartContext) {
  const titles = axisTitles(ctx);
  const across = uniqueKeys(ctx.data.rows.map((row) => row.dimension));
  const down = uniqueKeys(ctx.data.rows.map((row) => row.series));
  const xLabels = across.map((key) => keyLabel(ctx, key, 'dimension'));
  const yLabels = down.map((key) => keyLabel(ctx, key, 'series'));
  const cells = ctx.data.rows.map((row) => ({
    x: across.findIndex((key) => keyId(key) === keyId(row.dimension)),
    y: down.findIndex((key) => keyId(key) === keyId(row.series)),
    cell: cellOf(ctx, row.values[0], 0),
  }));
  const data = cells
    .filter((c): c is { x: number; y: number; cell: Cell } => c.cell !== null)
    .map(({ x, y, cell }) => ({
      value: [x, y, cell.value],
      text: cell.text,
      lines: [
        [columnOf(ctx.data, 'dimension')?.label ?? '', xLabels[x] ?? ''],
        [columnOf(ctx.data, 'series')?.label ?? '', yLabels[y] ?? ''],
        [measureColumn(ctx.data, 0).label, cell.text],
      ],
    }));
  const base = baseOption(ctx, 'item', 'sequential');
  return {
    ...base,
    grid: gridOption(ctx, null),
    xAxis: { ...categoryAxis(ctx, xLabels, titles.x), splitArea: { show: true } },
    yAxis: { ...categoryAxis(ctx, yLabels, titles.y), splitArea: { show: true } },
    visualMap: {
      type: 'piecewise',
      show: base.legend.show,
      dimension: 2,
      orient: base.legend.orient,
      ...positionOf(base.legend),
      pieces: bins(
        ctx,
        data.map((item) => item.value[2] ?? 0),
        base.color,
      ),
      textStyle: { color: ctx.theme.ink },
    },
    legend: { show: false },
    series: [
      {
        type: 'heatmap',
        name: measureColumn(ctx.data, 0).label,
        data,
        label: labelOption(ctx, false),
        itemStyle: { borderColor: ctx.theme.surface, borderWidth: 1 },
      },
    ],
  };
}

/** Equal-width bins between the smallest and largest value, each exactly one palette colour. */
export function bins(ctx: ChartContext, values: number[], colours: string[]) {
  if (values.length === 0) {
    return [];
  }
  const low = Math.min(...values);
  const high = Math.max(...values);
  const count = low === high ? 1 : BINS;
  const width = (high - low) / count;
  const column = measureColumn(ctx.data, 0);
  return Array.from({ length: count }, (_, i) => {
    const from = low + width * i;
    const to = i === count - 1 ? high : low + width * (i + 1);
    const colour = colours[Math.round((i * (colours.length - 1)) / Math.max(count - 1, 1))];
    return {
      gte: from,
      ...(i === count - 1 ? { lte: to } : { lt: to }),
      color: colour,
      label: `${formatTick(from, column, ctx.display)} – ${formatTick(to, column, ctx.display)}`,
    };
  });
}

function pointOf(ctx: ChartContext, row: Row, largest: number) {
  const x = cellOf(ctx, row.values[0], 0);
  const y = cellOf(ctx, row.values[1], 1);
  if (x === null || y === null) {
    return null;
  }
  const size = cellOf(ctx, row.values[2], 2);
  const lines: [string, string][] = [
    [measureColumn(ctx.data, 0).label, x.text],
    [measureColumn(ctx.data, 1).label, y.text],
  ];
  if (size) {
    lines.push([measureColumn(ctx.data, 2).label, size.text]);
  }
  const grows = size && largest > 0 ? Math.sqrt(Math.max(size.value, 0) / largest) : 0.3;
  return {
    name: keyLabel(ctx, row.dimension, 'dimension'),
    value: [x.value, y.value],
    symbolSize: SMALLEST_POINT + POINT_GROWTH * grows,
    lines,
  };
}

function positionOf(legend: Record<string, unknown>) {
  const edges = ['top', 'bottom', 'left', 'right'] as const;
  return Object.fromEntries(edges.filter((e) => e in legend).map((e) => [e, legend[e]]));
}
