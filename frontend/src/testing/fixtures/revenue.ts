import type { components } from '@/api/generated/schema';

type RevenueReport = components['schemas']['RevenueReport'];
type RevenueRanked = components['schemas']['RevenueRanked'];
type RevenueFigures = components['schemas']['RevenueFigures'];
type RevenueEntry = components['schemas']['RevenueEntry'];
type FunnelReport = components['schemas']['FunnelReport'];
type FunnelRanked = components['schemas']['FunnelRanked'];
type FunnelFigures = components['schemas']['FunnelFigures'];
type FunnelEntry = components['schemas']['FunnelEntry'];

const usd = (amount: string) => ({ amount, currency: 'USD' });

function revenuePeriods(): RevenueReport['periods'] {
  return Array.from({ length: 12 }, (_, i) => {
    const key = `2026-${String(i + 1).padStart(2, '0')}`;
    const past = i < 6;
    const inProgress = i === 6;
    return {
      key,
      state: past ? 'actual' : inProgress ? 'inProgress' : 'planned',
      goal: usd('100000.00'),
      actual: past || inProgress ? usd('50000.00') : null,
      projected: !past ? usd('60000.00') : null,
      inProgress,
    };
  });
}

/** The fixture mirroring 047's own worked example: Lubricants 1.0, 273 deals, 2,730,000.00 USD (SC-001). */
export function revenueReport(overrides: Partial<RevenueReport> = {}): RevenueReport {
  return {
    cards: {
      goal: { value: usd('2730000.00'), reason: '' },
      actual: { value: usd('300000.00'), reason: '' },
      projected: { value: usd('2000000.00'), reason: '' },
      performance: { value: 73, reason: '' },
    },
    period: 'month',
    periods: revenuePeriods(),
    projection: { modelName: 'Lubricants 1.0', modelChangedAt: '2026-01-02T09:00:00Z' },
    reason: '',
    ...overrides,
  };
}

function revenueFigures(goal: string, actual: string | null): RevenueFigures {
  return {
    goal: usd(goal),
    actual: actual ? usd(actual) : null,
    projected: usd(goal),
    performance: 80,
    actualCost: null,
    roi: null,
    actualCostReason: 'attributionNotConfigured',
    roiReason: 'attributionNotConfigured',
  };
}

const revenueEntry = (
  id: string,
  level: RevenueEntry['level'],
  name: string,
  f: RevenueFigures,
  children?: RevenueEntry[],
): RevenueEntry => ({
  id,
  level,
  name,
  planId: 'P-1001',
  figures: f,
  ...(children && { children }),
});

export function revenueRanked(overrides: Partial<RevenueRanked> = {}): RevenueRanked {
  return {
    level: 'campaign',
    sort: 'projectedRevenue',
    order: 'desc',
    group: 'none',
    page: 1,
    size: 10,
    total: 1,
    trail: [],
    groups: [
      {
        key: '',
        label: '',
        totals: revenueFigures('2730000.00', '300000.00'),
        entries: [
          revenueEntry(
            'C-1001',
            'campaign',
            'Helix launch',
            revenueFigures('2730000.00', '300000.00'),
            [
              revenueEntry(
                'A-1001',
                'activity',
                'Launch film',
                revenueFigures('2730000.00', '300000.00'),
              ),
            ],
          ),
        ],
      },
    ],
    drillLost: false,
    ...overrides,
  };
}

function funnelPeriods(): FunnelReport['periods'] {
  return Array.from({ length: 12 }, (_, i) => {
    const key = `2026-${String(i + 1).padStart(2, '0')}`;
    const past = i < 6;
    const inProgress = i === 6;
    return {
      key,
      state: past ? 'actual' : inProgress ? 'inProgress' : 'planned',
      goal: 31000,
      actual: past || inProgress ? 30000 : null,
      projected: !past ? 31000 : null,
      inProgress,
    };
  });
}

export function funnelReport(overrides: Partial<FunnelReport> = {}): FunnelReport {
  return {
    stage: 'inq',
    availableStages: ['inq', 'mql'],
    uncoveredCount: 0,
    cards: {
      goal: { value: 365000, reason: '' },
      actual: { value: 180000, reason: '' },
      projected: { value: 365000, reason: '' },
      performance: { value: 100, reason: '' },
      conversion: { value: null, reason: 'noPreviousStage' },
    },
    period: 'month',
    periods: funnelPeriods(),
    projection: { modelName: 'Lubricants 1.0', modelChangedAt: '2026-01-02T09:00:00Z' },
    reason: '',
    ...overrides,
  };
}

function funnelFigures(goal: number, actual: number | null): FunnelFigures {
  return { goal, actual, projected: goal, performance: 80, conversion: 2.5 };
}

const funnelEntry = (
  id: string,
  level: FunnelEntry['level'],
  name: string,
  f: FunnelFigures,
  children?: FunnelEntry[],
): FunnelEntry => ({
  id,
  level,
  name,
  planId: 'P-1001',
  figures: f,
  ...(children && { children }),
});

export function funnelRanked(overrides: Partial<FunnelRanked> = {}): FunnelRanked {
  return {
    level: 'campaign',
    sort: 'projected',
    order: 'desc',
    group: 'none',
    page: 1,
    size: 10,
    total: 1,
    trail: [],
    groups: [
      {
        key: '',
        label: '',
        totals: funnelFigures(365000, 180000),
        entries: [
          funnelEntry('C-1001', 'campaign', 'Helix launch', funnelFigures(365000, 180000), [
            funnelEntry('A-1001', 'activity', 'Launch film', funnelFigures(365000, 180000)),
          ]),
        ],
      },
    ],
    drillLost: false,
    ...overrides,
  };
}
