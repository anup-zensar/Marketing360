import type { GsapLinkReport, GsapPath } from './api';

const PATH: GsapPath = {
  planId: 'P-1001',
  plan: 'FY2026 Lubricants',
  campaign: 'Launch',
  initiative: 'Retail',
  activityId: 'A-1001',
  activity: 'Stand',
};

/** The SWBS tab of two line items: one waiting 7 days (over 48 hours), one 30 hours. */
export function swbsReport(overrides: Partial<GsapLinkReport> = {}): GsapLinkReport {
  return {
    tab: 'swbs',
    currency: 'EUR',
    total: 2,
    page: 1,
    pageSize: 20,
    sort: '-waiting',
    finance: null,
    swbs: {
      waiting: 2,
      overTurnaround: 1,
      bands: [
        { key: 'under48h', label: 'Under 48 hours', count: 1 },
        { key: 'days2to5', label: '2 – 5 days', count: 0 },
        { key: 'days5to10', label: '5 – 10 days', count: 1 },
        { key: 'over10days', label: 'Over 10 days', count: 0 },
      ],
      rows: [
        {
          lineItemId: 'L-1006',
          lineItem: 'Stand build',
          path: PATH,
          owner: 'Priya Raman',
          companyCode: 'GB01',
          approvedAt: '2026-03-10T09:00:00Z',
          waitingHours: 171,
          overTurnaround: true,
        },
        {
          lineItemId: 'L-1007',
          lineItem: 'Venue hire',
          path: PATH,
          owner: '',
          companyCode: '',
          approvedAt: '2026-03-17T09:00:00Z',
          waitingHours: 30,
          overTurnaround: false,
        },
      ],
    },
    waiting: null,
    orders: null,
    ...overrides,
  };
}
