import { screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { components } from '@/api/generated/schema';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { findMessages, renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import type { Answer } from '@/testing/replies';
import { SavedViewsMenu } from './SavedViewsMenu';

type ReportView = components['schemas']['ReportView'];

const VIEWS = '/api/v1/reports/plan-budget/views';

function view(overrides: Partial<ReportView>): ReportView {
  return {
    id: 'RV-1001',
    report: 'planBudget',
    name: 'Europe',
    isDefault: false,
    query: 'owner=U-1003',
    version: 1,
    createdAt: '2026-10-30T09:00:00Z',
    updatedAt: '2026-10-30T09:00:00Z',
    ...overrides,
  };
}

const EUROPE = view({ isDefault: true });
const ASIA = view({ id: 'RV-1002', name: 'Asia', query: 'owner=U-1002' });

function renderMenu(
  route: string,
  patch: Answer<'PATCH /api/v1/reports/{report}/views/{viewId}'> = { body: view({ version: 2 }) },
) {
  const views = [EUROPE, ASIA];
  const api = mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/reports/{report}/views', { body: { items: views, total: views.length } }),
    reply('POST /api/v1/reports/{report}/views', async (request) => ({
      status: 201,
      body: view({ id: 'RV-1003', ...(await request.json()) }),
    })),
    reply('PATCH /api/v1/reports/{report}/views/{viewId}', patch),
    reply('DELETE /api/v1/reports/{report}/views/{viewId}', { body: ASIA }),
  ]);
  const rendered = renderWithProviders(<SavedViewsMenu report="planBudget" />, {
    route,
    user: testUser(),
    path: '/reports/plan-budget',
  });
  return { ...rendered, api };
}

const search = (router: { state: { location: { search: string } } }) =>
  router.state.location.search;

describe('SavedViewsMenu (039 US2, FR-024 – FR-029)', () => {
  it('lists the views, ticks the one on screen and tags the default (FR-026)', async () => {
    const { user } = renderMenu('/reports/plan-budget?owner=U-1003');

    await user.click(await screen.findByRole('button', { name: 'View: Europe' }));

    const items = await screen.findAllByRole('menuitemcheckbox');
    expect(items.map((item) => item.textContent)).toEqual(['EuropeDefault', 'Asia']);
    expect(items[0]).toHaveAttribute('aria-checked', 'true');
    expect(screen.getAllByRole('menuitem').map((item) => item.textContent)).toEqual([
      'Save current filters…',
      'Manage views…',
    ]);
    await expectNoAxeViolations();
  });

  it('applies a view as an address, so Back returns (FR-026)', async () => {
    const { user, router } = renderMenu('/reports/plan-budget?owner=U-1003');

    await user.click(await screen.findByRole('button', { name: 'View: Europe' }));
    await user.click(await screen.findByRole('menuitemcheckbox', { name: 'Asia' }));

    await waitFor(() => expect(search(router)).toBe('?owner=U-1002'));
    expect(await screen.findByRole('button', { name: 'View: Asia' })).toHaveFocus();
  });

  it('opens on the default view when the address is empty (FR-027)', async () => {
    const { router } = renderMenu('/reports/plan-budget');

    await waitFor(() => expect(search(router)).toBe('?owner=U-1003'));
  });

  it('says Modified and updates the view only after asking (FR-026)', async () => {
    const { user, router, api } = renderMenu('/reports/plan-budget?owner=U-1003');
    await screen.findByRole('button', { name: 'View: Europe' });

    await router.navigate('/reports/plan-budget?owner=U-1003&planType=strategic');
    await user.click(await screen.findByRole('button', { name: 'View: Europe Modified' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Update view…' }));
    await user.click(await screen.findByRole('button', { name: 'Update view' }));

    await waitFor(() =>
      expect(api.bodiesOf(`PATCH ${VIEWS}/RV-1001`)).toEqual([
        { version: 1, query: 'owner=U-1003&planType=strategic' },
      ]),
    );
  });

  it('saves the filters under a new name, and asks before replacing a name in use (FR-024)', async () => {
    const { user, api } = renderMenu('/reports/plan-budget?owner=U-1004');

    await user.click(await screen.findByRole('button', { name: 'Saved views' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Save current filters…' }));
    await user.type(await screen.findByRole('textbox', { name: /Name/ }), 'europe');
    await user.click(screen.getByRole('button', { name: 'Save view' }));

    expect(await screen.findByText("Replace your view 'Europe'?")).toBeInTheDocument();
    expect(api.bodiesOf(`POST ${VIEWS}`)).toEqual([]);
    await user.click(screen.getByRole('button', { name: 'Replace view' }));
    await waitFor(() =>
      expect(api.bodiesOf(`PATCH ${VIEWS}/RV-1001`)).toEqual([
        { version: 1, query: 'owner=U-1004' },
      ]),
    );
  });

  it('deletes at once and Undo saves the view again with its star (FR-028)', async () => {
    const { user, api } = renderMenu('/reports/plan-budget');

    await user.click(await screen.findByRole('button', { name: /View: Europe|Saved views/ }));
    await user.click(await screen.findByRole('menuitem', { name: 'Manage views…' }));
    const panel = await screen.findByRole('dialog', { name: 'Your saved views' });
    await user.click(within(panel).getByRole('button', { name: 'Delete Asia' }));

    expect(await screen.findByText("View 'Asia' deleted")).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Undo', hidden: true }));
    await waitFor(() =>
      expect(api.bodiesOf(`POST ${VIEWS}`)).toEqual([
        { name: 'Asia', query: 'owner=U-1002', isDefault: false },
      ]),
    );
  });

  it('says a view was changed in another tab instead of overwriting it (FR-028)', async () => {
    const { user, router } = renderMenu('/reports/plan-budget?owner=U-1003', {
      status: 409,
      body: apiError(
        'version_conflict',
        'This view was changed in another tab — reload it to see the latest',
      ),
    });
    await screen.findByRole('button', { name: 'View: Europe' });

    await router.navigate('/reports/plan-budget?owner=U-1003&team=T-1001');
    await user.click(await screen.findByRole('button', { name: 'View: Europe Modified' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Update view…' }));
    await user.click(await screen.findByRole('button', { name: 'Update view' }));

    const messages = await findMessages();
    expect(
      await within(messages).findByText(
        'This view was changed in another tab — reload it to see the latest',
      ),
    ).toBeInTheDocument();
  });
});
