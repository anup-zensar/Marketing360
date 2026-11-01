import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { describe, expect, it } from 'vitest';
import { mockApi } from '@/testing/mock-api';
import { apiError, reply } from '@/testing/replies';
import type { CustomDashboardFilters, DataSource, Widget, WidgetDisplay } from '../model';
import { NO_FILTERS } from '../model';
import { useDataSources, useWidgetData, widgetDataRequest } from './api';
import { query, widget, widgetData } from './widget-fixtures';

const ACTIVITIES: DataSource = {
  key: 'activities',
  label: 'Activities',
  description: 'Every activity you may see',
  kind: 'builtIn',
  fields: [],
  filters: { date: 'startDate', plan: true, team: true },
  owner: null,
};
const QUERY_PATH = 'POST /api/v1/data-sources/query';
const FILTERS: CustomDashboardFilters = {
  ...NO_FILTERS,
  dateFrom: '2026-01-01',
  dateTo: '2026-03-31',
};

function wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

function kpi(display: WidgetDisplay): Widget {
  return widget('kpi', display, { query: query({ dimension: null }) });
}

describe('051 FR-016: the data catalogue', () => {
  it('lists the sources the person may read', async () => {
    mockApi([reply('GET /api/v1/data-sources', { body: { items: [ACTIVITIES] } })]);
    const { result } = renderHook(() => useDataSources(), { wrapper });
    await waitFor(() => expect(result.current.data).toEqual([ACTIVITIES]));
  });
});

describe('051 FR-032: a widget’s figures', () => {
  it('asks SP&A One for the widget’s query with the dashboard filters', async () => {
    const api = mockApi([reply(QUERY_PATH, { body: widgetData() })]);
    const { result } = renderHook(() => useWidgetData(widget('bar'), FILTERS), { wrapper });

    await waitFor(() => expect(result.current.data).toEqual(widgetData()));
    expect(api.bodiesOf(QUERY_PATH)).toEqual([
      { type: 'bar', query: query(), filters: FILTERS, comparison: null },
    ]);
  });

  it('passes on the API’s refusal', async () => {
    mockApi([reply(QUERY_PATH, { status: 403, body: apiError('forbidden', 'No access') })]);
    const { result } = renderHook(() => useWidgetData(widget('bar'), null), { wrapper });
    await waitFor(() => expect(result.current.error?.kind).toBe('forbidden'));
  });

  it('asks nothing for a text note', () => {
    const api = mockApi([]);
    const { result } = renderHook(() => useWidgetData(widget('text'), null), { wrapper });
    expect(result.current.fetchStatus).toBe('idle');
    expect(api.requests).toEqual([]);
  });

  it('waits for typing to settle before asking again, showing the last figures meanwhile', async () => {
    const api = mockApi([reply(QUERY_PATH, { body: widgetData() })]);
    const first = widget('bar');
    const { result, rerender } = renderHook(
      ({ shown }) => useWidgetData(shown, null, { debounceMs: 400 }),
      { wrapper, initialProps: { shown: first } },
    );
    await waitFor(() => expect(result.current.data).toBeDefined());

    rerender({ shown: { ...first, query: query({ limit: 9 }) } });
    rerender({ shown: { ...first, query: query({ limit: 10 }) } });
    expect(result.current.data).toEqual(widgetData());
    expect(api.bodiesOf(QUERY_PATH)).toHaveLength(1);

    await waitFor(() => expect(api.bodiesOf(QUERY_PATH)).toHaveLength(2));
    expect(api.bodiesOf(QUERY_PATH)[1]).toMatchObject({ query: { limit: 10 } });
  });
});

describe('051 FR-015, FR-023: what a widget asks for', () => {
  it.each([
    ['a text note', widget('text')],
    ['a bar without its group', widget('bar', {}, { query: query({ dimension: null }) })],
    ['a heatmap without its breakdown', widget('heatmap')],
    [
      'a sum without its field',
      widget(
        'bar',
        {},
        { query: query({ measures: [{ aggregation: 'sum', field: null, label: '' }] }) },
      ),
    ],
    [
      'a pie with two measures',
      widget(
        'pie',
        {},
        {
          query: query({
            measures: [
              { aggregation: 'count', field: null, label: '' },
              { aggregation: 'count', field: null, label: '' },
            ],
          }),
        },
      ),
    ],
    ['a source not chosen', widget('bar', {}, { query: query({ source: '' }) })],
    [
      'a target comparison without its target',
      kpi({ kpi: { comparison: 'target', target: null, higherIsBetter: true } }),
    ],
  ])('asks nothing for %s', (_name, shown) => {
    expect(widgetDataRequest(shown, null)).toBeNull();
  });

  it('compares a key figure with its target', () => {
    const shown = kpi({ kpi: { comparison: 'target', target: '1000', higherIsBetter: true } });
    expect(widgetDataRequest(shown, null)?.comparison).toEqual({ kind: 'target', target: '1000' });
  });

  it('compares with the previous period only when the dashboard has a date range', () => {
    const shown = kpi({
      kpi: { comparison: 'previousPeriod', target: null, higherIsBetter: true },
    });
    expect(widgetDataRequest(shown, FILTERS)?.comparison).toEqual({
      kind: 'previousPeriod',
      target: null,
    });
    expect(widgetDataRequest(shown, NO_FILTERS)?.comparison).toBeNull();
  });
});
