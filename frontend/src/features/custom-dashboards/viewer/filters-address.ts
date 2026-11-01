import { z } from 'zod';
import type { CustomDashboardFilters } from '../model';

const KEYS = ['from', 'to', 'plan', 'team'] as const;
const MAX_IDS = 20;

const DATE = z
  .string()
  .regex(/^\d{4}-\d{2}-\d{2}$/)
  .refine(isCalendarDate);
const PLAN_IDS = z
  .array(z.string().regex(/^P-\d{1,9}$/))
  .max(MAX_IDS)
  .transform(unique);
const TEAM_IDS = z
  .array(z.string().regex(/^T-\d{1,9}$/))
  .max(MAX_IDS)
  .transform(unique);

// An empty value ("from=", "plan=") means "none", so a viewer can clear a saved filter.
export function readFilters(
  params: URLSearchParams,
  saved: CustomDashboardFilters,
): CustomDashboardFilters {
  const dateFrom = readDate(params, 'from', saved.dateFrom);
  const dateTo = readDate(params, 'to', saved.dateTo);
  const datesInOrder = dateFrom === null || dateTo === null || dateFrom <= dateTo;
  return {
    dateFrom: datesInOrder ? dateFrom : saved.dateFrom,
    dateTo: datesInOrder ? dateTo : saved.dateTo,
    planIds: readIds(params, 'plan', PLAN_IDS, saved.planIds),
    teamIds: readIds(params, 'team', TEAM_IDS, saved.teamIds),
  };
}

export function writeFilters(filters: CustomDashboardFilters): URLSearchParams {
  const params = new URLSearchParams();
  params.set('from', filters.dateFrom ?? '');
  params.set('to', filters.dateTo ?? '');
  writeIds(params, 'plan', filters.planIds);
  writeIds(params, 'team', filters.teamIds);
  return params;
}

export function hasAddressFilters(params: URLSearchParams): boolean {
  return KEYS.some((key) => params.has(key));
}

function readDate(params: URLSearchParams, key: string, fallback: string | null): string | null {
  const value = params.get(key);
  if (value === null) {
    return fallback;
  }
  if (value === '') {
    return null;
  }
  const parsed = DATE.safeParse(value);
  return parsed.success ? parsed.data : fallback;
}

function readIds(
  params: URLSearchParams,
  key: string,
  schema: typeof PLAN_IDS,
  fallback: string[],
): string[] {
  if (!params.has(key)) {
    return fallback;
  }
  const parsed = schema.safeParse(params.getAll(key).filter((value) => value !== ''));
  return parsed.success ? parsed.data : fallback;
}

function writeIds(params: URLSearchParams, key: string, ids: string[]) {
  if (ids.length === 0) {
    params.set(key, '');
    return;
  }
  for (const id of ids) {
    params.append(key, id);
  }
}

function unique(ids: string[]): string[] {
  return [...new Set(ids)];
}

function isCalendarDate(text: string): boolean {
  const [year = 0, month = 0, day = 0] = text.split('-').map(Number);
  const date = new Date(Date.UTC(year, month - 1, day));
  return (
    date.getUTCFullYear() === year && date.getUTCMonth() === month - 1 && date.getUTCDate() === day
  );
}
