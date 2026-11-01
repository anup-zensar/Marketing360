import fc from 'fast-check';
import { describe, expect, it } from 'vitest';
import { NO_FILTERS } from '../model';
import type { CustomDashboardFilters } from '../model';
import { hasAddressFilters, readFilters, writeFilters } from './filters-address';

const SAVED: CustomDashboardFilters = {
  dateFrom: '2026-01-01',
  dateTo: '2026-12-31',
  planIds: ['P-1001'],
  teamIds: ['T-1001'],
};

function read(search: string, saved = SAVED) {
  return readFilters(new URLSearchParams(search), saved);
}

describe('051 FR-037: dashboard filters in the address', () => {
  it('uses the saved filters when the address has none', () => {
    expect(read('')).toEqual(SAVED);
    expect(hasAddressFilters(new URLSearchParams(''))).toBe(false);
  });

  it('lets the address override each saved filter', () => {
    expect(read('from=2026-07-01&to=2026-12-31&plan=P-1002&plan=P-1003&team=T-1002')).toEqual({
      dateFrom: '2026-07-01',
      dateTo: '2026-12-31',
      planIds: ['P-1002', 'P-1003'],
      teamIds: ['T-1002'],
    });
  });

  it('reads an empty value as no filter, so a saved filter can be cleared', () => {
    expect(read('from=&to=&plan=&team=')).toEqual(NO_FILTERS);
  });

  it.each([
    ['a date that is not a date', 'from=2026-02-30', { dateFrom: '2026-01-01' }],
    ['a date in another format', 'to=31/12/2026', { dateTo: '2026-12-31' }],
    ['dates in the wrong order', 'from=2026-09-01&to=2026-03-01', { dateFrom: '2026-01-01' }],
    ['an unknown plan ID shape', 'plan=DROP TABLE', { planIds: ['P-1001'] }],
    ['a team ID in the plan list', 'plan=T-1002', { planIds: ['P-1001'] }],
    ['more than 20 teams', Array.from({ length: 21 }, (_, i) => `team=T-${i + 1}`).join('&'), {}],
  ])('falls back to the saved filter for %s', (_case, search, expected) => {
    expect(read(search)).toEqual({ ...SAVED, ...expected });
  });

  it('writes every filter, empty ones as an empty value', () => {
    const params = writeFilters({ ...NO_FILTERS, planIds: ['P-1002', 'P-1003'] });

    expect(params.toString()).toBe('from=&to=&plan=P-1002&plan=P-1003&team=');
    expect(hasAddressFilters(params)).toBe(true);
  });

  it('never throws and always answers filters of the right shape, whatever the address', () => {
    fc.assert(
      fc.property(fc.string(), (search) => {
        const filters = read(search);
        expect(filters.planIds.length).toBeLessThanOrEqual(20);
        expect(filters.teamIds.length).toBeLessThanOrEqual(20);
        if (filters.dateFrom !== null && filters.dateTo !== null) {
          expect(filters.dateFrom <= filters.dateTo).toBe(true);
        }
      }),
      { seed: 51 },
    );
  });

  it('reads back exactly the filters it wrote', () => {
    const date = fc
      .date({ min: new Date('2020-01-01'), max: new Date('2030-12-31'), noInvalidDate: true })
      .map((d) => d.toISOString().slice(0, 10));
    const ids = (prefix: string) =>
      fc.uniqueArray(
        fc.integer({ min: 1000, max: 9999 }).map((n) => `${prefix}-${n}`),
        { maxLength: 20 },
      );
    const filters = fc
      .record({
        dateFrom: fc.option(date, { nil: null }),
        dateTo: fc.option(date, { nil: null }),
        planIds: ids('P'),
        teamIds: ids('T'),
      })
      .filter((f) => f.dateFrom === null || f.dateTo === null || f.dateFrom <= f.dateTo);

    fc.assert(
      fc.property(filters, (written) => {
        expect(readFilters(writeFilters(written), SAVED)).toEqual(written);
      }),
      { seed: 51 },
    );
  });
});
