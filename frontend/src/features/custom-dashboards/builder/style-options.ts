import type { ChartPalette, Widget, WidgetDisplay, WidgetType } from '../model';

export interface StyleSectionProps {
  widget: Widget;
  prefix: string;
  errorAt: (path: string) => string | undefined;
  set: (patch: Partial<WidgetDisplay>, typing?: string) => void;
}

export type StyleOption =
  | 'palette'
  | 'legend'
  | 'dataLabels'
  | 'axisTitles'
  | 'gridlines'
  | 'smooth'
  | 'stack'
  | 'orientation'
  | 'numberFormat'
  | 'zoom'
  | 'referenceLine'
  | 'measureStyles'
  | 'gauge'
  | 'kpi'
  | 'table'
  | 'note';

const CHARTS: readonly WidgetType[] = [
  'bar',
  'line',
  'area',
  'combo',
  'pie',
  'donut',
  'scatter',
  'funnel',
  'gauge',
  'heatmap',
  'treemap',
  'radar',
];
const AXES: readonly WidgetType[] = ['bar', 'line', 'area', 'combo', 'scatter', 'heatmap'];

const TYPES_BY_OPTION: Record<StyleOption, readonly WidgetType[]> = {
  palette: CHARTS,
  legend: CHARTS.filter((type) => type !== 'gauge' && type !== 'treemap'),
  dataLabels: CHARTS,
  axisTitles: AXES,
  gridlines: AXES,
  smooth: ['line', 'area', 'combo'],
  stack: ['bar', 'line', 'area'],
  orientation: ['bar', 'funnel'],
  numberFormat: [...CHARTS, 'kpi', 'table'],
  zoom: ['bar', 'line', 'area', 'combo', 'scatter'],
  referenceLine: ['bar', 'line', 'area', 'combo'],
  measureStyles: ['combo'],
  gauge: ['gauge'],
  kpi: ['kpi'],
  table: ['table'],
  note: ['text'],
};

export function shows(type: WidgetType, option: StyleOption): boolean {
  return TYPES_BY_OPTION[option].includes(type);
}

export const PALETTES: readonly { value: ChartPalette; label: string; tokens: string[] }[] = [
  { value: 'categorical', label: 'Categorical', tokens: range('--chart-cat-', 8) },
  {
    value: 'planning',
    label: 'Planning',
    tokens: ['--chart-budget', '--chart-planned', '--chart-committed', '--chart-actual'],
  },
  {
    value: 'levels',
    label: 'Levels',
    tokens: [
      '--level-plan',
      '--level-campaign',
      '--level-initiative',
      '--level-activity',
      '--level-line-item',
    ],
  },
  { value: 'sequential', label: 'Sequential', tokens: range('--chart-seq-', 5) },
];

export function defaultPalette(type: WidgetType): ChartPalette {
  return type === 'heatmap' ? 'sequential' : 'categorical';
}

export function defaultDataLabels(type: WidgetType): boolean {
  return type === 'pie' || type === 'donut' || type === 'funnel';
}

function range(prefix: string, count: number): string[] {
  return Array.from({ length: count }, (_, index) => `${prefix}${index + 1}`);
}
