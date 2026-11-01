import { screen, waitFor, within } from '@testing-library/react';
import type { HttpHandler } from 'msw';
import { describe, expect, it } from 'vitest';
import type { Permission } from '@/session/access';
import { expectNoAxeViolations } from '@/testing/axe';
import { deleted } from '@/testing/fixtures/settings';
import {
  customDashboard,
  DANIEL_PERSON,
  dashboardSummary,
  VIEWER_ACCESS,
  withPermissions,
} from '@/testing/fixtures/shared-dashboards';
import { mockApi } from '@/testing/mock-api';
import { findMessages, renderWithProviders } from '@/testing/render';
import { apiError, pending, reply } from '@/testing/replies';
import { DashboardListPage } from './DashboardListPage';

const ALL: Permission[] = ['plans.view', 'dashboards.view', 'dashboards.build'];
const VIEW_ONLY: Permission[] = ['plans.view', 'dashboards.view'];
const SHARED = dashboardSummary({
  id: 'DB-1002',
  name: 'China leads',
  description: '',
  owner: DANIEL_PERSON,
  sharedWith: [],
  widgetCount: 1,
  access: VIEWER_ACCESS,
});
const LIST = { body: { mine: [dashboardSummary()], sharedWithMe: [SHARED] } };

function renderPage(permissions: Permission[] = ALL, answers: HttpHandler[] = []) {
  const api = mockApi([reply('GET /api/v1/custom-dashboards', LIST), ...answers]);
  const view = renderWithProviders(<DashboardListPage />, {
    route: '/dashboards',
    path: '/dashboards',
    user: withPermissions(permissions),
  });
  return { ...view, api };
}

function section(name: string) {
  return screen.getByRole('region', { name });
}

async function choose(user: ReturnType<typeof renderPage>['user'], card: string, item: string) {
  await user.click(await screen.findByRole('button', { name: `Actions for ${card}` }));
  await user.click(await screen.findByRole('menuitem', { name: item }));
}

describe('051 FR-003: the Dashboards list', () => {
  it('shows Built-in, My dashboards and Shared with me with each card’s facts (US2 scenario 1)', async () => {
    renderPage();

    expect(await screen.findByRole('link', { name: 'Lubricants Europe overview' })).toHaveAttribute(
      'href',
      '/dashboards/DB-1001',
    );
    expect(screen.getByRole('heading', { level: 1, name: 'Dashboards' })).toBeInTheDocument();
    expect(document.title).toBe('Dashboards · SP&A One');
    const mine = section('My dashboards');
    expect(mine).toHaveTextContent('Planned cost and leads for the European market.');
    expect(mine).toHaveTextContent(/By Priya Raman · Updated 28 Oct 2026.* · 6 widgets/);
    expect(within(mine).getByText('Lubricants Europe')).toBeInTheDocument();
    const shared = section('Shared with me');
    expect(within(shared).getByRole('link', { name: 'China leads' })).toBeInTheDocument();
    expect(shared).toHaveTextContent('By Daniel Okoro');
    expect(shared).toHaveTextContent('1 widget');
    expect(shared).toHaveTextContent('Not shared');
    await expectNoAxeViolations();
  });

  it('opens 036’s pages from the built-in cards (US2 scenario 2, FR-004)', async () => {
    renderPage();

    const builtIn = section('Built-in');
    expect(within(builtIn).getByRole('link', { name: 'Executive dashboard' })).toHaveAttribute(
      'href',
      '/dashboards/executive',
    );
    expect(within(builtIn).getByRole('link', { name: 'Regional dashboard' })).toHaveAttribute(
      'href',
      '/dashboards/regional',
    );
    expect(builtIn).toHaveTextContent('Budget, planned cost and spend of every market you may see');
    await screen.findByRole('link', { name: 'China leads' });
  });

  it('leaves out Built-in without View plans (US2 scenario 11)', async () => {
    renderPage(['dashboards.view']);

    await screen.findByRole('link', { name: 'China leads' });
    expect(screen.queryByRole('region', { name: 'Built-in' })).not.toBeInTheDocument();
  });

  it('shows only Built-in without See dashboards, and asks for nothing (US2 scenario 11)', () => {
    const { api } = renderPage(['plans.view']);

    expect(section('Built-in')).toBeInTheDocument();
    expect(screen.queryByRole('region', { name: 'My dashboards' })).not.toBeInTheDocument();
    expect(api.requests).toEqual([]);
  });

  it('says No access to someone with neither permission', () => {
    renderPage(['finance.view']);

    expect(screen.getByRole('heading', { level: 1, name: 'No access' })).toBeInTheDocument();
  });

  it('offers New dashboard and Datasets only to builders', async () => {
    renderPage();

    expect(screen.getByRole('link', { name: 'New dashboard' })).toHaveAttribute('href', '/builder');
    expect(screen.getByRole('link', { name: 'Datasets' })).toHaveAttribute('href', '/datasets');
    await screen.findByRole('link', { name: 'China leads' });
  });

  it('hides New dashboard and Datasets from viewers', async () => {
    renderPage(['plans.view', 'dashboards.view']);

    await screen.findByRole('link', { name: 'China leads' });
    expect(screen.queryByRole('link', { name: 'New dashboard' })).not.toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Datasets' })).not.toBeInTheDocument();
  });

  it('shows skeleton cards while loading', () => {
    renderPage(ALL, [reply('GET /api/v1/custom-dashboards', pending)]);

    expect(screen.getByLabelText('Loading dashboards')).toBeInTheDocument();
  });

  it('shows the error with Try again, which loads the list', async () => {
    let calls = 0;
    const { user } = renderPage(ALL, [
      reply('GET /api/v1/custom-dashboards', () => {
        calls += 1;
        return calls === 1 ? { status: 500, body: apiError('internal_error', 'Broken') } : LIST;
      }),
    ]);

    expect(await screen.findByRole('alert')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('link', { name: 'China leads' })).toBeInTheDocument();
  });

  it('says each section is empty, with New dashboard for builders', async () => {
    renderPage(ALL, [
      reply('GET /api/v1/custom-dashboards', { body: { mine: [], sharedWithMe: [] } }),
    ]);

    expect(await screen.findByText("You haven't created any dashboards yet")).toBeInTheDocument();
    const mine = section('My dashboards');
    expect(mine).toHaveTextContent('Would you like to create one?');
    expect(within(mine).getByRole('list', { name: 'Ideas to start with' })).toHaveTextContent(
      'Planned cost by month',
    );
    expect(within(mine).getByRole('link', { name: 'Create your first dashboard' })).toHaveAttribute(
      'href',
      '/builder',
    );
    expect(within(mine).getByRole('link', { name: 'Upload your own data' })).toHaveAttribute(
      'href',
      '/datasets',
    );
    expect(section('Shared with me')).toHaveTextContent('Nothing shared with you yet.');
    await expectNoAxeViolations();
  });

  it('invites only builders to create one', async () => {
    renderPage(VIEW_ONLY, [
      reply('GET /api/v1/custom-dashboards', { body: { mine: [], sharedWithMe: [] } }),
    ]);

    expect(await screen.findByText('You have no dashboards of your own')).toBeInTheDocument();
    expect(
      within(section('My dashboards')).queryByRole('link', { name: 'Create your first dashboard' }),
    ).not.toBeInTheDocument();
  });
});

describe('051 FR-006, FR-011: the card menu', () => {
  it('offers the owner Open, Edit, Make a copy and Delete', async () => {
    const { user } = renderPage();

    await user.click(
      await screen.findByRole('button', { name: 'Actions for Lubricants Europe overview' }),
    );

    const items = await screen.findAllByRole('menuitem');
    expect(items.map((item) => item.textContent)).toEqual([
      'Open',
      'Edit',
      'Make a copy',
      'Delete',
    ]);
  });

  it('offers a shared dashboard view-only: no Edit, no Delete (US4 scenario 2)', async () => {
    const { user } = renderPage();

    await user.click(await screen.findByRole('button', { name: 'Actions for China leads' }));

    const items = await screen.findAllByRole('menuitem');
    expect(items.map((item) => item.textContent)).toEqual(['Open', 'Make a copy']);
  });

  it('opens the dashboard in the builder from Edit', async () => {
    const { user, router } = renderPage();

    await choose(user, 'Lubricants Europe overview', 'Edit');

    await waitFor(() => expect(router.state.location.pathname).toBe('/builder/DB-1001'));
  });

  it('makes a copy and opens it in the builder (US4 scenario 6)', async () => {
    const { user, router, api } = renderPage(ALL, [
      reply(
        'POST /api/v1/custom-dashboards/{dashboardId}/copy',
        { status: 201, body: customDashboard({ id: 'DB-1003', name: 'Copy of China leads' }) },
        { dashboardId: 'DB-1002' },
      ),
    ]);

    await choose(user, 'China leads', 'Make a copy');

    expect(await screen.findByText('Copy made')).toBeInTheDocument();
    await waitFor(() => expect(router.state.location.pathname).toBe('/builder/DB-1003'));
    expect(api.bodiesOf('POST /api/v1/custom-dashboards/DB-1002/copy')).toEqual([{}]);
  });

  it('deletes after a question naming the teams, with Undo (FR-010)', async () => {
    const { user, api } = renderPage(ALL, [
      reply(
        'DELETE /api/v1/custom-dashboards/{dashboardId}',
        { body: deleted('dashboard', 'DB-1001', 'Lubricants Europe overview') },
        { dashboardId: 'DB-1001' },
      ),
      reply('POST /api/v1/recycle-bin/{entryId}/undo', {
        body: {
          recordType: 'customDashboard',
          recordTypeLabel: 'dashboard',
          recordId: 'DB-1001',
          recordName: 'Lubricants Europe overview',
          details: '',
        },
      }),
    ]);

    await choose(user, 'Lubricants Europe overview', 'Delete');
    const dialog = await screen.findByRole('alertdialog', {
      name: "Delete dashboard 'Lubricants Europe overview'?",
    });
    expect(dialog).toHaveTextContent(
      "It's shared with Lubricants Europe; they will no longer see it. It moves to the recycle bin; you can restore it.",
    );
    await user.click(within(dialog).getByRole('button', { name: 'Delete dashboard' }));

    expect(
      await screen.findByText("Dashboard 'Lubricants Europe overview' deleted"),
    ).toBeInTheDocument();
    const messages = await findMessages();
    await user.click(within(messages).getByRole('button', { name: 'Undo' }));

    expect(
      await screen.findByText("Dashboard 'Lubricants Europe overview' restored"),
    ).toBeInTheDocument();
    expect(api.requests.map((r) => `${r.method} ${r.path}`)).toContain(
      'POST /api/v1/recycle-bin/RB-1001/undo',
    );
  });
});
