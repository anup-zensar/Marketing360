import type { components } from '@/api/generated/schema';

type Report = components['schemas']['PlanBudgetReport'];
type Ranked = components['schemas']['PlanBudgetRanked'];
type Figures = components['schemas']['PlanBudgetFigures'];
type Entry = components['schemas']['RankedEntry'];

const usd = (amount: string) => ({ amount, currency: 'USD' });

function months(): Report['periods'] {
  return Array.from({ length: 12 }, (_, i) => {
    const key = `2026-${String(i + 1).padStart(2, '0')}`;
    const state = i < 9 ? 'actual' : i === 9 ? 'inProgress' : 'planned';
    const past = i <= 9;
    return {
      key,
      state,
      budget: usd(i < 2 ? '500.00' : '0.00'),
      planned: usd(i === 2 ? '300.00' : '0.00'),
      actual: past ? usd(i === 1 ? '100.00' : '0.00') : null,
      budgetToDate: usd(i < 1 ? '500.00' : '1000.00'),
      plannedToDate: usd(i < 2 ? '0.00' : '300.00'),
      actualToDate: past ? usd(i < 1 ? '0.00' : '100.00') : null,
    };
  });
}

/** The fixture of the 040 backend tests: one plan of 1,000 USD with Helix launch and Rimula fleet. */
export function planBudgetReport(overrides: Partial<Report> = {}): Report {
  return {
    cards: {
      budget: usd('1000.00'),
      planned: usd('500.00'),
      committed: usd('400.00'),
      actual: usd('180.00'),
      leftToPlan: usd('500.00'),
      unspent: usd('820.00'),
      overAllocated: false,
      budgetUsed: 50,
      budgetSpent: 18,
      committedShare: 80,
      notYetCommitted: usd('100.00'),
    },
    period: 'month',
    periods: months(),
    notes: { notPhased: null, noDates: null, otherYears: null, actualOtherYears: usd('50.00') },
    ...overrides,
  };
}

function figures(planned: string, budget: string | null, over = false): Figures {
  return {
    budget: budget ? usd(budget) : null,
    planned: usd(planned),
    committed: usd('0.00'),
    actual: usd('0.00'),
    leftToPlan: budget ? usd(String(Number(budget) - Number(planned)) + '.00') : null,
    unspent: budget ? usd(budget) : null,
    overAllocated: over,
  };
}

const entry = (
  id: string,
  level: Entry['level'],
  name: string,
  f: Figures,
  children?: Entry[],
): Entry => ({
  id,
  level,
  name,
  planId: 'P-1001',
  figures: f,
  ...(children && { children }),
});

export function planBudgetRanked(overrides: Partial<Ranked> = {}): Ranked {
  return {
    level: 'campaign',
    sort: 'planned',
    order: 'desc',
    group: 'none',
    page: 1,
    size: 10,
    total: 2,
    trail: [],
    groups: [
      {
        key: 'none',
        label: '',
        totals: figures('500', null),
        entries: [
          entry('C-1001', 'campaign', 'Helix launch', figures('300', '600'), [
            entry('I-1001', 'initiative', 'Launch events', figures('300', '400')),
          ]),
          entry('C-1002', 'campaign', 'Rimula fleet', figures('200', '100', true), [
            entry('I-1002', 'initiative', 'Fleet days', figures('200', '100', true)),
          ]),
        ],
      },
    ],
    drillLost: false,
    ...overrides,
  };
}
