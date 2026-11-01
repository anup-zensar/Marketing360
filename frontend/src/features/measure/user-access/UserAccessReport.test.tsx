import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { TEAMS } from '@/testing/fixtures/plans';
import { reportFrame } from '@/testing/fixtures/reports';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import type { Answer } from '@/testing/replies';
import { userAccessReport } from './user-access-fixture';
import { UserAccessReport } from './UserAccessReport';

const PATH = '/api/v1/reports/user-access';

function renderReport(
  answer: Answer<'GET /api/v1/reports/user-access'> = { body: userAccessReport() },
  route = '/reports/user-access',
  forbidden = false,
) {
  const api = mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply(
      'GET /api/v1/reports/frame',
      forbidden
        ? { status: 403, body: apiError('forbidden', "Changing access isn't part of your role.") }
        : { body: reportFrame({ report: 'userAccess' }) },
    ),
    reply('GET /api/v1/reports/{report}/views', { body: { items: [], total: 0 } }),
    reply('GET /api/v1/reports/user-access', answer),
    reply('GET /api/v1/teams', { body: TEAMS }),
  ]);
  const rendered = renderWithProviders(<UserAccessReport />, {
    route,
    user: testUser(),
    path: '/reports/user-access',
  });
  return { api, ...rendered };
}

const lastSearch = (api: ReturnType<typeof mockApi>) =>
  api.requests.filter((r) => r.path === PATH).at(-1)?.search ?? '';

describe('UserAccessReport (045 US1, FR-001, FR-005 – FR-008, FR-012)', () => {
  it('lists everyone with roles, teams and last sign-in, a name opening Settings › People', async () => {
    renderReport();

    const table = await screen.findByRole('table', { name: 'People' });
    expect(within(table).getByRole('link', { name: 'Priya Raman' })).toHaveAttribute(
      'href',
      '/settings/people/U-1002',
    );
    expect(within(table).getByText('Planner (My teams)')).toBeInTheDocument();
    expect(within(table).getByText('Archive (inactive), Lubricants Europe')).toBeInTheDocument();
    expect(within(table).getByText('No role')).toBeInTheDocument();
    expect(within(table).getByText('Never')).toBeInTheDocument();
    expect(screen.getByText(/^Sign-ins recorded since/)).toBeInTheDocument();
    expect(screen.queryByRole('combobox', { name: 'Currency' })).not.toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('keeps the search, status and sort in the address (FR-006)', async () => {
    const { router, user, api } = renderReport();
    const table = await screen.findByRole('table', { name: 'People' });

    await user.click(screen.getByRole('button', { name: 'All' }));
    expect(router.state.location.search).toBe('?status=all');
    await user.click(within(table).getByRole('button', { name: 'Last sign-in' }));
    expect(router.state.location.search).toBe('?status=all&sort=-lastSignedIn');
    expect(lastSearch(api)).toContain('sort=-lastSignedIn');
  });

  it('says no one matches a search, with Clear search (FR-012)', async () => {
    const { router, user } = renderReport(
      { body: userAccessReport({ total: 0, rows: [] }) },
      '/reports/user-access?q=zz',
    );

    expect(await screen.findByText('No people match "zz"')).toBeInTheDocument();
    const clear = screen.getAllByRole('button', { name: 'Clear search' }).at(-1);
    if (!clear) {
      throw new Error('No Clear search button.');
    }
    await user.click(clear);
    expect(router.state.location.search).toBe('');
  });

  it('says no one matches the filters, with Reset filters (FR-012)', async () => {
    const { router, user } = renderReport(
      { body: userAccessReport({ total: 0, rows: [] }) },
      '/reports/user-access?status=inactive&team=T-1002',
    );

    expect(await screen.findByText('No people match these filters')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Reset filters' }));
    expect(router.state.location.search).toBe('');
  });

  it('shows the standard no-access page without access.manage (FR-002)', async () => {
    renderReport({ body: userAccessReport() }, '/reports/user-access', true);

    expect(await screen.findByRole('heading', { name: /access/i })).toBeInTheDocument();
    expect(screen.queryByRole('table', { name: 'People' })).not.toBeInTheDocument();
  });
});
