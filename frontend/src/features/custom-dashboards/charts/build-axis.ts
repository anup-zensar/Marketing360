import type { WidgetDisplay } from '../model';
import { axisLook, axisTitles, baseOption, gridOption, itemOf, labelOption } from './build-common';
import { hasSeries, measureColumn, shareCell, toGrid } from './chart-data';
import type { ChartContext, ChartSeries } from './chart-data';
import { formatAmount, formatTick } from './format';

type MeasureStyle = NonNullable<WidgetDisplay['measureStyles']>[number];
type Stack = NonNullable<WidgetDisplay['stack']>;

const DEFAULT_STYLE: MeasureStyle = { as: 'line', axis: 'right' };
const FIRST_STYLE: MeasureStyle = { as: 'bar', axis: 'left' };

/** Bar, line and area: categories along one axis, one series per breakdown group or measure. */
export function axisChartOption(ctx: ChartContext) {
  const { widget, display } = ctx;
  const isHorizontal = widget.type === 'bar' && display.orientation === 'horizontal';
  const stack = stackOf(ctx);
  const isPercent = stack === 'percent';
  const grid = toGrid(ctx);
  const titles = axisTitles(ctx);
  const category = categoryAxis(ctx, grid.categories, isHorizontal ? titles.y : titles.x);
  const value = {
    ...valueAxis(ctx, 0, isHorizontal ? titles.x : titles.y),
    ...(isPercent
      ? { max: 100, axisLabel: { ...axisLook(ctx, '').axisLabel, formatter: '{value}%' } }
      : {}),
  };
  const series = grid.series.map((s, index) => ({
    type: widget.type === 'bar' ? 'bar' : 'line',
    name: s.name,
    data: s.cells.map((cell) => itemOf(cell && isPercent ? shareCell(cell) : cell)),
    stack: stack === 'none' ? undefined : 'total',
    smooth: display.smooth ?? false,
    areaStyle: widget.type === 'area' ? { opacity: 0.35 } : undefined,
    label: labelOption(ctx, false),
    emphasis: { focus: 'series' },
    markLine: index === 0 ? referenceLine(ctx, isHorizontal) : undefined,
  }));
  const zoomEdge = isHorizontal ? 'right' : 'bottom';
  return {
    ...baseOption(ctx, 'axis', 'categorical'),
    grid: gridOption(ctx, display.zoom ? zoomEdge : null),
    xAxis: isHorizontal ? value : category,
    yAxis: isHorizontal ? { ...category, inverse: true } : value,
    series,
    dataZoom: zoomOption(ctx, isHorizontal),
  };
}

/** Bar and line: each measure drawn as bars or a line, on the left or right value axis. */
export function comboOption(ctx: ChartContext) {
  const grid = toGrid(ctx);
  const titles = axisTitles(ctx);
  const styles = grid.series.map((s) => styleOf(ctx, s));
  const rightMeasure = styles.findIndex((style) => style.axis === 'right');
  const leftMeasure = Math.max(
    styles.findIndex((style) => style.axis === 'left'),
    0,
  );
  const yAxis = [valueAxis(ctx, leftMeasure, titles.y)];
  if (rightMeasure >= 0) {
    yAxis.push({
      ...valueAxis(ctx, rightMeasure, titles.y2),
      splitLine: { show: false, lineStyle: { color: ctx.theme.hairline } },
    });
  }
  const series = grid.series.map((s, index) => {
    const style = styles[index] ?? DEFAULT_STYLE;
    return {
      type: style.as,
      name: s.name,
      yAxisIndex: style.axis === 'right' && rightMeasure >= 0 ? 1 : 0,
      data: s.cells.map((cell) => itemOf(cell)),
      smooth: ctx.display.smooth ?? false,
      label: labelOption(ctx, false),
      emphasis: { focus: 'series' },
      markLine: index === 0 ? referenceLine(ctx, false) : undefined,
    };
  });
  return {
    ...baseOption(ctx, 'axis', 'categorical'),
    grid: gridOption(ctx, ctx.display.zoom ? 'bottom' : null),
    xAxis: categoryAxis(ctx, grid.categories, titles.x),
    yAxis,
    series,
    dataZoom: zoomOption(ctx, false),
  };
}

export function valueAxis(ctx: ChartContext, measure: number, title: string) {
  const look = axisLook(ctx, title);
  const column = measureColumn(ctx.data, measure);
  return {
    type: 'value',
    ...look,
    nameGap: title ? 48 : 0,
    axisLabel: {
      ...look.axisLabel,
      formatter: (tick: number) => formatTick(tick, column, ctx.display),
    },
  };
}

export function categoryAxis(ctx: ChartContext, categories: string[], title: string) {
  return {
    type: 'category',
    data: categories,
    ...axisLook(ctx, title),
    splitLine: { show: false },
  };
}

function stackOf(ctx: ChartContext): Stack {
  if (!hasSeries(ctx.data)) {
    return 'none';
  }
  return ctx.display.stack ?? (ctx.widget.type === 'area' ? 'stacked' : 'none');
}

function styleOf(ctx: ChartContext, series: ChartSeries): MeasureStyle {
  const chosen = ctx.display.measureStyles?.[series.measure];
  return chosen ?? (series.measure === 0 ? FIRST_STYLE : DEFAULT_STYLE);
}

function referenceLine(ctx: ChartContext, isHorizontal: boolean) {
  const line = ctx.display.referenceLine;
  if (!line) {
    return undefined;
  }
  const column = measureColumn(ctx.data, 0);
  const text = line.label || formatAmount(line.value, column, ctx.display, 'USD');
  const at = Number(line.value);
  return {
    silent: true,
    symbol: 'none',
    lineStyle: { color: ctx.theme.ink, type: 'dashed' },
    label: { formatter: () => text, color: ctx.theme.ink },
    data: [isHorizontal ? { xAxis: at } : { yAxis: at }],
  };
}

function zoomOption(ctx: ChartContext, isHorizontal: boolean) {
  if (!ctx.display.zoom) {
    return undefined;
  }
  const axis = isHorizontal ? { yAxisIndex: 0, orient: 'vertical', right: 8 } : { xAxisIndex: 0 };
  return [{ type: 'slider', ...axis, textStyle: { color: ctx.theme.ink2 } }];
}
