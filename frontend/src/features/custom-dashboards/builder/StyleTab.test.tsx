import { screen, within } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { testUser } from '@/testing/fixtures/users';
import type { Widget, WidgetType } from '../model';
import { BAR_QUERY, BAR_WIDGET } from './builder-fixtures';
import { StyleTab } from './StyleTab';

let latest: Widget = BAR_WIDGET;

function Harness({ start }: { start: Widget }) {
  const [widget, setWidget] = useState(start);
  const keep = (next: Widget) => {
    latest = next;
    setWidget(next);
  };
  return (
    <StyleTab
      widget={widget}
      errorAt={() => undefined}
      onDisplay={(display) => keep({ ...widget, display })}
      onText={(patch) => keep({ ...widget, ...patch })}
    />
  );
}

function renderTab(type: WidgetType, overrides: Partial<Widget> = {}) {
  latest = { ...BAR_WIDGET, type, ...overrides };
  mockApi([]);
  return renderWithProviders(<Harness start={latest} />, { user: testUser() });
}

const TWO_MEASURES = {
  ...BAR_QUERY,
  measures: [
    { aggregation: 'sum' as const, field: 'plannedCost', label: 'Planned' },
    { aggregation: 'count' as const, field: null, label: '' },
  ],
};

describe('051 US1, FR-025: the Style tab offers the options the type draws', () => {
  it('gives a bar chart horizontal bars, data labels and money format', async () => {
    const { user } = renderTab('bar');
    await user.selectOptions(screen.getByLabelText('Orientation'), 'Horizontal');
    await user.click(screen.getByLabelText('Data labels'));
    await user.selectOptions(screen.getByLabelText('Number format'), 'Money');
    await user.selectOptions(screen.getByLabelText('Decimals'), '0');

    expect(latest.display).toEqual({
      orientation: 'horizontal',
      dataLabels: true,
      numberFormat: 'money',
      decimals: 0,
    });
  });

  it('sets the title and the subtitle', async () => {
    const { user } = renderTab('bar', { title: '' });
    await user.type(screen.getByLabelText('Title'), 'Spend');
    await user.type(screen.getByLabelText('Subtitle'), 'By type');
    expect(latest).toMatchObject({ title: 'Spend', subtitle: 'By type' });
  });

  it('chooses a palette of Shell chart colours, never a free colour (FR-024)', async () => {
    const { user } = renderTab('bar');
    expect(screen.getByRole('radio', { name: 'Categorical' })).toBeChecked();
    await user.click(screen.getByRole('radio', { name: 'Planning' }));
    expect(latest.display.palette).toBe('planning');
  });

  it('starts a heatmap on the Sequential palette and a pie with data labels', () => {
    renderTab('heatmap');
    expect(screen.getByRole('radio', { name: 'Sequential' })).toBeChecked();
  });

  it('shows no axes for a pie, and its data labels on', () => {
    renderTab('pie');
    expect(screen.queryByRole('group', { name: 'Axis titles' })).not.toBeInTheDocument();
    expect(screen.getByLabelText('Data labels')).toBeChecked();
  });

  it('draws a reference line at a value with its label (US5-1)', async () => {
    const { user } = renderTab('line');
    await user.type(screen.getByLabelText('Value'), '25000');
    await user.type(screen.getByLabelText('Label'), 'Target');
    expect(latest.display.referenceLine).toEqual({ value: '25000', label: 'Target' });
  });

  it('sets bars or line and the axis of each measure of a bar and line chart', async () => {
    const { user } = renderTab('combo', { query: TWO_MEASURES });
    expect(screen.getByRole('group', { name: 'Planned' })).toBeInTheDocument();
    const second = within(screen.getByRole('group', { name: 'Measure 2' }));
    await user.selectOptions(second.getByLabelText('Bars or line'), 'Bars');
    expect(latest.display.measureStyles).toEqual([
      { as: 'bar', axis: 'left' },
      { as: 'bar', axis: 'right' },
    ]);
    expect(screen.getByLabelText('Right')).toBeInTheDocument();
  });

  it('compares a key figure with a target, higher being better (US5-2)', async () => {
    const { user } = renderTab('kpi');
    await user.selectOptions(screen.getByLabelText('Compare with'), 'Target');
    await user.type(screen.getByLabelText(/^Target/), '1000000');
    await user.click(screen.getByLabelText('Higher is better'));

    expect(latest.display.kpi).toEqual({
      comparison: 'target',
      target: '1000000',
      higherIsBetter: false,
    });
  });

  it('says a previous-period comparison needs a date range', async () => {
    const { user } = renderTab('kpi');
    await user.selectOptions(screen.getByLabelText('Compare with'), 'Previous period');
    expect(screen.getByText("Needs a date range in the dashboard's filters.")).toBeInTheDocument();
  });

  it('sets a gauge minimum and maximum', async () => {
    const { user } = renderTab('gauge');
    await user.clear(screen.getByLabelText('Minimum'));
    await user.type(screen.getByLabelText('Minimum'), '10');
    await user.type(screen.getByLabelText(/^Maximum/), '90');
    expect(latest.display.gauge).toEqual({ min: '10', max: '90' });
  });

  it('chooses the table columns and a totals row', async () => {
    const { user } = renderTab('table', { query: TWO_MEASURES });
    await user.click(screen.getByLabelText('Measure 2'));
    await user.click(screen.getByLabelText('Totals row'));
    expect(latest.display.table).toEqual({ columns: ['dimension', 'm0'], totals: true });
  });

  it('writes a text note as plain text, with no chart options', async () => {
    const { user } = renderTab('text', { query: null, note: '' });
    expect(screen.queryByRole('group', { name: 'Colours' })).not.toBeInTheDocument();
    await user.type(screen.getByLabelText('Note'), '<b>Not bold</b>');
    expect(latest.note).toBe('<b>Not bold</b>');
  });

  it('has no accessibility violations (SC-006)', async () => {
    renderTab('combo', { query: TWO_MEASURES });
    await expectNoAxeViolations();
  });
});
