import { keepPreviousData, skipToken, useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import { useEffect, useMemo, useState } from 'react';
import { api, call } from '@/api/client';
import type { ApiError } from '@/api/errors';
import { dataSourceKeys, stableHash, widgetDataKeys } from '../keys';
import type {
  CustomDashboardFilters,
  DataSource,
  Widget,
  WidgetDataRequest,
  WidgetData,
} from '../model';
import { widgetTypeInfo } from '../widget-types';

const DEADLINE_MS = 30_000;

export function useDataSources(): UseQueryResult<DataSource[], ApiError> {
  return useQuery<DataSource[], ApiError>({
    queryKey: dataSourceKeys.all,
    queryFn: async ({ signal }) =>
      (await call(api.GET('/api/v1/data-sources', { signal: bounded(signal) }))).items,
  });
}

/** What the API is asked for this widget; null for a text note or a widget not yet complete. */
export function widgetDataRequest(
  widget: Widget,
  filters: CustomDashboardFilters | null,
): WidgetDataRequest | null {
  const { query } = widget;
  if (widget.type === 'text' || query === null || !isComplete(widget, query)) {
    return null;
  }
  const kpi = widget.display.kpi;
  if (widget.type === 'kpi' && kpi?.comparison === 'target' && kpi.target === null) {
    return null;
  }
  return { type: widget.type, query, filters, comparison: comparisonOf(widget, filters) };
}

/** The widget's figures, worked out by the API; the builder waits for typing to settle first. */
export function useWidgetData(
  widget: Widget,
  filters: CustomDashboardFilters | null,
  options: { debounceMs?: number } = {},
): UseQueryResult<WidgetData, ApiError> {
  const request = useMemo(() => widgetDataRequest(widget, filters), [widget, filters]);
  const settled = useSettled(request, options.debounceMs ?? 0);
  return useQuery<WidgetData, ApiError>({
    queryKey: settled ? widgetDataKeys.one(settled) : [...widgetDataKeys.all, 'incomplete'],
    queryFn: settled
      ? ({ signal }) =>
          call(api.POST('/api/v1/data-sources/query', { body: settled, signal: bounded(signal) }))
      : skipToken,
    placeholderData: keepPreviousData,
  });
}

function useSettled(request: WidgetDataRequest | null, delayMs: number) {
  const hash = request ? stableHash(request) : '';
  const [settled, setSettled] = useState({ hash, request });
  useEffect(() => {
    if (delayMs <= 0 || hash === settled.hash) {
      return undefined;
    }
    const timer = window.setTimeout(() => setSettled({ hash, request }), delayMs);
    return () => window.clearTimeout(timer);
  }, [delayMs, hash, request, settled.hash]);
  return delayMs <= 0 ? request : settled.request;
}

function isComplete(widget: Widget, query: NonNullable<Widget['query']>): boolean {
  const info = widgetTypeInfo(widget.type);
  const [fewest, most] = info.measures;
  return (
    query.source !== '' &&
    (info.dimension !== 'required' || Boolean(query.dimension?.field)) &&
    (info.series !== 'required' || Boolean(query.series?.field)) &&
    query.measures.length >= fewest &&
    query.measures.length <= most &&
    query.measures.every((measure) => measure.aggregation === 'count' || Boolean(measure.field))
  );
}

function comparisonOf(
  widget: Widget,
  filters: CustomDashboardFilters | null,
): WidgetDataRequest['comparison'] {
  const kpi = widget.display.kpi;
  if (widget.type !== 'kpi' || !kpi) {
    return null;
  }
  if (kpi.comparison === 'target') {
    return { kind: 'target', target: kpi.target };
  }
  const hasPeriod = Boolean(filters?.dateFrom && filters.dateTo);
  return kpi.comparison === 'previousPeriod' && hasPeriod
    ? { kind: 'previousPeriod', target: null }
    : null;
}

function bounded(signal: AbortSignal): AbortSignal {
  return AbortSignal.any([signal, AbortSignal.timeout(DEADLINE_MS)]);
}
