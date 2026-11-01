/* eslint-disable @typescript-eslint/no-unused-vars -- placeholders with the lane's signatures until it lands */
import { skipToken, useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import type { EChartsCoreOption } from 'echarts/core';
import { createElement } from 'react';
import type { ReactElement } from 'react';
import type { ApiError } from '@/api/errors';
import { dataSourceKeys, widgetDataKeys } from '../keys';
import type {
  ChartPalette,
  CustomDashboardFilters,
  DataSource,
  Widget,
  WidgetData,
  WidgetDataColumn,
  WidgetDataRequest,
  WidgetDisplay,
  WidgetValue,
} from '../model';

const COMING = 'Coming in this feature';

export interface ChartHandle {
  toPng(): Promise<Blob>;
}

export interface EChartProps {
  option: EChartsCoreOption;
  description: string;
  height: number;
  onReady?: (handle: ChartHandle) => void;
}

export function EChart(props: EChartProps): ReactElement {
  return createElement('div', { role: 'img', 'aria-label': props.description }, COMING);
}

export interface ChartTheme {
  ink: string;
  ink2: string;
  hairline: string;
  surface: string;
  palettes: Record<ChartPalette, string[]>;
}

export function useChartTheme(): ChartTheme {
  return {
    ink: '',
    ink2: '',
    hairline: '',
    surface: '',
    palettes: { categorical: [], planning: [], levels: [], sequential: [] },
  };
}

export function buildOption(
  _widget: Widget,
  _data: WidgetData,
  _theme: ChartTheme,
): EChartsCoreOption {
  return {};
}

export function describeWidget(widget: Widget, _data: WidgetData): string {
  return widget.title;
}

export function formatValue(
  _value: WidgetValue,
  _column: WidgetDataColumn,
  _display: WidgetDisplay,
): string {
  return '';
}

export function widgetDataRequest(
  _widget: Widget,
  _filters: CustomDashboardFilters | null,
): WidgetDataRequest | null {
  return null;
}

export function useDataSources(): UseQueryResult<DataSource[], ApiError> {
  return useQuery<DataSource[], ApiError>({ queryKey: dataSourceKeys.all, queryFn: skipToken });
}

export function useWidgetData(
  widget: Widget,
  _filters: CustomDashboardFilters | null,
  _options?: { debounceMs?: number },
): UseQueryResult<WidgetData, ApiError> {
  return useQuery<WidgetData, ApiError>({
    queryKey: [...widgetDataKeys.all, widget.id],
    queryFn: skipToken,
  });
}

export interface WidgetViewProps {
  widget: Widget;
  filters: CustomDashboardFilters | null;
  mode: 'view' | 'edit';
  dashboardName: string;
  viewerOwnsDashboard: boolean;
}

export function WidgetView(_props: WidgetViewProps): ReactElement {
  return createElement('p', null, COMING);
}

export function WidgetTable(_props: { widget: Widget; data: WidgetData }): ReactElement {
  return createElement('p', null, COMING);
}
