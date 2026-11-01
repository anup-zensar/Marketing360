/** Custom dashboards of feature 051 for the list and viewer tests, shaped as the API answers them. */
import type { components } from '@/api/generated/schema';
import type { Permission } from '@/session/access';
import type { CurrentUser } from '@/session/useSession';
import { testUser } from './users';

type S = components['schemas'];

export const AVAILABLE = { available: true, reason: '' };
export const HIDDEN = { available: false, reason: '' };

export const PRIYA_PERSON: S['PersonRef'] = {
  id: 'U-1002',
  name: 'Priya Raman',
  initials: 'PR',
  jobTitle: 'Marketer',
  active: true,
};
export const DANIEL_PERSON: S['PersonRef'] = {
  id: 'U-1003',
  name: 'Daniel Okoro',
  initials: 'DO',
  jobTitle: '',
  active: true,
};
export const EUROPE_REF: S['TeamRef'] = {
  id: 'T-1001',
  name: 'Lubricants Europe',
  colour: 'blue',
  active: true,
};
export const CHINA_REF: S['TeamRef'] = {
  id: 'T-1002',
  name: 'Mobility China',
  colour: 'teal',
  active: true,
};

const OWNER_ACCESS: S['CustomDashboardAccess'] = {
  edit: AVAILABLE,
  share: AVAILABLE,
  delete: AVAILABLE,
  copy: AVAILABLE,
};
export const VIEWER_ACCESS: S['CustomDashboardAccess'] = {
  edit: HIDDEN,
  share: HIDDEN,
  delete: HIDDEN,
  copy: AVAILABLE,
};

export function dashboardSummary(
  overrides: Partial<S['CustomDashboardSummary']> = {},
): S['CustomDashboardSummary'] {
  return {
    id: 'DB-1001',
    name: 'Lubricants Europe overview',
    description: 'Planned cost and leads for the European market.',
    owner: PRIYA_PERSON,
    sharedWith: [EUROPE_REF],
    widgetCount: 6,
    updatedAt: '2026-10-28T09:30:00Z',
    access: OWNER_ACCESS,
    ...overrides,
  };
}

export const TEXT_WIDGET: S['Widget'] = {
  id: 'w1',
  type: 'text',
  title: 'About these figures',
  subtitle: '',
  layout: { x: 0, y: 0, w: 4, h: 2 },
  query: null,
  display: {},
  note: 'Planned cost comes from the activities.',
};

export const BAR_WIDGET: S['Widget'] = {
  id: 'w2',
  type: 'bar',
  title: 'Planned cost by type',
  subtitle: '',
  layout: { x: 4, y: 0, w: 6, h: 4 },
  query: {
    source: 'activities',
    dimension: { field: 'activityType', dateGroup: null },
    series: null,
    measures: [{ aggregation: 'sum', field: 'plannedCost', label: '' }],
    filters: [],
    sort: { by: 'measure', measureIndex: 0, direction: 'desc' },
    limit: 50,
    groupRest: false,
  },
  display: {},
  note: '',
};

export function customDashboard(
  overrides: Partial<S['CustomDashboard']> = {},
): S['CustomDashboard'] {
  return {
    id: 'DB-1001',
    name: 'Lubricants Europe overview',
    description: 'Planned cost and leads for the European market.',
    owner: PRIYA_PERSON,
    isOwner: true,
    sharedWith: [EUROPE_REF],
    shareableTeams: [EUROPE_REF],
    filters: { dateFrom: null, dateTo: null, planIds: [], teamIds: [] },
    widgets: [TEXT_WIDGET, BAR_WIDGET],
    access: OWNER_ACCESS,
    version: 3,
    createdAt: '2026-10-01T09:00:00Z',
    createdBy: { id: 'U-1002', name: 'Priya Raman' },
    updatedAt: '2026-10-28T09:30:00Z',
    updatedBy: { id: 'U-1002', name: 'Priya Raman' },
    ...overrides,
  };
}

/** The same dashboard as Daniel sees it: shared with his team, not his own. */
export function sharedDashboard(overrides: Partial<S['CustomDashboard']> = {}) {
  return customDashboard({
    isOwner: false,
    shareableTeams: [],
    access: VIEWER_ACCESS,
    ...overrides,
  });
}

export function withPermissions(keys: Permission[]): CurrentUser {
  return testUser({
    access: {
      roles: [],
      permissions: keys.map((key) => ({ key, all: true, teams: [] })),
      administrators: [],
      noSelfApproval: true,
    },
  });
}
