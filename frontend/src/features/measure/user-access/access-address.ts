import { useSearchParams } from 'react-router';
import type { components } from '@/api/generated/schema';

export type AccessSort = components['parameters']['UserAccessSort'];
export type AccessStatus = 'active' | 'inactive' | 'all';

const SORTS: AccessSort[] = [
  'name',
  '-name',
  'email',
  '-email',
  'jobTitle',
  '-jobTitle',
  'roles',
  '-roles',
  'teams',
  '-teams',
  'status',
  '-status',
  'lastSignedIn',
  '-lastSignedIn',
];
const STATUSES: AccessStatus[] = ['active', 'inactive', 'all'];

/** Rows per page, 5 to 500 (045 FR-006). */
export const PAGE_SIZES = [5, 10, 20, 50, 100, 200, 500] as const;
const DEFAULT_PAGE_SIZE = 20;

export interface AccessTable {
  q: string;
  teams: string[];
  status: AccessStatus;
  sort: AccessSort;
  page: number;
  pageSize: number;
}

export function sortOf(value: string | null): AccessSort {
  return SORTS.find((s) => s === value) ?? 'name';
}

/** Search, teams, status, sort and page in the address (FR-006); any change but the page goes back to page 1. */
export function useAccessTable() {
  const [params, setParams] = useSearchParams();
  const page = Number(params.get('page'));
  const size = Number(params.get('pageSize'));
  const table: AccessTable = {
    q: params.get('q') ?? '',
    teams: params.getAll('team'),
    status: STATUSES.find((s) => s === params.get('status')) ?? 'active',
    sort: sortOf(params.get('sort')),
    page: Number.isInteger(page) && page > 1 ? page : 1,
    pageSize: PAGE_SIZES.find((s) => s === size) ?? DEFAULT_PAGE_SIZE,
  };
  const change = (patch: Partial<AccessTable>) => {
    const next = { ...table, page: 1, ...patch };
    const written = new URLSearchParams(params);
    for (const name of ['q', 'team', 'status', 'sort', 'page', 'pageSize']) {
      written.delete(name);
    }
    const values: [string, string, boolean][] = [
      ['q', next.q, next.q !== ''],
      ['status', next.status, next.status !== 'active'],
      ['sort', next.sort, next.sort !== 'name'],
      ['page', String(next.page), next.page > 1],
      ['pageSize', String(next.pageSize), next.pageSize !== DEFAULT_PAGE_SIZE],
    ];
    for (const [name, value, kept] of values) {
      if (kept) {
        written.set(name, value);
      }
    }
    for (const team of next.teams) {
      written.append('team', team);
    }
    // Typing a search replaces its address, so Back skips the keystrokes.
    setParams(written, { replace: Object.keys(patch).length === 1 && patch.q !== undefined });
  };
  return { table, change };
}
