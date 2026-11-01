import {
  ChartArea,
  ChartColumn,
  ChartLine,
  ChartNoAxesCombined,
  ChartPie,
  ChartScatter,
  CircleDot,
  Funnel,
  Gauge,
  Grid3x3,
  Hash,
  LayoutPanelLeft,
  Radar,
  Table,
  Type,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { WidgetType } from './model';

export type Need = 'required' | 'optional' | 'none';

export interface WidgetTypeInfo {
  type: WidgetType;
  label: string;
  group: 'charts' | 'figures' | 'text';
  icon: LucideIcon;
  size: { w: number; h: number };
  dimension: Need;
  series: Need;
  measures: readonly [min: number, max: number];
}

const CHART = { w: 6, h: 4 };

const BY_TYPE: Record<WidgetType, WidgetTypeInfo> = {
  bar: chart({
    type: 'bar',
    label: 'Bar',
    icon: ChartColumn,
    series: 'optional',
    measures: [1, 4],
  }),
  line: chart({
    type: 'line',
    label: 'Line',
    icon: ChartLine,
    series: 'optional',
    measures: [1, 4],
  }),
  area: chart({
    type: 'area',
    label: 'Area',
    icon: ChartArea,
    series: 'optional',
    measures: [1, 4],
  }),
  combo: chart({
    type: 'combo',
    label: 'Bar and line',
    icon: ChartNoAxesCombined,
    series: 'none',
    measures: [2, 4],
  }),
  pie: chart({ type: 'pie', label: 'Pie', icon: ChartPie, series: 'none', measures: [1, 1] }),
  donut: chart({
    type: 'donut',
    label: 'Donut',
    icon: CircleDot,
    series: 'none',
    measures: [1, 1],
  }),
  scatter: chart({
    type: 'scatter',
    label: 'Scatter',
    icon: ChartScatter,
    series: 'optional',
    measures: [2, 3],
  }),
  funnel: chart({
    type: 'funnel',
    label: 'Funnel',
    icon: Funnel,
    series: 'none',
    measures: [1, 1],
  }),
  gauge: {
    type: 'gauge',
    label: 'Gauge',
    group: 'charts',
    icon: Gauge,
    size: { w: 4, h: 4 },
    dimension: 'none',
    series: 'none',
    measures: [1, 1],
  },
  heatmap: chart({
    type: 'heatmap',
    label: 'Heatmap',
    icon: Grid3x3,
    series: 'required',
    measures: [1, 1],
  }),
  treemap: chart({
    type: 'treemap',
    label: 'Treemap',
    icon: LayoutPanelLeft,
    series: 'optional',
    measures: [1, 1],
  }),
  radar: chart({
    type: 'radar',
    label: 'Radar',
    icon: Radar,
    series: 'optional',
    measures: [1, 1],
  }),
  kpi: {
    type: 'kpi',
    label: 'Key figure',
    group: 'figures',
    icon: Hash,
    size: { w: 3, h: 2 },
    dimension: 'none',
    series: 'none',
    measures: [1, 1],
  },
  table: {
    type: 'table',
    label: 'Table',
    group: 'figures',
    icon: Table,
    size: { w: 6, h: 5 },
    dimension: 'optional',
    series: 'optional',
    measures: [1, 8],
  },
  text: {
    type: 'text',
    label: 'Text note',
    group: 'text',
    icon: Type,
    size: { w: 4, h: 2 },
    dimension: 'none',
    series: 'none',
    measures: [0, 0],
  },
};

export const WIDGET_TYPES: readonly WidgetTypeInfo[] = Object.values(BY_TYPE);

export function widgetTypeInfo(type: WidgetType): WidgetTypeInfo {
  return BY_TYPE[type];
}

type ChartParts = Pick<WidgetTypeInfo, 'type' | 'label' | 'icon' | 'series' | 'measures'>;

function chart(parts: ChartParts): WidgetTypeInfo {
  return { ...parts, group: 'charts', size: CHART, dimension: 'required' };
}
