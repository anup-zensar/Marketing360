import type { DateFormat } from '@/format/intl';
import type { components } from '@/api/generated/schema';
import type { Widget, WidgetData, WidgetDataColumn, WidgetDisplay, WidgetValue } from '../model';
import { formatKey, formatShare, formatValue } from './format';
import type { ChartTheme } from './theme';

type WidgetKey = components['schemas']['WidgetKey'];

export interface ChartContext {
  widget: Widget;
  data: WidgetData;
  theme: ChartTheme;
  display: WidgetDisplay;
  dateFormat: DateFormat;
}

/** One drawn value: the number the chart scales by and the text people read, both from the API. */
export interface Cell {
  value: number;
  text: string;
  share: string | null;
}

export interface ChartSeries {
  name: string;
  measure: number;
  cells: (Cell | null)[];
}

export interface ChartGrid {
  categories: string[];
  series: ChartSeries[];
}

export function columnOf(data: WidgetData, key: string): WidgetDataColumn | undefined {
  return data.columns.find((column) => column.key === key);
}

export function measureColumn(data: WidgetData, measure: number): WidgetDataColumn {
  return (
    columnOf(data, `m${measure}`) ?? {
      key: `m${measure}`,
      label: `Measure ${measure + 1}`,
      kind: 'number',
      dateGroup: null,
    }
  );
}

export function numberOf(value: WidgetValue | undefined): number | null {
  const text = value?.number ?? value?.money?.amount ?? null;
  return text === null ? null : Number(text);
}

export function cellOf(
  ctx: ChartContext,
  value: WidgetValue | undefined,
  measure: number,
): Cell | null {
  const number = numberOf(value);
  if (value === undefined || number === null) {
    return null;
  }
  const text = formatValue(value, measureColumn(ctx.data, measure), ctx.display, ctx.dateFormat);
  return { value: number, text, share: value.share };
}

export function shareCell(cell: Cell): Cell {
  const share = cell.share ?? '0';
  return { value: Number(share), text: `${formatShare(share)} (${cell.text})`, share };
}

export function keyLabel(ctx: ChartContext, key: WidgetKey | null, column: string): string {
  return key === null ? '' : formatKey(key, columnOf(ctx.data, column), ctx.dateFormat);
}

export function hasSeries(data: WidgetData): boolean {
  return data.rows.some((row) => row.series !== null);
}

/** Categories from the dimension; one series per breakdown group, or else one per measure. */
export function toGrid(ctx: ChartContext): ChartGrid {
  const dimensions = uniqueKeys(ctx.data.rows.map((row) => row.dimension));
  const categories = dimensions.map((key) => keyLabel(ctx, key, 'dimension'));
  const index = new Map(dimensions.map((key, i) => [keyId(key), i]));
  if (hasSeries(ctx.data)) {
    const seriesKeys = uniqueKeys(ctx.data.rows.map((row) => row.series));
    const series = seriesKeys.map((key) => ({
      name: keyLabel(ctx, key, 'series'),
      measure: 0,
      cells: cellsFor(ctx, index, (row) => keyId(row.series) === keyId(key) && 0),
    }));
    return { categories, series };
  }
  const measures = ctx.data.columns.filter((column) => /^m\d+$/.test(column.key)).length;
  const series = Array.from({ length: Math.max(measures, 1) }, (_, measure) => ({
    name: measureColumn(ctx.data, measure).label,
    measure,
    cells: cellsFor(ctx, index, () => measure),
  }));
  return { categories, series };
}

/** A round number at or above the value, for gauges and radar spokes. */
export function niceMax(value: number): number {
  if (!(value > 0)) {
    return 1;
  }
  const magnitude = 10 ** Math.floor(Math.log10(value));
  const step = [1, 2, 2.5, 5, 10].find((s) => s * magnitude >= value) ?? 10;
  return step * magnitude;
}

type Row = WidgetData['rows'][number];

function cellsFor(
  ctx: ChartContext,
  index: Map<string, number>,
  measureOf: (row: Row) => number | false,
): (Cell | null)[] {
  const cells: (Cell | null)[] = Array.from({ length: index.size }, () => null);
  for (const row of ctx.data.rows) {
    const measure = measureOf(row);
    const at = index.get(keyId(row.dimension));
    if (measure !== false && at !== undefined) {
      cells[at] = cellOf(ctx, row.values[measure], measure);
    }
  }
  return cells;
}

export function uniqueKeys(keys: (WidgetKey | null)[]): (WidgetKey | null)[] {
  const seen = new Map<string, WidgetKey | null>();
  for (const key of keys) {
    if (!seen.has(keyId(key))) {
      seen.set(keyId(key), key);
    }
  }
  return [...seen.values()];
}

export function keyId(key: WidgetKey | null): string {
  return key === null ? 'none' : `${key.kind}:${key.value ?? ''}`;
}
