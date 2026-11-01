import { screen, waitFor, within } from '@testing-library/react';
import type { HttpHandler } from 'msw';
import { describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { mockApi } from '@/testing/mock-api';
import { findMessages, renderWithProviders } from '@/testing/render';
import { apiError, pending, reply } from '@/testing/replies';
import { NO_PLANS } from '@/testing/fixtures/session';
import { testUser } from '@/testing/fixtures/users';
import { widgetData } from '../charts/widget-fixtures';
import { dashboard, EUROPE_TEAM, PRIYA_REF, SOURCES } from './builder-fixtures';
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

function renderBuilder(route: string, answers: HttpHandler[] = []) {
  const api = mockApi([...CHOICES, ...answers]);
  const rendered = renderWithProviders(<DashboardBuilderPage />, {
    user: testUser(),
    route,
    path: '/builder/:dashboardId?',
  });
  return { api, ...rendered };
}

const DANIEL_REF = { ...PRIYA_REF, id: 'U-1003', name: 'Daniel Okoro', initials: 'DO' };

describe('051 US1-1, FR-026: the empty builder', () => {
  it('shows "New dashboard", the palette, the dashboard details and one yellow Save', async () => {
    renderBuilder('/builder');

    expect(screen.getByRole('heading', { level: 1, name: 'New dashboard' })).toBeInTheDocument();
    const palette = screen.getByRole('region', { name: 'Widgets' });
    expect(within(palette).getByRole('heading', { name: 'Charts' })).toBeInTheDocument();
    expect(
      within(palette).getByRole('heading', { name: 'Figures and tables' }),
    ).toBeInTheDocument();
    expect(within(palette).getByRole('button', { name: 'Add text note' })).toBeInTheDocument();
    const inspector = screen.getByRole('complementary', { name: 'Inspector' });
    expect(within(inspector).getByLabelText(/^Name/)).toBeInTheDocument();
    expect(within(inspector).getByLabelText('Description')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Save' })).toHaveClass(/primary/);
    expect(screen.queryByRole('button', { name: 'Sharing' })).not.toBeInTheDocument();
    expect(document.title).toBe('New dashboard · SP&A One');
    await waitFor(() => expect(screen.getByLabelText('Plans')).toBeInTheDocument());
  });

  it('has no accessibility violations (SC-006)', async () => {
    renderBuilder('/builder');
    await screen.findByLabelText('Teams');
    await expectNoAxeViolations();
  });
});

describe('051 US1-2, US1-5: adding a widget and changing it in the inspector', () => {
  it('selects a clicked widget and opens its Data tab', async () => {
    const { user } = renderBuilder('/builder');
    await user.click(screen.getByRole('button', { name: 'Add bar' }));

    const inspector = screen.getByRole('complementary', { name: 'Inspector' });
    expect(within(inspector).getByRole('heading', { name: 'Bar chart' })).toBeInTheDocument();
    expect(within(inspector).getByRole('tab', { name: 'Data' })).toHaveAttribute(
      'aria-selected',
      'true',
    );
    expect(within(inspector).getByLabelText(/Source/)).toBeInTheDocument();
  });

  it('follows a change at once on the canvas, saving nothing', async () => {
    const { user, api } = renderBuilder('/builder');
    await user.click(screen.getByRole('button', { name: 'Add text note' }));
    await user.type(screen.getByLabelText('Title'), 'Read me');

    expect(screen.getByRole('group', { name: cell('Read me') })).toBeInTheDocument();
    expect(api.requests.filter((request) => request.method !== 'GET')).toEqual([]);
  });

  it('goes back to the dashboard details', async () => {
    const { user } = renderBuilder('/builder');
    await user.click(screen.getByRole('button', { name: 'Add text note' }));
    await user.click(screen.getByRole('button', { name: 'Dashboard details' }));
    expect(screen.getByLabelText(/^Name/)).toBeInTheDocument();
  });
});

describe('051 US1-7, FR-030: Undo and Redo', () => {
  it('reverses and repeats the last change with the buttons and announces it', async () => {
    const { user } = renderBuilder('/builder');
    await user.click(screen.getByRole('button', { name: 'Add bar' }));
    await user.click(screen.getByRole('button', { name: 'Undo' }));

    expect(screen.queryByRole('group', { name: cell('Bar chart') })).not.toBeInTheDocument();
    expect(screen.getByText('Undone: added Bar chart.')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Redo' }));
    expect(screen.getByRole('group', { name: cell('Bar chart') })).toBeInTheDocument();
  });

  it('takes Ctrl+Z, Ctrl+Shift+Z and Ctrl+Y outside text fields', async () => {
    const { user } = renderBuilder('/builder');
    await user.click(screen.getByRole('button', { name: 'Add bar' }));
    await user.keyboard('{Control>}z{/Control}');
    expect(screen.queryByRole('group', { name: cell('Bar chart') })).not.toBeInTheDocument();

    await user.keyboard('{Control>}{Shift>}z{/Shift}{/Control}');
    expect(screen.getByRole('group', { name: cell('Bar chart') })).toBeInTheDocument();

    await user.keyboard('{Control>}z{/Control}{Control>}y{/Control}');
    expect(screen.getByRole('group', { name: cell('Bar chart') })).toBeInTheDocument();
  });

  it('leaves Ctrl+Z to the browser while typing in a field', async () => {
    const { user } = renderBuilder('/builder');
    await user.click(screen.getByRole('button', { name: 'Add bar' }));
    await user.click(screen.getByRole('button', { name: 'Dashboard details' }));
    await user.type(screen.getByLabelText(/^Name/), 'Europe');
    await user.keyboard('{Control>}z{/Control}');

    expect(screen.getByRole('group', { name: cell('Bar chart') })).toBeInTheDocument();
  });
});

describe('051 FR-035, FR-039: the builder page states', () => {
  it('shows loading while the dashboard comes', () => {
    renderBuilder('/builder/DB-1002', [
      reply('GET /api/v1/custom-dashboards/{dashboardId}', () => pending()),
    ]);
    expect(screen.getByText('Loading the dashboard…')).toBeInTheDocument();
  });

  it('opens the owner’s dashboard to edit, named in the title', async () => {
    renderBuilder('/builder/DB-1002', [
      reply('GET /api/v1/custom-dashboards/{dashboardId}', { body: dashboard() }),
    ]);
    expect(
      await screen.findByRole('heading', { level: 1, name: "Edit 'Lubricants Europe overview'" }),
    ).toBeInTheDocument();
    expect(screen.getByRole('group', { name: cell('Planned cost by type') })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Sharing' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'View' })).toHaveAttribute(
      'href',
      '/dashboards/DB-1002',
    );
  });

  it('answers "Dashboard not found" for a dashboard the person may not see', async () => {
    renderBuilder('/builder/DB-1009', [
      reply('GET /api/v1/custom-dashboards/{dashboardId}', {
        status: 404,
        body: apiError('not_found', 'Dashboard not found'),
      }),
    ]);
    expect(
      await screen.findByRole('heading', { level: 1, name: 'Dashboard not found' }),
    ).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Go to Dashboards' })).toHaveAttribute(
      'href',
      '/dashboards',
    );
  });

  it('shows the no-access page when refused', async () => {
    renderBuilder('/builder/DB-1002', [
      reply('GET /api/v1/me', { body: testUser() }),
      reply('GET /api/v1/custom-dashboards/{dashboardId}', {
        status: 403,
        body: apiError('forbidden', "You don't have access to this page."),
      }),
    ]);
    expect(await screen.findByText(/You don't have access to this page/)).toBeInTheDocument();
  });

  it('shows an error with Try again, which asks again', async () => {
    let calls = 0;
    const { user } = renderBuilder('/builder/DB-1002', [
      reply('GET /api/v1/custom-dashboards/{dashboardId}', () => {
        calls += 1;
        return calls === 1
          ? { status: 500, body: apiError('internal_error', 'Something went wrong.') }
          : { body: dashboard() };
      }),
    ]);
    await user.click(await screen.findByRole('button', { name: 'Try again' }));
    expect(
      await screen.findByRole('heading', { name: "Edit 'Lubricants Europe overview'" }),
    ).toBeInTheDocument();
  });

  it('shows another owner’s dashboard view-only, with Make a copy (US4-6)', async () => {
    const copy = dashboard({ id: 'DB-1003', name: 'Copy of Lubricants Europe overview' });
    const { user, router } = renderBuilder('/builder/DB-1002', [
      reply(
        'GET /api/v1/custom-dashboards/{dashboardId}',
        { body: dashboard({ isOwner: false, owner: DANIEL_REF }) },
        { dashboardId: 'DB-1002' },
      ),
      reply(
        'GET /api/v1/custom-dashboards/{dashboardId}',
        { body: copy },
        { dashboardId: 'DB-1003' },
      ),
      reply('POST /api/v1/custom-dashboards/{dashboardId}/copy', { status: 201, body: copy }),
    ]);
    expect(
      await screen.findByText('Only Daniel Okoro can edit this dashboard.'),
    ).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Save' })).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Make a copy' }));
    expect(await screen.findByText('Copy made')).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/builder/DB-1003');
  });

  it('says the builder needs a wider window, with a way to view the dashboard', async () => {
    renderBuilder('/builder/DB-1002', [
      reply('GET /api/v1/custom-dashboards/{dashboardId}', { body: dashboard() }),
    ]);
    expect(await screen.findByText('The builder needs a wider window.')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'View the dashboard' })).toHaveAttribute(
      'href',
      '/dashboards/DB-1002',
    );
  });
});

describe('051 US4-1, FR-046: sharing from the builder', () => {
  it('saves the chosen teams through the sharing dialog and saves the dashboard next with the new version', async () => {
    const { user, api } = renderBuilder('/builder/DB-1002', [
      reply('GET /api/v1/custom-dashboards/{dashboardId}', { body: dashboard() }),
      reply('PUT /api/v1/custom-dashboards/{dashboardId}/sharing', {
        body: dashboard({ sharedWith: [EUROPE_TEAM], version: 4 }),
      }),
      reply('PATCH /api/v1/custom-dashboards/{dashboardId}', {
        body: dashboard({ sharedWith: [EUROPE_TEAM], version: 5 }),
      }),
    ]);
    await user.click(await screen.findByRole('button', { name: 'Sharing' }));
    const dialog = await screen.findByRole('dialog', {
      name: "Share 'Lubricants Europe overview'",
    });
    await user.click(within(dialog).getByLabelText('Lubricants Europe'));
    await user.click(within(dialog).getByRole('button', { name: 'Save' }));

    expect(await findMessages()).toHaveTextContent('Sharing saved');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(api.bodiesOf('PUT /api/v1/custom-dashboards/DB-1002/sharing')).toEqual([
      { version: 3, teamIds: ['T-1001'] },
    ]);

    await user.click(screen.getByRole('button', { name: 'Sharing' }));
    expect(
      within(await screen.findByRole('dialog')).getByLabelText('Lubricants Europe'),
    ).toBeChecked();
    await user.keyboard('{Escape}');
    await user.click(screen.getByRole('button', { name: 'Save' }));
    await waitFor(() =>
      expect(api.bodiesOf('PATCH /api/v1/custom-dashboards/DB-1002')).toEqual([
        expect.objectContaining({ version: 4 }),
      ]),
    );
  });
});
