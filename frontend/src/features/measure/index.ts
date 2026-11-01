// The Reports area (features 039–050): every report's page from the start, so no lane edits this file (D-369).
import type { ComponentType } from 'react';
import { PlanBudgetReport } from './plan-budget/PlanBudgetReport';
import { ApprovalsReport } from './approvals/ApprovalsReport';
import { PurchaseRequestsReport } from './purchase-requests/PurchaseRequestsReport';
import { PlanDetailsReport } from './plan-details/PlanDetailsReport';
import { GsapLinkReport } from './gsap-link/GsapLinkReport';
import { UserAccessReport } from './user-access/UserAccessReport';
import { SummaryReport } from './summary/SummaryReport';
import { RevenueReport } from './revenue/RevenueReport';
import { FunnelReport } from './funnel/FunnelReport';
import type { ReportKey } from './reports';

export { ReportsPage } from './ReportsPage';
export { AttributionSettingsPage } from './attribution/AttributionSettingsPage';
export { ATTRIBUTION_SETTINGS_SECTIONS } from './attribution/settings-sections';
export { REPORT_ROUTES } from './reports';
export type { ReportKey } from './reports';
export { RankedList, ReportTable, ReportWidget } from './shared';
export type { RankedColumn, RankedData } from './shared';
export type { ReportTableData } from './shared';

export const REPORT_PAGES: Record<ReportKey, ComponentType> = {
  planBudget: PlanBudgetReport,
  approvals: ApprovalsReport,
  purchaseRequests: PurchaseRequestsReport,
  planDetails: PlanDetailsReport,
  gsapLink: GsapLinkReport,
  userAccess: UserAccessReport,
  summary: SummaryReport,
  revenue: RevenueReport,
  funnel: FunnelReport,
};
