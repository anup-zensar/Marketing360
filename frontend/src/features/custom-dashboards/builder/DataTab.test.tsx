import { screen } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { testUser } from '@/testing/fixtures/users';
import type { Widget, WidgetQuery, WidgetType } from '../model';
import { BAR_WIDGET, SOURCES } from './builder-fixtures';
import { DataTab } from './DataTab';
import { newWidget } from './widget-defaults';

let latest: WidgetQuery | null = null;

function Harness({ start }: { start: Widget }) {
  const [query, setQuery] = useState(start.query);
  if (!query) {
    return null;
  }
  return (
    <DataTab
      widget={{ ...start, query }}
      query={query}
      sources={SOURCES}
      sourcesFailed={false}
      errorAt={() => undefined}
      onQuery={(next) => {
        latest = next;
        setQuery(next);
      }}
    />
  );
}

function renderTab(start: Widget) {
  latest = null;
  mockApi([]);
  return renderWithProviders(<Harness start={start} />, { user: testUser() });
}

function empty(type: WidgetType): Widget {
  return newWidget(type, 'w1', { x: 0, y: 0, w: 6, h: 4 });
}

describe('051 US1-3, FR-016–FR-020: the Data tab builds the query from the data catalogue', () => {
  it('chooses a source, a group and Sum of Planned cost', async () => {
    const { user } = renderTab(empty('bar'));
    expect(screen.queryByLabelText(/Group by/)).not.toBeInTheDocument();

    await user.selectOptions(screen.getByLabelText(/Source/), 'Activities');
    await user.selectOptions(screen.getByLabelText(/Group by/), 'Activity type');
    await user.click(screen.getByRole('button', { name: 'Add measure' }));
    await user.selectOptions(screen.getByLabelText('What to measure'), 'Planned cost');

    expect(latest).toMatchObject({
      source: 'activities',
      dimension: { field: 'activityType', dateGroup: null },
      measures: [{ aggregation: 'sum', field: 'plannedCost', label: '' }],
      sort: { by: 'measure', measureIndex: 0, direction: 'desc' },
    });
  });

  it('groups the sources as SP&A One data and Uploaded data, with Upload data', () => {
    renderTab(empty('bar'));
    expect(screen.getByRole('group', { name: 'SP&A One data' })).toBeInTheDocument();
    expect(screen.getByRole('group', { name: 'Uploaded data' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Upload data' })).toHaveAttribute('href', '/datasets');
  });

  it('groups a date field by month and sorts by the dates', async () => {
    const { user } = renderTab(empty('bar'));
    await user.selectOptions(screen.getByLabelText(/Source/), 'Activities');
    await user.selectOptions(screen.getByLabelText(/Group by/), 'Start date');

    expect(screen.getByLabelText('Group dates by (group by)')).toHaveValue('month');
    await user.selectOptions(screen.getByLabelText('Group dates by (group by)'), 'Quarter');
    expect(latest).toMatchObject({
      dimension: { field: 'startDate', dateGroup: 'quarter' },
      sort: { by: 'dimension', direction: 'asc' },
    });
  });

  it('offers only what the type takes: no breakdown for a pie, no group for a key figure', async () => {
    const { user } = renderTab(empty('pie'));
    await user.selectOptions(screen.getByLabelText(/Source/), 'Activities');
    expect(screen.getByLabelText(/Group by/)).toBeInTheDocument();
    expect(screen.queryByLabelText(/Break down by/)).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Add measure' }));
    expect(screen.queryByRole('button', { name: 'Add measure' })).not.toBeInTheDocument();
  });

  it('shows no group for a key figure', async () => {
    const { user } = renderTab(empty('kpi'));
    await user.selectOptions(screen.getByLabelText(/Source/), 'Activities');
    expect(screen.queryByLabelText(/Group by/)).not.toBeInTheDocument();
  });
});

describe('051 FR-031: changing the source keeps what fits and says what was cleared', () => {
  it('clears the group and the measure that Event leads 2026 does not have', async () => {
    const { user } = renderTab(BAR_WIDGET);
    await user.selectOptions(screen.getByLabelText(/Source/), 'Event leads 2026');

    expect(screen.getByRole('status')).toHaveTextContent(
      "Cleared what doesn't fit Event leads 2026: Group by, 1 measure.",
    );
    expect(latest).toMatchObject({ source: 'DS-1001', dimension: null, measures: [] });
  });
});

describe('051 FR-019, FR-020: filters, sort and top N', () => {
  it('adds a filter with is one of and one value per line', async () => {
    const { user } = renderTab(BAR_WIDGET);
    await user.click(screen.getByRole('button', { name: 'Add filter' }));
    await user.selectOptions(screen.getByLabelText('Field'), 'Status');
    await user.selectOptions(screen.getByLabelText('Comparison'), 'is one of');
    await user.type(screen.getByLabelText('Values'), 'Planned{Enter}Done');

    expect(latest?.filters).toEqual([
      { field: 'status', operator: 'in', values: ['Planned', 'Done'] },
    ]);
  });

  it('offers comparisons by the field: between for amounts, contains for text', async () => {
    const { user } = renderTab(BAR_WIDGET);
    await user.click(screen.getByRole('button', { name: 'Add filter' }));
    await user.selectOptions(screen.getByLabelText('Field'), 'Planned cost');
    expect(screen.queryByRole('option', { name: 'contains' })).not.toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText('Comparison'), 'between');
    await user.type(screen.getByLabelText('From'), '100');
    await user.type(screen.getByLabelText('To'), '500');

    expect(latest?.filters).toEqual([
      { field: 'plannedCost', operator: 'between', values: ['100', '500'] },
    ]);
  });

  it('sorts by the group, shows the top 10 and groups the rest as Other', async () => {
    const { user } = renderTab(BAR_WIDGET);
    await user.selectOptions(screen.getByLabelText('Sort by'), 'The group');
    await user.selectOptions(screen.getByLabelText('Order'), 'Ascending');
    await user.clear(screen.getByLabelText('Show top'));
    await user.type(screen.getByLabelText('Show top'), '10');
    await user.click(screen.getByLabelText('Group the rest as Other'));

    expect(latest).toMatchObject({
      sort: { by: 'dimension', measureIndex: 0, direction: 'asc' },
      limit: 10,
      groupRest: true,
    });
  });

  it('has no accessibility violations (SC-006)', async () => {
    renderTab(BAR_WIDGET);
    await expectNoAxeViolations();
  });
});
