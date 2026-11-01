import type { Widget, WidgetLayout, WidgetQuery, WidgetType } from '../model';
import { widgetTypeInfo } from '../widget-types';

type WidgetSort = WidgetQuery['sort'];

export const DEFAULT_LIMIT = 50;
const MAX_WIDGET_NUMBER = 999;
const BY_DIMENSION_TYPES: readonly WidgetType[] = ['line', 'area', 'combo', 'heatmap'];

export function newWidget(type: WidgetType, id: string, layout: WidgetLayout): Widget {
  return {
    id,
    type,
    title: '',
    subtitle: '',
    layout,
    query: type === 'text' ? null : emptyQuery(type),
    display: {},
    note: '',
  };
}

export function emptyQuery(type: WidgetType): WidgetQuery {
  return {
    source: '',
    dimension: null,
    series: null,
    measures: [],
    filters: [],
    sort: defaultSort(type, false),
    limit: DEFAULT_LIMIT,
    groupRest: false,
  };
}

export function defaultSort(type: WidgetType, isDateDimension: boolean): WidgetSort {
  const byDimension = type !== 'funnel' && (isDateDimension || BY_DIMENSION_TYPES.includes(type));
  return byDimension
    ? { by: 'dimension', measureIndex: 0, direction: 'asc' }
    : { by: 'measure', measureIndex: 0, direction: 'desc' };
}

export function typeName(type: WidgetType): string {
  const info = widgetTypeInfo(type);
  return info.group === 'charts' ? `${info.label} chart` : info.label;
}

export function widgetName(widget: Widget): string {
  return widget.title.trim() || typeName(widget.type);
}

export function nextWidgetId(widgets: readonly Widget[]): string | null {
  const used = widgets.map((widget) => Number(/^w(\d+)$/.exec(widget.id)?.[1] ?? 0));
  const next = Math.max(0, ...used) + 1;
  if (next <= MAX_WIDGET_NUMBER) {
    return `w${next}`;
  }
  const free = Array.from({ length: MAX_WIDGET_NUMBER }, (_, index) => index + 1).find(
    (number) => !used.includes(number),
  );
  return free === undefined ? null : `w${free}`;
}

type MeasureStyle = NonNullable<Widget['display']['measureStyles']>[number];

export function defaultMeasureStyle(index: number): MeasureStyle {
  return index === 0 ? { as: 'bar', axis: 'left' } : { as: 'line', axis: 'right' };
}

export function tableColumnKeys(query: WidgetQuery | null): string[] {
  if (!query) {
    return [];
  }
  return [
    ...(query.dimension ? ['dimension'] : []),
    ...(query.series ? ['series'] : []),
    ...query.measures.map((_, index) => `m${index}`),
  ];
}

export function fitDisplay(widget: Widget): Widget {
  const { display, query } = widget;
  const styles = display.measureStyles;
  const columns = display.table?.columns;
  const keys = tableColumnKeys(query);
  return {
    ...widget,
    display: {
      ...display,
      ...(styles && {
        measureStyles: (query?.measures ?? []).map(
          (_, index) => styles[index] ?? defaultMeasureStyle(index),
        ),
      }),
      ...(display.table &&
        columns && {
          table: { ...display.table, columns: columns.filter((key) => keys.includes(key)) },
        }),
    },
  };
}
