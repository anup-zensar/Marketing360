import type { GsapLinkReport, GsapPath } from './api';

const PATH: GsapPath = {
  planId: 'P-1001',
  plan: 'FY2026 Lubricants',
  campaign: 'Launch',
  initiative: 'Retail',
  activityId: 'A-1001',
  activity: 'Stand',
};

/** The Waiting for GSAP tab: an original routed 3 working days ago, and a top-up on a plan whose amounts are hidden. */
export function waitingReport(overrides: Partial<GsapLinkReport> = {}): GsapLinkReport {
  return {
    tab: 'waiting',
    currency: 'EUR',
    total: 2,
    page: 1,
    pageSize: 20,
    sort: '-waiting',
    finance: { plans: 1, of: 2 },
    swbs: null,
    waiting: {
      waiting: 2,
      amount: { amount: '23000.00', currency: 'EUR' },
      overTurnaround: 1,
      rows: [
        {
          purchaseRequestId: 'PR-1001',
          name: 'PR for Agency fee (Ireland)',
          vendor: 'Brightside Media BV',
          owner: 'Priya Raman',
          path: PATH,
          routedAt: '2026-03-13T16:00:00Z',
          routing: 'original',
          gsapPrNumber: '',
          amount: { amount: '23000.00', currency: 'EUR' },
          waitingWorkingDays: 3,
          overTurnaround: true,
        },
        {
          purchaseRequestId: 'PR-1002',
          name: 'Top-up for Booth',
          vendor: 'Northlight Events GmbH',
          owner: 'Mei Lin',
          path: { ...PATH, planId: 'P-1002', plan: 'FY2026 Mobility' },
          routedAt: '2026-03-18T10:00:00Z',
          routing: 'modification',
          gsapPrNumber: '1000123',
          amount: null,
          waitingWorkingDays: 0,
          overTurnaround: false,
        },
      ],
    },
    orders: null,
    ...overrides,
  };
}

/** The Purchase orders tab: one PO with two lines and a credit note; totals over three rows on two pages. */
export function ordersReport(overrides: Partial<GsapLinkReport> = {}): GsapLinkReport {
  return {
    tab: 'orders',
    currency: 'EUR',
    total: 3,
    page: 1,
    pageSize: 20,
    sort: 'lineItem',
    finance: { plans: 1, of: 2 },
    swbs: null,
    waiting: null,
    orders: {
      rows: [
        {
          lineItemId: 'L-1003',
          lineItem: 'Venue hire Berlin',
          path: PATH,
          purchaseRequestId: 'PR-1001',
          gsapPrNumber: '1000777',
          poNumber: '4500001001',
          poLines: ['10', '20'],
          poDate: '2026-01-12',
          vendor: 'Northlight Events GmbH',
          poAmount: { amount: '15800.00', currency: 'EUR' },
          invoiced: { amount: '15000.00', currency: 'EUR' },
          paid: { amount: '12000.00', currency: 'EUR' },
        },
        {
          lineItemId: 'L-1004',
          lineItem: 'Speaker fees',
          path: PATH,
          purchaseRequestId: '',
          gsapPrNumber: '',
          poNumber: '4500001002',
          poLines: ['10'],
          poDate: '2026-05-04',
          vendor: 'Keynote Speakers Ltd',
          poAmount: { amount: '5000.00', currency: 'EUR' },
          invoiced: { amount: '-500.00', currency: 'EUR' },
          paid: { amount: '0.00', currency: 'EUR' },
        },
      ],
      totals: {
        poAmount: { amount: '30800.00', currency: 'EUR' },
        invoiced: { amount: '24500.00', currency: 'EUR' },
        paid: { amount: '12000.00', currency: 'EUR' },
      },
    },
    ...overrides,
  };
}

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
