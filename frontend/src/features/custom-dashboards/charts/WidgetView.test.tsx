import { screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { apiError, pending, reply } from '@/testing/replies';
import { renderWithProviders } from '@/testing/render';
import type { Widget, WidgetData } from '../model';
import { NO_FILTERS } from '../model';
import { WidgetView } from './WidgetView';
import type { WidgetViewProps } from './WidgetView';
import { column, key, money, query, row, widget, widgetData } from './widget-fixtures';

const echarts = vi.hoisted(() => ({
  chart: {
    setOption: () => {},
    resize: () => {},
    dispose: () => {},
    getDataURL: () => 'data:image/svg+xml,',
    getWidth: () => 400,
    getHeight: () => 240,
  },
}));

vi.mock('echarts/core', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  init: () => echarts.chart,
}));

const QUERY_PATH = 'POST /api/v1/data-sources/query';
const CHART_NAME = /^Bar chart: Planned cost by type\./;

function show(shown: Widget, props: Partial<WidgetViewProps> = {}) {
  return renderWithProviders(
    <WidgetView
      widget={shown}
      filters={NO_FILTERS}
      mode="view"
      dashboardName="Market view"
      viewerOwnsDashboard={false}
      {...props}
    />,
    { user: testUser() },
  );
}

function answer(data: WidgetData) {
  return mockApi([reply(QUERY_PATH, { body: data })]);
}

describe('051 FR-039: every state of a widget', () => {
  it('reserves its space while loading', () => {
    mockApi([reply(QUERY_PATH, () => pending())]);
    show(widget('bar'));
    expect(screen.getByRole('status')).toHaveTextContent('Loading…');
    expect(screen.getByRole('status').parentElement).toHaveAttribute('aria-busy', 'true');
  });

  it('shows the error with Retry, and the chart once it loads', async () => {
    let calls = 0;
    mockApi([
      reply(QUERY_PATH, () => {
        calls += 1;
        return calls === 1
          ? { status: 400, body: apiError('validation_failed', 'Choose a field to measure.') }
          : { body: widgetData() };
      }),
    ]);
    const { user } = show(widget('bar'));

    expect(await screen.findByText("This widget couldn't be loaded")).toBeInTheDocument();
    await expectNoAxeViolations();
    await user.click(screen.getByRole('button', { name: 'Retry' }));
    expect(await screen.findByRole('img', { name: CHART_NAME })).toBeInTheDocument();
  });

  it('says there is no data to show and what to change', async () => {
    answer(widgetData({ rows: [], groupCount: 0 }));
    show(widget('bar'));
    expect(await screen.findByText('No data to show')).toBeInTheDocument();
    expect(screen.getByText(/Try a wider date range/)).toBeInTheDocument();
  });

  it.each([
    ['a source the viewer may not read', 403, 'forbidden'],
    ['a dataset the viewer may not see', 404, 'not_found'],
  ] as const)('says the viewer has no access to %s', async (_name, status, code) => {
    mockApi([
      reply(QUERY_PATH, { status, body: apiError(code, 'No') }),
      reply('GET /api/v1/me', { body: testUser() }),
    ]);
    show(widget('bar'));
    expect(await screen.findByText("You don't have access to this data")).toBeInTheDocument();
  });

  it('tells the dashboard’s owner that a deleted dataset is no longer available', async () => {
    mockApi([reply(QUERY_PATH, { status: 404, body: apiError('not_found', 'Gone') })]);
    show(widget('bar'), { viewerOwnsDashboard: true });
    expect(await screen.findByText('This data is no longer available')).toBeInTheDocument();
  });

  it('asks what to measure in the builder instead of asking SP&A One', () => {
    const api = mockApi([]);
    show(widget('bar', {}, { query: query({ dimension: null }) }), { mode: 'edit' });
    expect(screen.getByText('Choose what to measure')).toBeInTheDocument();
    expect(api.requests).toEqual([]);
  });

  it('draws the chart with its text summary and the widget’s notes', async () => {
    answer(
      widgetData({
        rows: [row(key('Event'), [money('3.00')])],
        groupCount: 37,
        notes: ['Not filtered by date'],
      }),
    );
    show(widget('bar', {}, { subtitle: 'This year' }));

    expect(await screen.findByRole('img', { name: CHART_NAME })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Planned cost by type' })).toBeInTheDocument();
    expect(screen.getByText('This year')).toBeInTheDocument();
    expect(screen.getByText('Showing 1 of 37 groups')).toBeInTheDocument();
    expect(screen.getByText('Not filtered by date')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('shows a key figure with its comparison in words', async () => {
    answer(
      widgetData({
        columns: [column('m0', 'Sum of Planned cost')],
        rows: [],
        comparison: {
          kind: 'previousPeriod',
          value: money('20000.00'),
          change: '-12.5',
          period: { from: '2025-10-01', to: '2025-12-31' },
        },
      }),
    );
    const display = {
      kpi: { comparison: 'previousPeriod', target: null, higherIsBetter: false },
    } as const;
    show(widget('kpi', display, { query: query({ dimension: null }) }), {
      filters: { ...NO_FILTERS, dateFrom: '2026-01-01', dateTo: '2026-03-31' },
    });

    expect(await screen.findByText('$25,000.00')).toBeInTheDocument();
    expect(screen.getByText(/^Down 12.5%, better than the previous period \(/)).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('shows a key figure against its target', async () => {
    answer(
      widgetData({
        columns: [column('m0', 'Sum of Planned cost')],
        comparison: { kind: 'target', value: money('40000.00'), change: '62.5', period: null },
      }),
    );
    const display = {
      kpi: { comparison: 'target', target: '40000', higherIsBetter: true },
    } as const;
    show(widget('kpi', display, { query: query({ dimension: null }) }));
    expect(await screen.findByText('62.5% of the target of $40,000.00')).toBeInTheDocument();
  });

  it('shows a text note as plain text, never as HTML, without asking for data', () => {
    const api = mockApi([]);
    show(widget('text', {}, { title: '', note: 'Read me <b>first</b>\nthen this' }));
    expect(screen.getByText(/Read me <b>first<\/b>/)).toBeInTheDocument();
    expect(document.querySelector('b')).toBeNull();
    expect(api.requests).toEqual([]);
  });

  it('shows a table widget as a table', async () => {
    answer(widgetData());
    show(widget('table'));
    expect(await screen.findByRole('table', { name: 'Planned cost by type' })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Actions for/ })).not.toBeInTheDocument();
  });
});

describe('051 FR-038: the widget menu', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('switches between the chart and a real table of the same rows', async () => {
    answer(widgetData());
    const { user } = show(widget('bar'));
    await screen.findByRole('img', { name: CHART_NAME });

    await user.click(screen.getByRole('button', { name: 'Actions for Planned cost by type' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Show as table' }));
    expect(screen.getByRole('table', { name: 'Planned cost by type' })).toBeInTheDocument();
    expect(screen.queryByRole('img', { name: CHART_NAME })).not.toBeInTheDocument();
    await expectNoAxeViolations();

    await user.click(screen.getByRole('button', { name: 'Actions for Planned cost by type' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Show as chart' }));
    expect(await screen.findByRole('img', { name: CHART_NAME })).toBeInTheDocument();
  });

  it('says so when the chart cannot be saved as a picture', async () => {
    answer(widgetData());
    const { user } = show(widget('bar'));
    await screen.findByRole('img', { name: CHART_NAME });

    await user.click(screen.getByRole('button', { name: 'Actions for Planned cost by type' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Download as PNG' }));
    await waitFor(() =>
      expect(
        screen.getByText("The chart couldn't be saved as a picture. Please try again."),
      ).toBeInTheDocument(),
    );
  });

  it('offers no menu in the builder', async () => {
    answer(widgetData());
    show(widget('bar'), { mode: 'edit' });
    await screen.findByRole('img', { name: CHART_NAME });
    expect(screen.queryByRole('button', { name: /Actions for/ })).not.toBeInTheDocument();
  });
});

describe('051 SC-006: a builder preview fits its cell', () => {
  it('draws the chart smaller under the builder’s handle row, and never scrolls', async () => {
    answer(widgetData());
    show(widget('bar'), { mode: 'edit' });
    const chart = await screen.findByRole('img', { name: CHART_NAME });

    expect(chart).toHaveStyle({ height: '232px' });
    expect(chart.closest('[aria-busy]')).toHaveClass(/preview/);
  });

  it('keeps the full height on a dashboard', async () => {
    answer(widgetData());
    show(widget('bar'));
    expect(await screen.findByRole('img', { name: CHART_NAME })).toHaveStyle({ height: '296px' });
  });
});
