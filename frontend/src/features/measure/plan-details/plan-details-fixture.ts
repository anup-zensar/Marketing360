import type { PlanDetailsReport, PlanDetailsValue, PlanDetailsWidget } from './api';

/** The labels the API gives the values that are not options (043 contracts/api.md). */
const LABELS: Record<string, string> = {
  none: 'Not filled in',
  doesNotApply: 'Does not apply',
  filled: 'Filled in',
  other: 'Other',
  total: 'Total',
};

/** A value of a widget for the tests: a count, and a planned cost in EUR when given. */
export function value(
  key: string,
  count: number,
  amount?: string,
  linkable = true,
): PlanDetailsValue {
  return {
    key,
    label: LABELS[key] ?? key,
    inactive: false,
    count,
    plannedCost: amount === undefined ? null : { amount, currency: 'EUR' },
    linkable: linkable && key !== 'doesNotApply',
  };
}

export function widget(overrides: Partial<PlanDetailsWidget> = {}): PlanDetailsWidget {
  return {
    key: 'attr12Count',
    title: 'Activities by Company code',
    measure: 'count',
    values: [value('GB01', 12), value('CN04', 10), value('none', 6)],
    other: null,
    total: value('total', 28, undefined, false),
    filledInPercent: null,
    countsInEach: false,
    filterParam: 'attr',
    ...overrides,
  };
}

/** The Activity level of 043 US1: 28 activities, Company code with 6 not filled in. */
export function planDetails(overrides: Partial<PlanDetailsReport> = {}): PlanDetailsReport {
  return {
    show: 'activity',
    currency: 'EUR',
    cards: {
      records: 28,
      plannedCost: { amount: '280000.00', currency: 'EUR' },
      allFilledIn: 20,
      allFilledInPercent: 71,
    },
    builtIn: [
      widget({
        key: 'statusCount',
        title: 'Activities by status',
        filterParam: 'activityStatus',
        values: [value('approved', 20), value('created', 8)],
      }),
    ],
    attributes: [
      {
        attributeId: 'SA-12',
        label: 'Company code',
        type: 'singleChoice',
        widgets: [
          widget(),
          widget({
            key: 'attr12FilledIn',
            title: 'Company code filled in',
            measure: 'filledIn',
            values: [value('filled', 22, undefined, false), value('none', 6)],
            filledInPercent: 79,
          }),
        ],
      },
    ],
    attributesSetUp: true,
    ...overrides,
  };
}
