import type { components } from '@/api/generated/schema';

type S = components['schemas'];
export type CustomDashboard = S['CustomDashboard'];
export type CustomDashboardSummary = S['CustomDashboardSummary'];
export type CustomDashboardFilters = S['CustomDashboardFilters'];
export type Widget = S['Widget'];
export type WidgetType = S['WidgetType'];
export type WidgetLayout = S['WidgetLayout'];
export type WidgetQuery = S['WidgetQuery'];
export type WidgetDisplay = S['WidgetDisplay'];
export type WidgetData = S['WidgetData'];
export type WidgetDataRequest = S['WidgetDataRequest'];
export type WidgetValue = S['WidgetValue'];
export type WidgetDataColumn = S['WidgetDataColumn'];
export type ChartPalette = S['ChartPalette'];
export type DataSource = S['DataSource'];
export type DataField = S['DataField'];
export type Dataset = S['Dataset'];
export type DatasetSummary = S['DatasetSummary'];
export type DatasetPreview = S['DatasetPreview'];
export type TeamRef = S['TeamRef'];

export const NO_FILTERS: CustomDashboardFilters = {
  dateFrom: null,
  dateTo: null,
  planIds: [],
  teamIds: [],
};
export const GRID_COLUMNS = 12;
export const GRID_ROWS = 200;
export const MAX_WIDGETS = 30;
