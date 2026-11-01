import { describe, expect, it } from 'vitest';
import { periodStates } from './periods';

const today = new Date(2026, 9, 18); // 18 October 2026, in Q4

describe('periodStates, the Today marker (039 FR-019)', () => {
  it('in the current year: actual before today, in progress today, planned after', () => {
    const months = periodStates(2026, today, 'month');

    expect(months.slice(0, 9)).toEqual(Array(9).fill('actual'));
    expect(months[9]).toBe('inProgress');
    expect(months.slice(10)).toEqual(['planned', 'planned']);
    expect(periodStates(2026, today, 'quarter')).toEqual([
      'actual',
      'actual',
      'actual',
      'inProgress',
    ]);
  });

  it('a past year is only actual, a future year only planned', () => {
    expect(new Set(periodStates(2025, today, 'month'))).toEqual(new Set(['actual']));
    expect(periodStates(2027, today, 'quarter')).toEqual(Array(4).fill('planned'));
  });
});
