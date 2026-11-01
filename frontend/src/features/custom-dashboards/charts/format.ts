import { exact } from '@/format/decimal';
import { DEFAULT_DATE_FORMAT, formatDate, formatMonth } from '@/format/intl';
import type { DateFormat } from '@/format/intl';
import type { components } from '@/api/generated/schema';
import type { WidgetDataColumn, WidgetDisplay, WidgetValue } from '../model';

type WidgetKey = components['schemas']['WidgetKey'];
type NumberFormat = NonNullable<WidgetDisplay['numberFormat']>;

const NOTHING = '—';

export function formatValue(
  value: WidgetValue,
  column: WidgetDataColumn,
  display: WidgetDisplay,
  dateFormat: DateFormat = DEFAULT_DATE_FORMAT,
): string {
  if (value.date !== null) {
    return formatDate(value.date, dateFormat);
  }
  if (value.money !== null) {
    return formatAmount(value.money.amount, column, display, value.money.currency);
  }
  if (value.number !== null) {
    return formatAmount(value.number, column, display, 'USD');
  }
  return NOTHING;
}

export function numberFormatOf(column: WidgetDataColumn, display: WidgetDisplay): NumberFormat {
  return display.numberFormat ?? (column.kind === 'money' ? 'money' : 'number');
}

export function formatAmount(
  amount: string,
  column: WidgetDataColumn,
  display: WidgetDisplay,
  currency: string,
): string {
  const format = numberFormatOf(column, display);
  const digits = fractionDigits(display.decimals, format, column.kind);
  return numberFormatter(format, digits, currency).format(exact(amount)) + percentSign(format);
}

/** Axis ticks are worked out by the chart, so they are numbers, written short. */
export function formatTick(tick: number, column: WidgetDataColumn, display: WidgetDisplay): string {
  const format = numberFormatOf(column, display);
  if (format === 'percent') {
    return `${compact(1).format(tick)}%`;
  }
  return format === 'money' ? compactMoney.format(tick) : compact(1).format(tick);
}

export function formatShare(share: string): string {
  return `${share}%`;
}

/** A group's name; date buckets in the person's date format. */
export function formatKey(
  key: WidgetKey,
  column: WidgetDataColumn | undefined,
  dateFormat: DateFormat = DEFAULT_DATE_FORMAT,
): string {
  if (key.kind !== 'value' || key.value === null || column?.kind !== 'date') {
    return key.label;
  }
  return formatBucket(key.value, column.dateGroup, dateFormat);
}

function formatBucket(
  day: string,
  group: WidgetDataColumn['dateGroup'],
  dateFormat: DateFormat,
): string {
  const [year = '', month = ''] = day.split('-');
  switch (group) {
    case 'year':
      return year;
    case 'quarter':
      return `Q${Math.ceil(Number(month) / 3)} ${year}`;
    case 'month':
      return formatMonth(day.slice(0, 7), dateFormat);
    case 'week':
      return `Week of ${formatDate(day, dateFormat)}`;
    case 'day':
    case null:
      return formatDate(day, dateFormat);
    default: {
      const unknown: never = group;
      return unknown;
    }
  }
}

interface Digits {
  min: number;
  max: number;
}

function fractionDigits(
  decimals: number | undefined,
  format: NumberFormat,
  kind: WidgetDataColumn['kind'],
): Digits {
  if (decimals !== undefined) {
    return { min: decimals, max: decimals };
  }
  if (format === 'compact') {
    return { min: 0, max: 1 };
  }
  return kind === 'money' && format === 'money' ? { min: 2, max: 2 } : { min: 0, max: 4 };
}

function numberFormatter(format: NumberFormat, digits: Digits, currency: string) {
  const fractions = { minimumFractionDigits: digits.min, maximumFractionDigits: digits.max };
  switch (format) {
    case 'money':
      return new Intl.NumberFormat('en-US', { style: 'currency', currency, ...fractions });
    case 'compact':
      return new Intl.NumberFormat('en-US', { notation: 'compact', ...fractions });
    case 'number':
    case 'percent':
      return new Intl.NumberFormat('en-US', fractions);
    default: {
      const unknown: never = format;
      return unknown;
    }
  }
}

function percentSign(format: NumberFormat): string {
  return format === 'percent' ? '%' : '';
}

function compact(maximumFractionDigits: number) {
  return new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits });
}

const compactMoney = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  notation: 'compact',
  maximumFractionDigits: 1,
});
