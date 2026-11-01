import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DateFormat } from '@/format/intl';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { renderWithProviders } from '@/testing/render';
import type { Widget, WidgetData } from '../model';
import { column, key, money, num, row, seriesData, widget, widgetData } from './widget-fixtures';
import { WidgetTable } from './WidgetTable';

function renderTable(shown: Widget, data: WidgetData, dateFormat: DateFormat = 'DD MMM YYYY') {
  const user = testUser();
  renderWithProviders(<WidgetTable widget={shown} data={data} />, {
    user: { ...user, preferences: { ...user.preferences, dateFormat } },
  });
  return screen.getByRole('table', { name: shown.title });
}

function cells(table: HTMLElement): string[][] {
  return within(table)
    .getAllByRole('row')
    .map((tableRow) => [...tableRow.querySelectorAll('th, td')].map((cell) => cell.textContent));
}

describe('051 FR-038: Show as table', () => {
  it('lists the same rows as the chart with headers and formatted values', async () => {
    const table = renderTable(widget('bar'), seriesData());

    expect(cells(table)).toEqual([
      ['Activity type', 'Plan', 'Sum of Planned cost'],
      ['Event', 'Mobility', '$10,000.00'],
      ['Event', 'Lubricants', '$2,500.00'],
      ['Digital', 'Mobility', '$7,500.00'],
    ]);
    expect(screen.getByRole('region', { name: 'Planned cost by type, table' })).toHaveAttribute(
      'tabindex',
      '0',
    );
    await expectNoAxeViolations();
  });

  it('writes date groups in the person’s date format', () => {
    const data = widgetData({
      columns: [column('dimension', 'Month', 'date', 'month'), column('m0', 'Count', 'number')],
      rows: [row(key('2026-03-01'), [num('4')])],
    });
    expect(cells(renderTable(widget('line'), data, 'YYYY-MM-DD'))[1]).toEqual(['2026-03', '4']);
  });
});

describe('051 FR-025: table widget columns and totals row', () => {
  it('shows only the chosen columns, in their order, and the totals of every group', async () => {
    const shown = widget('table', { table: { columns: ['m1', 'dimension'], totals: true } });
    const data = widgetData({
      columns: [
        column('dimension', 'Activity type', 'text'),
        column('m0', 'Sum of Planned cost'),
        column('m1', 'Count of records', 'number'),
      ],
      rows: [row(key('Event'), [money('1.00'), num('2')])],
      totals: [money('9.00'), num('12')],
    });
    const table = renderTable(shown, data);

    expect(cells(table)).toEqual([
      ['Count of records', 'Activity type'],
      ['2', 'Event'],
      ['12', ''],
    ]);
    await expectNoAxeViolations();
  });

  it('labels the totals row when the group comes first', () => {
    const shown = widget('table', { table: { columns: [], totals: true } });
    expect(cells(renderTable(shown, widgetData())).at(-1)).toEqual(['Total', '$25,000.00']);
  });
});
