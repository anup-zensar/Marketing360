import fc from 'fast-check';
import { describe, expect, it } from 'vitest';
import { formatMoney } from '@/format/intl';
import { formatKey, formatValue } from './format';
import { column, key, money, num } from './widget-fixtures';

const MONEY = column('m0', 'Sum of Planned cost');
const NUMBER = column('m0', 'Count of records', 'number');
const NOTHING = { number: null, money: null, date: null, share: null };

describe('051 FR-025: number formats', () => {
  it.each([
    ['money by default for money', money('1234.5'), MONEY, {}, '$1,234.50'],
    ['plain numbers by default otherwise', num('1234.5678'), NUMBER, {}, '1,234.5678'],
    ['chosen decimals', money('1234.567'), MONEY, { decimals: 0 }, '$1,235'],
    ['compact', money('1234567'), MONEY, { numberFormat: 'compact' }, '1.2M'],
    [
      'compact with decimals',
      num('5600000000'),
      NUMBER,
      { numberFormat: 'compact', decimals: 2 },
      '5.60B',
    ],
    [
      'percent without multiplying',
      num('37.5'),
      NUMBER,
      { numberFormat: 'percent', decimals: 1 },
      '37.5%',
    ],
    ['money format of a number', num('12'), NUMBER, { numberFormat: 'money' }, '$12'],
    [
      'number format of money',
      money('12.5'),
      MONEY,
      { numberFormat: 'number', decimals: 2 },
      '12.50',
    ],
    ['nothing as a dash', NOTHING, MONEY, {}, '—'],
  ] as const)('%s', (_name, value, col, display, expected) => {
    expect(formatValue(value, col, display)).toBe(expected);
  });

  it('writes money in its own currency', () => {
    const euros = { ...money('10'), money: { amount: '10.00', currency: 'EUR' } };
    expect(formatValue(euros, MONEY, {})).toBe('€10.00');
  });

  it('writes a date in the person’s date format', () => {
    const date = { ...NOTHING, date: '2026-03-12' };
    expect(formatValue(date, column('m0', 'Latest start', 'date'), {})).toBe('12 Mar 2026');
    expect(formatValue(date, column('m0', 'Latest start', 'date'), {}, 'MM/DD/YYYY')).toBe(
      '03/12/2026',
    );
  });

  it('writes any amount the API answers exactly as the app writes money elsewhere', () => {
    const cents = fc.bigInt({ min: -(10n ** 15n), max: 10n ** 15n });
    fc.assert(
      fc.property(cents, (value) => {
        const sign = value < 0n ? '-' : '';
        const abs = value < 0n ? -value : value;
        const amount = `${sign}${abs / 100n}.${String(abs % 100n).padStart(2, '0')}`;
        expect(formatValue(money(amount), MONEY, {})).toBe(formatMoney(amount, 'USD'));
      }),
    );
  });
});

describe('051 FR-017: group names', () => {
  it.each([
    ['day', '12 Mar 2026'],
    ['week', 'Week of 12 Mar 2026'],
    ['month', 'Mar 2026'],
    ['quarter', 'Q1 2026'],
    ['year', '2026'],
  ] as const)('writes a %s bucket as %s', (group, expected) => {
    expect(formatKey(key('2026-03-12'), column('dimension', 'Start', 'date', group))).toBe(
      expected,
    );
  });

  it('keeps the API’s words for Not set, Other and text groups', () => {
    const date = column('dimension', 'Start', 'date', 'month');
    expect(formatKey({ value: null, label: 'Not set', kind: 'empty' }, date)).toBe('Not set');
    expect(formatKey({ value: null, label: 'Other', kind: 'other' }, date)).toBe('Other');
    expect(formatKey(key('Event'), column('dimension', 'Type', 'text'))).toBe('Event');
  });
});
