import { screen, waitFor, within } from '@testing-library/react';
import type { HttpHandler } from 'msw';
import { describe, expect, it } from 'vitest';
import type { Permission } from '@/session/access';
import { expectNoAxeViolations } from '@/testing/axe';
import { seedList } from '@/testing/fixtures/plans';
import { deleted } from '@/testing/fixtures/settings';
import {
  BAR_WIDGET,
  customDashboard,
  sharedDashboard,
  TEXT_WIDGET,
  withPermissions,
} from '@/testing/fixtures/shared-dashboards';
import { TEAMS } from '@/testing/fixtures/teams';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, pending, reply } from '@/testing/replies';
import type { Answer } from '@/testing/replies';
import { DashboardViewerPage } from './DashboardViewerPage';

const ALL: Permission[] = ['plans.view', 'dashboards.view', 'dashboards.build'];
const SAVED_FILTERS = {
  dateFrom: '2026-01-01',
  dateTo: '2026-12-31',
  planIds: [],
  teamIds: [],
};

interface Options {
  route?: string;
  permissions?: Permission[];
  answers?: HttpHandler[];
}

function renderViewer(
  dashboard: Answer<'GET /api/v1/custom-dashboards/{dashboardId}'> = { body: customDashboard() },
  { route = '/dashboards/DB-1001', permissions = ALL, answers = [] }: Options = {},
) {
  const api = mockApi([
    reply('GET /api/v1/custom-dashboards/{dashboardId}', dashboard),
    reply('GET /api/v1/plans', { body: seedList() }),
    reply('GET /api/v1/teams', { body: TEAMS }),
    reply('POST /api/v1/data-sources/query', pending),
    ...answers,
  ]);
  const view = renderWithProviders(<DashboardViewerPage />, {
    route,
    path: '/dashboards/:dashboardId',
    user: withPermissions(permissions),
  });
  return { ...view, api };
}

async function heading(name: string) {
  return screen.findByRole('heading', { level: 1, name });
}

describe('051 FR-036, FR-039: the dashboard viewer', () => {
  it('shows the dashboard with its facts and its widgets in grid order (US2 scenario 3)', async () => {
    renderViewer({ body: customDashboard({ widgets: [BAR_WIDGET, TEXT_WIDGET] }) });

    expect(await heading('Lubricants Europe overview')).toBeInTheDocument();
    expect(document.title).toBe('Lubricants Europe overview · SP&A One');
    expect(screen.getByText('By Priya Raman')).toBeInTheDocument();
    expect(screen.getByText('Lubricants Europe')).toBeInTheDocument();
    const cells = within(screen.getByRole('list', { name: 'Widgets' })).getAllByRole('listitem');
    expect(cells).toHaveLength(2);
    expect(cells[0]?.style.getPropertyValue('--column')).toBe('1');
    expect(cells[1]?.style.getPropertyValue('--column')).toBe('5');
    expect(cells[1]?.style.getPropertyValue('--width')).toBe('6');
    expect(cells[1]?.style.getPropertyValue('--height')).toBe('4');
    await expectNoAxeViolations();
  });

  it('offers the owner Edit, opening the builder (US2 scenario 8)', async () => {
    renderViewer();

    expect(await screen.findByRole('link', { name: 'Edit' })).toHaveAttribute(
      'href',
      '/builder/DB-1001',
    );
  });

  it('shows a shared dashboard view-only: no Edit, no Delete (US2 scenario 8, US4 scenario 2)', async () => {
    renderViewer({ body: sharedDashboard() });

    await heading('Lubricants Europe overview');
    expect(screen.queryByRole('link', { name: 'Edit' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'More' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Make a copy' })).toBeInTheDocument();
  });

  it('offers Make a copy only to holders of Build dashboards', async () => {
    renderViewer({ body: sharedDashboard() }, { permissions: ['dashboards.view'] });

    await heading('Lubricants Europe overview');
    expect(screen.queryByRole('button', { name: 'Make a copy' })).not.toBeInTheDocument();
  });

  it('says "Dashboard not found" for a dashboard the person may not see (US2 scenario 9)', async () => {
    renderViewer({ status: 404, body: apiError('not_found', 'Not found') });

    expect(await screen.findByText('Dashboard not found')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Go to Dashboards' })).toHaveAttribute(
      'href',
      '/dashboards',
    );
    await expectNoAxeViolations();
  });

  it('shows the server’s refusal as No access', async () => {
    // A refusal makes the session ask who is signed in again.
    renderViewer(
      { status: 403, body: apiError('forbidden', "You can't see dashboards.") },
      { answers: [reply('GET /api/v1/me', { body: withPermissions(ALL) })] },
    );

    expect(await heading('No access')).toBeInTheDocument();
    expect(screen.getByText("You can't see dashboards.")).toBeInTheDocument();
  });

  it('shows the loading state', () => {
    renderViewer(pending);

    expect(screen.getByLabelText('Loading the dashboard')).toBeInTheDocument();
  });

  it('shows an error with Try again, which loads the dashboard', async () => {
    let calls = 0;
    const { user } = renderViewer(() => {
      calls += 1;
      return calls === 1
        ? { status: 500, body: apiError('internal_error', 'Broken') }
        : { body: customDashboard() };
    });

    await user.click(await screen.findByRole('button', { name: 'Try again' }));

    expect(await heading('Lubricants Europe overview')).toBeInTheDocument();
  });

  it('says so when the dashboard has no widgets', async () => {
    renderViewer({ body: sharedDashboard({ widgets: [] }) });

    expect(await screen.findByText('This dashboard has no widgets yet')).toBeInTheDocument();
    expect(screen.queryByRole('list', { name: 'Widgets' })).not.toBeInTheDocument();
  });

  it('asks every widget for its data again on Refresh (US2 scenario 12)', async () => {
    const { user, queryClient } = renderViewer();
    const key = ['widget-data', 'activities', 'earlier'];
    queryClient.setQueryData(key, { rows: [] });

    await user.click(await screen.findByRole('button', { name: 'Refresh' }));

    await waitFor(() => expect(queryClient.getQueryState(key)?.isInvalidated).toBe(true));
  });

  it('makes a copy and opens it in the builder (US4 scenario 6)', async () => {
    const { user, router } = renderViewer(
      { body: sharedDashboard() },
      {
        answers: [
          reply('POST /api/v1/custom-dashboards/{dashboardId}/copy', {
            status: 201,
            body: customDashboard({ id: 'DB-1009' }),
          }),
        ],
      },
    );

    await user.click(await screen.findByRole('button', { name: 'Make a copy' }));

    await waitFor(() => expect(router.state.location.pathname).toBe('/builder/DB-1009'));
    expect(screen.getByText('Copy made')).toBeInTheDocument();
  });

  it('deletes after the question and goes back to Dashboards with Undo (FR-010)', async () => {
    const { user, router } = renderViewer(undefined, {
      answers: [
        reply('DELETE /api/v1/custom-dashboards/{dashboardId}', {
          body: deleted('dashboard', 'DB-1001', 'Lubricants Europe overview'),
        }),
      ],
    });

    await user.click(await screen.findByRole('button', { name: 'More' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Delete' }));
    const dialog = await screen.findByRole('alertdialog');
    await user.click(within(dialog).getByRole('button', { name: 'Delete dashboard' }));

    await waitFor(() => expect(router.state.location.pathname).toBe('/dashboards'));
    expect(
      await screen.findByText("Dashboard 'Lubricants Europe overview' deleted"),
    ).toBeInTheDocument();
  });
});

describe('051 FR-037: the filters bar (US2 scenario 7)', () => {
  it('starts from the dashboard’s saved filters when the address has none', async () => {
    renderViewer({ body: customDashboard({ filters: SAVED_FILTERS }) });

    expect(await screen.findByLabelText('Start date')).toHaveValue('2026-01-01');
    expect(screen.getByLabelText('End date')).toHaveValue('2026-12-31');
    expect(screen.queryByRole('button', { name: 'Reset filters' })).not.toBeInTheDocument();
  });

  it('takes the filters from the address, and falls back for an unknown value', async () => {
    renderViewer(
      { body: customDashboard({ filters: SAVED_FILTERS }) },
      { route: '/dashboards/DB-1001?from=2026-07-01&to=not-a-date' },
    );

    expect(await screen.findByLabelText('Start date')).toHaveValue('2026-07-01');
    expect(screen.getByLabelText('End date')).toHaveValue('2026-12-31');
  });

  it('keeps a changed filter in the address', async () => {
    const { user, router } = renderViewer({ body: customDashboard({ filters: SAVED_FILTERS }) });

    await user.click(await screen.findByLabelText('Plans'));
    await user.click(await screen.findByRole('option', { name: 'FY2026 Mobility China' }));

    await waitFor(() =>
      expect(new URLSearchParams(router.state.location.search).getAll('plan')).toEqual(['P-1002']),
    );
    expect(router.state.location.search).toContain('from=2026-01-01');
  });

  it('returns to the saved filters with Reset filters', async () => {
    const { user, router } = renderViewer(
      { body: customDashboard({ filters: SAVED_FILTERS }) },
      { route: '/dashboards/DB-1001?from=2026-07-01&to=2026-09-30' },
    );

    await user.click(await screen.findByRole('button', { name: 'Reset filters' }));

    await waitFor(() => expect(router.state.location.search).toBe(''));
    expect(screen.getByLabelText('Start date')).toHaveValue('2026-01-01');
  });
});
