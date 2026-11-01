import type { WidgetDataRequest } from './model';

export const dashboardKeys = {
  all: ['custom-dashboards'] as const,
  one: (id: string) => ['custom-dashboards', id] as const,
};

export const datasetKeys = {
  all: ['datasets'] as const,
  one: (id: string) => ['datasets', id] as const,
  rows: (id: string, offset: number) => ['datasets', id, 'rows', offset] as const,
};

export const dataSourceKeys = { all: ['data-sources'] as const };

export const widgetDataKeys = {
  all: ['widget-data'] as const,
  one: (request: WidgetDataRequest) =>
    ['widget-data', request.query.source, stableHash(request)] as const,
};

/** The same text for equal values whatever the order of their object keys. */
export function stableHash(value: unknown): string {
  return JSON.stringify(value, sortedKeys);
}

function sortedKeys(_key: string, value: unknown): unknown {
  if (value === null || typeof value !== 'object' || Array.isArray(value)) {
    return value;
  }
  return Object.fromEntries(Object.entries(value).sort(([a], [b]) => (a < b ? -1 : a > b ? 1 : 0)));
}
