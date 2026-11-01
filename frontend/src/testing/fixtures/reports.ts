import type { components } from '@/api/generated/schema';

type ReportFrame = components['schemas']['ReportFrame'];
type ReportCatalogue = components['schemas']['ReportCatalogue'];

/** The Reports page's catalogue for a planner: every report but User access, Plan details available. */
export function catalogue(): ReportCatalogue {
  const entry = (
    key: ReportCatalogue['items'][number]['key'],
    path: string,
    name: string,
    available = false,
  ) => ({
    key,
    path,
    name,
    purpose: `What ${name} answers.`,
    available,
  });
  return {
    items: [
      entry('planBudget', 'plan-budget', 'Plan budget'),
      entry('approvals', 'approvals', 'Approvals'),
      entry('purchaseRequests', 'purchase-requests', 'Purchase requests'),
      entry('planDetails', 'plan-details', 'Plan details', true),
      entry('gsapLink', 'gsap-link', 'GSAP link'),
      entry('summary', 'summary', 'Summary'),
      entry('revenue', 'revenue', 'Revenue'),
      entry('funnel', 'funnel', 'Funnel stages'),
    ],
  };
}

/** A report frame of 2026 for Priya: two plans, two teams (hers first), Brand with an inactive option still used. */
export function reportFrame(overrides: Partial<ReportFrame> = {}): ReportFrame {
  return {
    report: 'planBudget',
    filter: {
      year: 2026,
      plans: [],
      planTypes: [],
      teams: [],
      owners: [],
      activityTypes: [],
      activityStatuses: [],
      attributes: [],
      period: 'month',
      show: null,
      currency: 'EUR',
      tables: [],
    },
    choices: {
      years: [2025, 2026],
      plans: [
        { id: 'P-1001', name: 'FY2026 Lubricants', planType: 'operational', teamIds: ['T-1001'] },
        { id: 'P-1002', name: 'FY2026 Mobility', planType: 'strategic', teamIds: ['T-1002'] },
      ],
      planTypes: ['operational', 'strategic', 'crossMarket'],
      teams: [
        { id: 'T-1002', name: 'Mobility China', mine: false },
        { id: 'T-1001', name: 'Lubricants Europe', mine: true },
      ],
      owners: [
        { id: 'U-1003', name: 'Daniel Okoro' },
        { id: 'U-1002', name: 'Priya Raman' },
      ],
      activityTypes: [{ id: 'AT-1001', name: 'Event' }],
      activityStatuses: ['created', 'submitted', 'declined', 'approved', 'inProgress', 'complete'],
      attributeGroups: [
        {
          level: 'activity',
          attributeId: 'SA-1022',
          label: 'Brand',
          values: [
            { id: 'SO-1301', label: 'Helix', active: true },
            { id: 'SO-1302', label: 'Pennzoil', active: false },
            { id: 'none', label: 'Not filled in', active: true },
          ],
        },
      ],
      currencies: [
        { code: 'EUR', name: 'Euro' },
        { code: 'USD', name: 'US dollar' },
      ],
    },
    inView: { plans: 2, campaigns: 6, initiatives: 2, activities: 14, lineItems: 30 },
    finance: { plans: 2, of: 2 },
    leftOut: 0,
    displayCurrency: 'EUR',
    generatedAt: '2026-10-18T14:05:00Z',
    financeAsOf: null,
    ...overrides,
  };
}
