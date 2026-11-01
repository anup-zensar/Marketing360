import type { ChartPalette, WidgetDisplay } from '../model';
import type { Cell, ChartContext } from './chart-data';
import { labelText, tooltipContent } from './tooltip';
import type { ItemText } from './tooltip';

type Position = NonNullable<WidgetDisplay['legend']>['position'];

const DEFAULT_LEGEND = { show: true, position: 'bottom' } as const;
const EDGE = 12;
const LEGEND_ROOM = 32;
const SIDE_LEGEND_ROOM = 140;
const ZOOM_ROOM = 40;

export interface DataItem extends ItemText {
  value: number | (number | null)[];
  name?: string;
}

/** What every chart shares: its palette, text colours, patterns, tooltip and legend. */
export function baseOption(ctx: ChartContext, trigger: 'axis' | 'item', fallback: ChartPalette) {
  const { theme } = ctx;
  return {
    color: theme.palettes[ctx.display.palette ?? fallback],
    textStyle: { color: theme.ink },
    aria: { enabled: true, decal: { show: true } },
    tooltip: {
      trigger,
      confine: true,
      formatter: tooltipContent,
      backgroundColor: theme.surface,
      borderColor: theme.hairline,
      textStyle: { color: theme.ink },
    },
    legend: legendOption(ctx),
  };
}

export function legendOption(ctx: ChartContext) {
  const { show, position } = ctx.display.legend ?? DEFAULT_LEGEND;
  const isSide = position === 'left' || position === 'right';
  return {
    show,
    type: 'scroll',
    orient: isSide ? 'vertical' : 'horizontal',
    [position]: 0,
    ...(isSide ? { top: 'middle' } : { left: 'center' }),
    textStyle: { color: ctx.theme.ink },
    pageTextStyle: { color: ctx.theme.ink2 },
  };
}

/** Room around the plot for the legend and the zoom slider. */
export function gridOption(ctx: ChartContext, zoomEdge: Position | null) {
  const room: Record<Position, number> = { top: EDGE, bottom: EDGE, left: EDGE, right: EDGE };
  const legend = ctx.display.legend ?? DEFAULT_LEGEND;
  if (legend.show) {
    const isSide = legend.position === 'left' || legend.position === 'right';
    room[legend.position] += isSide ? SIDE_LEGEND_ROOM : LEGEND_ROOM;
  }
  if (zoomEdge) {
    room[zoomEdge] += ZOOM_ROOM;
  }
  return { ...room, containLabel: true };
}

export function labelOption(ctx: ChartContext, shownByDefault: boolean) {
  return {
    show: ctx.display.dataLabels ?? shownByDefault,
    formatter: labelText,
    color: ctx.theme.ink,
  };
}

export function itemOf(cell: Cell | null, extra: Partial<DataItem> = {}): DataItem | null {
  return cell === null ? null : { value: cell.value, text: cell.text, ...extra };
}

export function axisLook(ctx: ChartContext, title: string) {
  const { theme } = ctx;
  return {
    name: title,
    nameLocation: 'middle',
    nameGap: title ? 32 : 0,
    nameTextStyle: { color: theme.ink2 },
    axisLabel: { color: theme.ink2, hideOverlap: true },
    axisLine: { lineStyle: { color: theme.hairline } },
    axisTick: { lineStyle: { color: theme.hairline } },
    splitLine: { show: ctx.display.gridlines ?? true, lineStyle: { color: theme.hairline } },
  };
}

export function axisTitles(ctx: ChartContext) {
  return { x: '', y: '', y2: '', ...ctx.display.axisTitles };
}
