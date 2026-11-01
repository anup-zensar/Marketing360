import { baseOption, labelOption } from './build-common';
import {
  cellOf,
  hasSeries,
  keyId,
  keyLabel,
  measureColumn,
  niceMax,
  toGrid,
  uniqueKeys,
} from './chart-data';
import type { ChartContext } from './chart-data';

interface TreeNode {
  name: string;
  value?: number;
  text: string;
  children?: TreeNode[];
  itemStyle?: { color: string };
}

/** Treemap: one box per group, split by the breakdown when there is one. */
export function treemapOption(ctx: ChartContext) {
  const { theme } = ctx;
  const palette = theme.palettes[ctx.display.palette ?? 'categorical'];
  const nodes = hasSeries(ctx.data) ? nested(ctx) : flat(ctx);
  const showValues = ctx.display.dataLabels ?? false;
  return {
    ...baseOption(ctx, 'item', 'categorical'),
    legend: { show: false },
    series: [
      {
        type: 'treemap',
        name: measureColumn(ctx.data, 0).label,
        roam: false,
        nodeClick: false,
        breadcrumb: { show: false },
        left: 0,
        right: 0,
        top: 0,
        bottom: 0,
        data: nodes.map((node, i) => ({
          ...node,
          itemStyle: { color: palette[i % palette.length] ?? theme.ink },
        })),
        label: { show: true, color: theme.surface, formatter: treeLabel(showValues) },
        upperLabel: { show: hasSeries(ctx.data), color: theme.surface, height: 20 },
        levels: [
          { itemStyle: { borderColor: theme.surface, borderWidth: 2, gapWidth: 2 } },
          { itemStyle: { borderColor: theme.surface, borderWidth: 1, gapWidth: 1 } },
        ],
      },
    ],
  };
}

/** Radar: the groups are the spokes; one shape per breakdown group, or one for the measure. */
export function radarOption(ctx: ChartContext) {
  const { theme } = ctx;
  const grid = toGrid(ctx);
  const largest = Math.max(
    0,
    ...grid.series.flatMap((s) => s.cells.map((cell) => cell?.value ?? 0)),
  );
  const max = niceMax(largest);
  return {
    ...baseOption(ctx, 'item', 'categorical'),
    radar: {
      indicator: grid.categories.map((name) => ({ name, max })),
      axisName: { color: theme.ink2 },
      axisLine: { lineStyle: { color: theme.hairline } },
      splitLine: { show: ctx.display.gridlines ?? true, lineStyle: { color: theme.hairline } },
      splitArea: { show: false },
    },
    series: [
      {
        type: 'radar',
        label: labelOption(ctx, false),
        areaStyle: { opacity: 0.2 },
        data: grid.series.map((s) => ({
          name: s.name,
          value: s.cells.map((cell) => cell?.value ?? 0),
          lines: grid.categories.map((category, i): [string, string] => [
            category,
            s.cells[i]?.text ?? '—',
          ]),
        })),
      },
    ],
  };
}

function flat(ctx: ChartContext): TreeNode[] {
  return ctx.data.rows.flatMap((row) => {
    const cell = cellOf(ctx, row.values[0], 0);
    return cell
      ? [{ name: keyLabel(ctx, row.dimension, 'dimension'), value: cell.value, text: cell.text }]
      : [];
  });
}

function nested(ctx: ChartContext): TreeNode[] {
  const groups = uniqueKeys(ctx.data.rows.map((row) => row.dimension));
  return groups.map((group) => {
    const children = ctx.data.rows
      .filter((row) => keyId(row.dimension) === keyId(group))
      .flatMap((row) => {
        const cell = cellOf(ctx, row.values[0], 0);
        return cell
          ? [{ name: keyLabel(ctx, row.series, 'series'), value: cell.value, text: cell.text }]
          : [];
      });
    return { name: keyLabel(ctx, group, 'dimension'), text: '', children };
  });
}

function treeLabel(showValues: boolean) {
  return (params: unknown): string => {
    if (typeof params !== 'object' || params === null || !('name' in params)) {
      return '';
    }
    const name = String(params.name);
    const data = 'data' in params ? params.data : null;
    const text =
      typeof data === 'object' && data !== null && 'text' in data ? String(data.text) : '';
    return showValues && text ? `${name}\n${text}` : name;
  };
}
