import type { TeamRef, UserAccessReport, UserAccessRow } from './api';

const EUROPE: TeamRef = { id: 'T-1001', name: 'Lubricants Europe', colour: 'blue', active: true };
const CHINA: TeamRef = { id: 'T-1002', name: 'Mobility China', colour: 'teal', active: true };

function row(id: string, name: string, overrides: Partial<UserAccessRow> = {}): UserAccessRow {
  return {
    person: {
      id,
      name,
      initials: name.slice(0, 2).toUpperCase(),
      jobTitle: 'Planner',
      active: true,
    },
    email: `${name.toLowerCase().replace(' ', '.')}@shell.com`,
    roles: [],
    teams: [],
    lastSignedInAt: null,
    ...overrides,
  };
}

/** Three people: Priya signed in, with a role and an inactive team; Alex signed in; Mei without a role, never. */
export function userAccessReport(overrides: Partial<UserAccessReport> = {}): UserAccessReport {
  return {
    total: 3,
    page: 1,
    pageSize: 20,
    sort: 'name',
    cards: { people: 3, signedInWithin30Days: 1, neverSignedIn: 1 },
    bands: [
      { key: 'within30Days', label: 'Within 30 days', count: 1 },
      { key: 'days31to90', label: '31 – 90 days', count: 1 },
      { key: 'over90Days', label: 'Over 90 days', count: 0 },
      { key: 'never', label: 'Never', count: 1 },
    ],
    recordedSince: '2026-11-01',
    rows: [
      row('U-1001', 'Alex Morgan', {
        lastSignedInAt: '2026-09-20T08:00:00Z',
        roles: [
          {
            id: 'RA-1001',
            role: { id: 'R-1001', name: 'Administrator', builtIn: 'administrator' },
            scope: { type: 'all', teams: [], text: 'All markets' },
            scopeEmpty: false,
          },
        ],
      }),
      row('U-1004', 'Mei Lin', { teams: [CHINA] }),
      row('U-1002', 'Priya Raman', {
        lastSignedInAt: '2026-10-30T09:15:00Z',
        teams: [{ ...EUROPE, id: 'T-1009', name: 'Archive', active: false }, EUROPE],
        roles: [
          {
            id: 'RA-1002',
            role: { id: 'R-1002', name: 'Planner', builtIn: 'planner' },
            scope: { type: 'myTeams', teams: [EUROPE], text: 'My teams' },
            scopeEmpty: false,
          },
        ],
      }),
    ],
    ...overrides,
  };
}
