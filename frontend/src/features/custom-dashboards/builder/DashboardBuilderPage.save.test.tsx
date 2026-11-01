import { screen, waitFor, within } from '@testing-library/react';
import type { HttpHandler } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { mockApi } from '@/testing/mock-api';
import { findMessages, renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import { NO_PLANS } from '@/testing/fixtures/session';
import { testUser } from '@/testing/fixtures/users';
import { widgetData } from '../charts/widget-fixtures';
import { dashboard, NOTE_WIDGET, SOURCES } from './builder-fixtures';
import { DashboardBuilderPage } from './DashboardBuilderPage';

vi.mock('echarts/core', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  init: () => ({ setOption: () => {}, resize: () => {}, dispose: () => {} }),
}));

const CHOICES: HttpHandler[] = [
  ...NO_PLANS,
  reply('GET /api/v1/teams', { body: { items: [], total: 0 } }),
  reply('GET /api/v1/data-sources', { body: { items: SOURCES } }),
  reply('POST /api/v1/data-sources/query', { body: widgetData() }),
];

function cell(name: string): RegExp {
  return new RegExp(`^Widget: ${name}, column \\d+, row \\d+$`);
}

const SAVED = dashboard();

function renderBuilder(route: string, answers: HttpHandler[] = []) {
  const api = mockApi([
    ...CHOICES,
    reply('GET /api/v1/custom-dashboards/{dashboardId}', { body: SAVED }),
    ...answers,
  ]);
  const rendered = renderWithProviders(<DashboardBuilderPage />, {
    user: testUser(),
    route,
    path: '/builder/:dashboardId?',
  });
  return { api, ...rendered };
}

async function rename(user: ReturnType<typeof renderBuilder>['user'], name: string) {
  const field = await screen.findByLabelText(/^Name/);
  await user.clear(field);
  await user.type(field, name);
}

describe('051 US1-8, FR-033: Save checks everything at once and saves nothing while problems remain', () => {
  it('lists every problem in a summary and on the widget, and focuses the summary', async () => {
    const { user, api } = renderBuilder('/builder');
    await user.click(screen.getByRole('button', { name: 'Add bar' }));
    await user.click(screen.getByRole('button', { name: 'Save' }));

    const summary = screen.getByRole('alert');
    expect(within(summary).getByText('2 problems need your attention')).toBeInTheDocument();
    expect(within(summary).getByRole('link', { name: 'Enter a name.' })).toBeInTheDocument();
    expect(
      within(summary).getByRole('link', { name: 'Bar chart: Choose a source.' }),
    ).toBeInTheDocument();
    expect(summary.parentElement).toHaveFocus();
    const widget = screen.getByRole('group', { name: cell('Bar chart') });
    expect(within(widget).getByText('Choose a source.')).toBeInTheDocument();
    expect(api.requests.filter((request) => request.method === 'POST')).toEqual([]);
  });

  it('takes the person to the problem from the summary', async () => {
    const { user } = renderBuilder('/builder');
    await user.click(screen.getByRole('button', { name: 'Add bar' }));
    await user.click(screen.getByRole('button', { name: 'Save' }));

    await user.click(screen.getByRole('link', { name: 'Enter a name.' }));
    expect(screen.getByLabelText(/^Name/)).toHaveFocus();
    expect(screen.getByLabelText(/^Name/)).toHaveAccessibleDescription(/Enter a name\./);

    await user.click(screen.getByRole('link', { name: 'Bar chart: Choose a source.' }));
    expect(screen.getByRole('button', { name: 'Bar chart', pressed: true })).toHaveFocus();
    expect(screen.getByLabelText(/Source/)).toHaveAccessibleDescription(/Choose a source\./);
  });
});

describe('051 US1-9, FR-012: saving the dashboard', () => {
  it('creates it, says so and moves to its address', async () => {
    const created = dashboard({ widgets: [{ ...NOTE_WIDGET, id: 'w1', title: '', note: '' }] });
    const { user, api, router } = renderBuilder('/builder', [
      reply('POST /api/v1/custom-dashboards', { status: 201, body: created }),
    ]);
    await rename(user, '  Lubricants Europe overview ');
    await user.click(screen.getByRole('button', { name: 'Add text note' }));
    await user.click(screen.getByRole('button', { name: 'Save' }));

    expect(await findMessages()).toHaveTextContent('Dashboard saved');
    expect(api.bodiesOf('POST /api/v1/custom-dashboards')).toEqual([
      {
        name: 'Lubricants Europe overview',
        description: '',
        filters: { dateFrom: null, dateTo: null, planIds: [], teamIds: [] },
        widgets: [
          {
            id: 'w1',
            type: 'text',
            title: '',
            subtitle: '',
            layout: { x: 0, y: 0, w: 4, h: 2 },
            query: null,
            display: {},
            note: '',
          },
        ],
      },
    ]);
    await waitFor(() => expect(router.state.location.pathname).toBe('/builder/DB-1002'));
    expect(
      await screen.findByRole('heading', { name: "Edit 'Lubricants Europe overview'" }),
    ).toBeInTheDocument();
  });

  it('changes it with the version it started from and stays in the builder', async () => {
    const { user, api, router } = renderBuilder('/builder/DB-1002', [
      reply('PATCH /api/v1/custom-dashboards/{dashboardId}', {
        body: dashboard({ name: 'Europe', version: 4 }),
      }),
    ]);
    await rename(user, 'Europe');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    expect(await findMessages()).toHaveTextContent('Dashboard saved');
    expect(api.bodiesOf('PATCH /api/v1/custom-dashboards/DB-1002')).toEqual([
      expect.objectContaining({ version: 3, name: 'Europe', widgets: SAVED.widgets }),
    ]);
    expect(router.state.location.pathname).toBe('/builder/DB-1002');
  });

  it('shows the problems the API finds at their widget', async () => {
    const { user } = renderBuilder('/builder/DB-1002', [
      reply('PATCH /api/v1/custom-dashboards/{dashboardId}', {
        status: 400,
        body: apiError('validation_failed', 'Check the dashboard.', {
          fields: [{ field: 'widgets[0].query.measures', message: 'Sum needs a number field.' }],
        }),
      }),
    ]);
    await rename(user, 'Europe');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    expect(
      await screen.findByRole('link', { name: 'Planned cost by type: Sum needs a number field.' }),
    ).toBeInTheDocument();
  });

  it('says when the name is already used', async () => {
    const { user } = renderBuilder('/builder/DB-1002', [
      reply('PATCH /api/v1/custom-dashboards/{dashboardId}', {
        status: 409,
        body: apiError('name_taken', "You already have a dashboard named 'Europe'."),
      }),
    ]);
    await rename(user, 'Europe');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    expect(
      await screen.findByRole('link', { name: "You already have a dashboard named 'Europe'." }),
    ).toBeInTheDocument();
  });

  it('keeps the changes when saving fails for another reason', async () => {
    const { user } = renderBuilder('/builder/DB-1002', [
      reply('PATCH /api/v1/custom-dashboards/{dashboardId}', {
        status: 500,
        body: apiError('internal_error', 'Something went wrong. Please try again.'),
      }),
    ]);
    await rename(user, 'Europe');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    expect(await screen.findByText('Something went wrong. Please try again.')).toBeInTheDocument();
    expect(screen.getByLabelText(/^Name/)).toHaveValue('Europe');
  });
});

describe('051 US1-11, FR-012: a dashboard saved elsewhere meanwhile', () => {
  const conflict = reply('PATCH /api/v1/custom-dashboards/{dashboardId}', {
    status: 409,
    body: apiError('version_conflict', 'This dashboard was changed by someone else.'),
  });

  it('overwrites nothing, keeps the changes on Keep editing and saves the newer version next', async () => {
    let gets = 0;
    const { user, api } = renderBuilder('/builder/DB-1002', [
      reply('GET /api/v1/custom-dashboards/{dashboardId}', () => {
        gets += 1;
        return { body: gets === 1 ? SAVED : dashboard({ name: 'Newer', version: 5 }) };
      }),
      conflict,
    ]);
    await rename(user, 'Europe');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    const dialog = await screen.findByRole('alertdialog', { name: 'Dashboard changed elsewhere' });
    expect(dialog).toHaveTextContent(
      'This dashboard was changed elsewhere since you opened it. Your changes are still here.',
    );
    await user.click(within(dialog).getByRole('button', { name: 'Keep editing' }));
    expect(screen.getByLabelText(/^Name/)).toHaveValue('Europe');

    await waitFor(() => expect(gets).toBe(2));
    await user.click(screen.getByRole('button', { name: 'Save' }));
    await screen.findByRole('alertdialog');
    expect(api.bodiesOf('PATCH /api/v1/custom-dashboards/DB-1002')).toEqual([
      expect.objectContaining({ version: 3 }),
      expect.objectContaining({ version: 5 }),
    ]);
  });

  it('shows the newer version on Reload', async () => {
    let gets = 0;
    const { user } = renderBuilder('/builder/DB-1002', [
      reply('GET /api/v1/custom-dashboards/{dashboardId}', () => {
        gets += 1;
        return { body: gets === 1 ? SAVED : dashboard({ name: 'Newer', version: 5 }) };
      }),
      conflict,
    ]);
    await rename(user, 'Europe');
    await user.click(screen.getByRole('button', { name: 'Save' }));
    await user.click(await screen.findByRole('button', { name: 'Reload' }));

    await waitFor(() => expect(screen.getByLabelText(/^Name/)).toHaveValue('Newer'));
  });
});

describe('051 US1-10, FR-034: leaving with unsaved changes', () => {
  it('asks "Discard changes?" with Keep editing focused', async () => {
    const { user, router } = renderBuilder('/builder/DB-1002');
    await rename(user, 'Europe');
    await user.click(screen.getByRole('link', { name: 'View' }));

    const dialog = await screen.findByRole('alertdialog', { name: 'Discard changes?' });
    expect(within(dialog).getByRole('button', { name: 'Keep editing' })).toHaveFocus();
    await user.click(within(dialog).getByRole('button', { name: 'Keep editing' }));
    expect(router.state.location.pathname).toBe('/builder/DB-1002');

    await user.click(screen.getByRole('link', { name: 'View' }));
    await user.click(await screen.findByRole('button', { name: 'Discard changes' }));
    expect(router.state.location.pathname).toBe('/dashboards/DB-1002');
  });

  it('leaves without asking when nothing changed', async () => {
    const { user, router } = renderBuilder('/builder/DB-1002');
    await user.click(await screen.findByRole('link', { name: 'View' }));
    expect(router.state.location.pathname).toBe('/dashboards/DB-1002');
  });
});
