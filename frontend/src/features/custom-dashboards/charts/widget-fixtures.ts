import type { components } from '@/api/generated/schema';
import type {
  Widget,
  WidgetData,
  WidgetDataColumn,
  WidgetDisplay,
  WidgetQuery,
  WidgetType,
  WidgetValue,
} from '../model';
import type { ChartTheme } from './theme';

type WidgetKey = components['schemas']['WidgetKey'];
type Row = WidgetData['rows'][number];

/** Stand-ins for the token colours, so tests can tell the palettes apart. */
export const THEME: ChartTheme = {
  ink: 'ink',
  ink2: 'ink-2',
  hairline: 'hairline',
  surface: 'surface',
  palettes: {
    categorical: ['cat-1', 'cat-2', 'cat-3', 'cat-4', 'cat-5', 'cat-6', 'cat-7', 'cat-8'],
    planning: ['budget', 'planned', 'committed', 'actual', 'cat-5', 'cat-6', 'cat-7', 'cat-8'],
    levels: ['plan', 'campaign', 'initiative', 'activity', 'line-item'],
    sequential: ['seq-1', 'seq-2', 'seq-3', 'seq-4', 'seq-5'],
  },
};

export function query(overrides: Partial<WidgetQuery> = {}): WidgetQuery {
  return {
    source: 'activities',
    dimension: { field: 'activityType', dateGroup: null },
    series: null,
    measures: [{ aggregation: 'sum', field: 'plannedCost', label: '' }],
    filters: [],
    sort: { by: 'measure', measureIndex: 0, direction: 'desc' },
    limit: 50,
    groupRest: false,
    ...overrides,
  };
}

export function widget(
  type: WidgetType,
  display: WidgetDisplay = {},
  overrides: Partial<Widget> = {},
): Widget {
  return {
    id: 'w1',
    type,
    title: 'Planned cost by type',
    subtitle: '',
    layout: { x: 0, y: 0, w: 6, h: 4 },
    query: type === 'text' ? null : query(),
    display,
    note: '',
    ...overrides,
  };
}

export function money(amount: string, share: string | null = null): WidgetValue {
  return { number: null, money: { amount, currency: 'USD' }, date: null, share };
}

export function num(number: string, share: string | null = null): WidgetValue {
  return { number, money: null, date: null, share };
}

export function key(label: string, value: string | null = label): WidgetKey {
  return { value, label, kind: 'value' };
}

export function column(
  columnKey: string,
  label: string,
  kind: WidgetDataColumn['kind'] = 'money',
  dateGroup: WidgetDataColumn['dateGroup'] = null,
): WidgetDataColumn {
  return { key: columnKey, label, kind, dateGroup };
}

export function row(
  dimension: WidgetKey | null,
  values: WidgetValue[],
  series: WidgetKey | null = null,
): Row {
  return { dimension, series, values };
}

export function widgetData(overrides: Partial<WidgetData> = {}): WidgetData {
  return {
    source: 'activities',
    columns: [column('dimension', 'Activity type', 'text'), column('m0', 'Sum of Planned cost')],
    rows: [
      row(key('Event'), [money('12500.00', '50.0')]),
      row(key('Digital'), [money('7500.00', '30.0')]),
      row(key('Print'), [money('5000.00', '20.0')]),
    ],
    totals: [money('25000.00')],
    groupCount: 3,
    notes: [],
    filtersApplied: { date: true, plan: true, team: true },
    comparison: null,
    ...overrides,
  };
}

/** Activity types broken down by plan: two series of one measure. */
export function seriesData(): WidgetData {
  return widgetData({
    columns: [
      column('dimension', 'Activity type', 'text'),
      column('series', 'Plan', 'text'),
      column('m0', 'Sum of Planned cost'),
    ],
    rows: [
      row(key('Event'), [money('10000.00', '80.0')], key('Mobility')),
      row(key('Event'), [money('2500.00', '20.0')], key('Lubricants')),
      row(key('Digital'), [money('7500.00', '100.0')], key('Mobility')),
    ],
    groupCount: 2,
  });
}

/** Two measures per group, as a combo or a scatter needs. */
export function twoMeasureData(): WidgetData {
  return widgetData({
    columns: [
      column('dimension', 'Activity type', 'text'),
      column('m0', 'Sum of Planned cost'),
      column('m1', 'Count of records', 'number'),
    ],
    rows: [
      row(key('Event'), [money('12500.00'), num('5')]),
      row(key('Digital'), [money('7500.00'), num('3')]),
    ],
    totals: [money('20000.00'), num('8')],
  });
}
