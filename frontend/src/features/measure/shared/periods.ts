import type { Period } from './report-address';

/** What a time widget's period shows (FR-019): actual figures, the period in progress, or planned figures. */
export type PeriodState = 'actual' | 'inProgress' | 'planned';

/**
 * The state of each month or quarter of the plan year for today: in the current year the periods before
 * today's are actual, today's is in progress and the later ones planned; a past year is only actual, a future
 * year only planned (FR-019).
 */
export function periodStates(year: number, today: Date, period: Period): PeriodState[] {
  const count = period === 'quarter' ? 4 : 12;
  const current = period === 'quarter' ? Math.floor(today.getMonth() / 3) : today.getMonth();
  return Array.from({ length: count }, (_, i): PeriodState => {
    if (year < today.getFullYear() || (year === today.getFullYear() && i < current)) {
      return 'actual';
    }
    return year === today.getFullYear() && i === current ? 'inProgress' : 'planned';
  });
}
