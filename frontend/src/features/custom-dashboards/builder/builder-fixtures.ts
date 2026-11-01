import { NO_FILTERS } from '../model';
import type { CustomDashboard, DataField, DataSource, Widget, WidgetQuery } from '../model';

const ALLOWED = { available: true, reason: '' };
const ALL_AGGREGATIONS: DataField['aggregations'] = [
  'count',
  'countDistinct',
  'sum',
  'avg',
  'min',
  'max',
];

export const PRIYA_REF = {
  id: 'U-1002',
  name: 'Priya Raman',
  initials: 'PR',
  jobTitle: 'Marketer',
  active: true,
};

export const EUROPE_TEAM = {
  id: 'T-1001',
  name: 'Lubricants Europe',
  colour: 'blue',
  active: true,
} as const;

export const BAR_QUERY: WidgetQuery = {
  source: 'activities',
  dimension: { field: 'activityType', dateGroup: null },
  series: null,
  measures: [{ aggregation: 'sum', field: 'plannedCost', label: '' }],
  filters: [],
  sort: { by: 'measure', measureIndex: 0, direction: 'desc' },
  limit: 50,
  groupRest: false,
};

export const BAR_WIDGET: Widget = {
  id: 'w1',
  type: 'bar',
  title: 'Planned cost by type',
  subtitle: '',
  layout: { x: 0, y: 0, w: 6, h: 4 },
  query: BAR_QUERY,
  display: {},
  note: '',
};

export const NOTE_WIDGET: Widget = {
  id: 'w2',
  type: 'text',
  title: 'About',
  subtitle: '',
  layout: { x: 6, y: 0, w: 4, h: 2 },
  query: null,
  display: {},
  note: 'Planned cost of Lubricants Europe.',
};

export function dashboard(overrides: Partial<CustomDashboard> = {}): CustomDashboard {
  return {
    id: 'DB-1002',
    name: 'Lubricants Europe overview',
    description: '',
    owner: PRIYA_REF,
    isOwner: true,
    sharedWith: [],
    shareableTeams: [EUROPE_TEAM],
    filters: NO_FILTERS,
    widgets: [BAR_WIDGET, NOTE_WIDGET],
    access: { edit: ALLOWED, share: ALLOWED, delete: ALLOWED, copy: ALLOWED },
    version: 3,
    createdAt: '2026-10-30T09:00:00Z',
    createdBy: { id: PRIYA_REF.id, name: PRIYA_REF.name },
    updatedAt: '2026-10-31T09:00:00Z',
    updatedBy: { id: PRIYA_REF.id, name: PRIYA_REF.name },
    ...overrides,
  };
}

function field(key: string, label: string, kind: DataField['kind']): DataField {
  const measured = kind === 'number' || kind === 'money';
  return {
    key,
    label,
    kind,
    multi: key === 'team',
    groupable: !measured,
    aggregations: measured
      ? ALL_AGGREGATIONS
      : kind === 'date'
        ? ['count', 'countDistinct', 'min', 'max']
        : ['count', 'countDistinct'],
    finance: false,
  };
}

export const ACTIVITIES: DataSource = {
  key: 'activities',
  label: 'Activities',
  description: 'Activities of the plans you may see',
  kind: 'builtIn',
  fields: [
    field('activity', 'Activity', 'text'),
    field('activityType', 'Activity type', 'text'),
    field('status', 'Status', 'text'),
    field('team', 'Team', 'text'),
    field('startDate', 'Start date', 'date'),
    field('plannedCost', 'Planned cost', 'money'),
  ],
  filters: { date: 'startDate', plan: true, team: true },
  owner: null,
};

export const EVENT_LEADS: DataSource = {
  key: 'DS-1001',
  label: 'Event leads 2026',
  description: '',
  kind: 'dataset',
  fields: [
    field('c1', 'Event', 'text'),
    field('c2', 'Date', 'date'),
    field('c3', 'Leads', 'number'),
  ],
  filters: { date: null, plan: false, team: false },
  owner: PRIYA_REF,
};

export const SOURCES: DataSource[] = [ACTIVITIES, EVENT_LEADS];
