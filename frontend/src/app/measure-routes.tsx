import type { RouteObject } from 'react-router';
import type { ReportKey } from '@/features/measure';
import type { Permission } from '@/session/access';
import type { PageHandle } from '@/session/page-handle';
import { RequirePermission } from '@/session/RequirePermission';
import { ReportPage, ReportsPage } from './pages';

interface ReportRoute {
  key: ReportKey;
  path: string;
  title: string;
  permission: Permission;
}

// Every report's address from the start (039 FR-001, D-369), listed here so the feature loads only when opened.
const REPORTS: ReportRoute[] = [
  { key: 'planBudget', path: 'plan-budget', title: 'Plan budget', permission: 'reports.view' },
  { key: 'approvals', path: 'approvals', title: 'Approvals', permission: 'reports.view' },
  {
    key: 'purchaseRequests',
    path: 'purchase-requests',
    title: 'Purchase requests',
    permission: 'reports.view',
  },
  { key: 'planDetails', path: 'plan-details', title: 'Plan details', permission: 'reports.view' },
  { key: 'gsapLink', path: 'gsap-link', title: 'GSAP link', permission: 'reports.view' },
  { key: 'userAccess', path: 'user-access', title: 'User access', permission: 'access.manage' },
  { key: 'summary', path: 'summary', title: 'Summary', permission: 'reports.view' },
  { key: 'revenue', path: 'revenue', title: 'Revenue', permission: 'reports.view' },
  { key: 'funnel', path: 'funnel', title: 'Funnel stages', permission: 'reports.view' },
];

export const MEASURE_ROUTES: RouteObject[] = [
  {
    path: '/reports',
    element: (
      <RequirePermission permission="reports.view">
        <ReportsPage />
      </RequirePermission>
    ),
    handle: { title: 'Reports' } satisfies PageHandle,
  },
  ...REPORTS.map((report) => ({
    path: `/reports/${report.path}`,
    element: (
      <RequirePermission permission={report.permission}>
        <ReportPage reportKey={report.key} />
      </RequirePermission>
    ),
    handle: { title: report.title } satisfies PageHandle,
  })),
];
