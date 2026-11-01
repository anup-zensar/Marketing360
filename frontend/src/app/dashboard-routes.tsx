import type { RouteObject } from 'react-router';
import {
  DashboardBuilderPage,
  DashboardListPage,
  DashboardsPage,
  DashboardViewerPage,
  DatasetPage,
  DatasetsPage,
} from './pages';
import { gated, titled } from './route-helpers';

// Feature 036's built-in dashboards and feature 051's custom ones, apart from routes.tsx only to
// keep that file within 300 lines. The list checks its own either-or permission.
export const DASHBOARD_ROUTES: RouteObject[] = [
  { path: '/dashboards', element: <DashboardListPage />, ...titled('Dashboards') },
  ...(['executive', 'regional'] as const).map((kind) => ({
    path: `/dashboards/${kind}`,
    element: gated('plans.view', <DashboardsPage kind={kind} />),
    ...titled(kind === 'executive' ? 'Executive dashboard' : 'Regional dashboard'),
  })),
  {
    path: '/dashboards/:dashboardId',
    element: gated('dashboards.view', <DashboardViewerPage />),
    ...titled('Dashboard'),
  },
  {
    path: '/builder',
    element: gated('dashboards.build', <DashboardBuilderPage />),
    ...titled('New dashboard'),
  },
  {
    path: '/builder/:dashboardId',
    element: gated('dashboards.build', <DashboardBuilderPage />),
    ...titled('Edit dashboard'),
  },
  {
    path: '/datasets',
    element: gated('dashboards.build', <DatasetsPage />),
    ...titled('Datasets'),
  },
  {
    path: '/datasets/:datasetId',
    element: gated('dashboards.build', <DatasetPage />),
    ...titled('Dataset'),
  },
];
