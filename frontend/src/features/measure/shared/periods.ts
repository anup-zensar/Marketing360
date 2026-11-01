import type { Period } from './report-address';

export type PeriodState = 'actual' | 'inProgress' | 'planned';

/** Before today's period actual, today's in progress, later planned; a past year all actual (FR-019). */
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
