import type { components } from '@/api/generated/schema';

export type ReportKey = components['schemas']['ReportKey'];

export interface ReportRoute {
  key: ReportKey;
  path: string;
  name: string;
  feature: string;
}

/** The catalogue's reports in its order (039 data-model §3): their address, page title and owning feature. */
export const REPORT_ROUTES: ReportRoute[] = [
  {
    key: 'planBudget',
    path: 'plan-budget',
    name: 'Plan budget',
    feature: '040',
  },
  {
    key: 'approvals',
    path: 'approvals',
    name: 'Approvals',
    feature: '041',
  },
  {
    key: 'purchaseRequests',
    path: 'purchase-requests',
    name: 'Purchase requests',
    feature: '042',
  },
  {
    key: 'planDetails',
    path: 'plan-details',
    name: 'Plan details',
    feature: '043',
  },
  {
    key: 'gsapLink',
    path: 'gsap-link',
    name: 'GSAP link',
    feature: '044',
  },
  {
    key: 'userAccess',
    path: 'user-access',
    name: 'User access',
    feature: '045',
  },
  { key: 'summary', path: 'summary', name: 'Summary', feature: '048' },
  { key: 'revenue', path: 'revenue', name: 'Revenue', feature: '049' },
  {
    key: 'funnel',
    path: 'funnel',
    name: 'Funnel stages',
    feature: '049',
  },
];
