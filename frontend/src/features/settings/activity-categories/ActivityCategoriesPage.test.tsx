import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { findMessages, renderWithProviders } from '@/testing/render';
import { TYPES } from '@/testing/fixtures/activity-types';
import {
  ASSET,
  CATEGORIES,
  CATEGORIES_PATH,
  LEGACY,
  renderActivityCategoriesPage,
} from '@/testing/fixtures/activity-categories';
import { deleted, inUse } from '@/testing/fixtures/settings';
import { chooseAction, rowOf } from '@/testing/fixtures/settings-pages';
import { VIEWER } from '@/testing/fixtures/viewer';
import { mockApi } from '@/testing/mock-api';
import { reply } from '@/testing/replies';
import { ActivityCategoriesPage } from './ActivityCategoriesPage';

function table() {
  return screen.getByRole('table');
}

describe('ActivityCategoriesPage: the list (FR-018)', () => {
  it('shows the active categories with their type count, default mark and status', async () => {
    renderActivityCategoriesPage();

    expect(screen.getByRole('navigation', { name: 'Breadcrumb' })).toHaveTextContent('Settings');
    expect(
      screen.getByRole('heading', { level: 1, name: 'Activity categories' }),
    ).toBeInTheDocument();
    expect(document.title).toBe('Activity categories · SP&A One');

    await screen.findByRole('table');
    const asset = rowOf('Asset');
    expect(asset).toHaveTextContent('1 activity type');
    expect(within(asset).getByText('Default')).toBeInTheDocument();
    expect(within(asset).getByText('Active')).toBeInTheDocument();
    expect(rowOf('Promotion')).toHaveTextContent('None yet');
    expect(within(table()).queryByText('Legacy')).not.toBeInTheDocument();
    expect(screen.getByText('2 of 3 activity categories')).toBeInTheDocument();
  });

  it('filters by status, keeping the choice in the address (FR-018)', async () => {
    const { user, router } = renderActivityCategoriesPage();
    await screen.findByRole('table');

    await user.click(screen.getByRole('button', { name: 'Inactive 1' }));
    expect(within(rowOf('Legacy')).getByText('Inactive')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'All 3' }));

    expect(within(table()).getAllByRole('rowheader')).toHaveLength(3);
    expect(router.state.location.search).toBe('?status=all');
  });

  it('invites to create the first category when there are none (FR-018)', async () => {
    renderActivityCategoriesPage([
      reply('GET /api/v1/settings/activity-categories', { body: { items: [], total: 0 } }),
    ]);

    expect(
      await screen.findByText('No activity categories yet — create your first category'),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'New category' })).toHaveLength(2);
  });

  it('says when the categories cannot be loaded, and tries again (FR-018)', async () => {
    let calls = 0;
    const { user } = renderActivityCategoriesPage([
      reply('GET /api/v1/settings/activity-categories', () => {
        calls += 1;
        return calls === 1
          ? {
              status: 500,
              body: {
                error: {
                  code: 'internal_error',
                  message: 'Something went wrong. Please try again.',
                  fields: [],
                },
              },
            }
          : { body: CATEGORIES };
      }),
    ]);

    expect(await screen.findByRole('alert')).toHaveTextContent('Something went wrong.');
    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('table')).toBeInTheDocument();
  });

  it('shows the page while the categories are still on their way (FR-018)', () => {
    renderActivityCategoriesPage();

    expect(
      screen.getByRole('heading', { level: 1, name: 'Activity categories' }),
    ).toBeInTheDocument();
    expect(screen.queryByRole('table')).not.toBeInTheDocument();
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
  });

  it('shows the details of a category, with its types and who changed it (FR-018)', async () => {
    const { user, router } = renderActivityCategoriesPage();

    await user.click(await screen.findByRole('button', { name: 'Asset' }));

    const panel = await screen.findByRole('dialog', { name: 'Asset' });
    expect(router.state.location.search).toBe('?id=ACAT-1001');
    expect(panel).toHaveTextContent('The reports open on this category.');
    expect(within(panel).getByText('Event')).toBeInTheDocument();
    expect(panel).toHaveTextContent('Used by 4 activities · Groups 1 activity type');
    expect(panel).toHaveTextContent(
      'Created by Priya Raman on 12 Mar 2026, last changed by Daniel Okoro on 14 Mar 2026',
    );
  });
});

describe('ActivityCategoriesPage: delete (FR-021)', () => {
  it('moves an unused category to the recycle bin and offers Undo (FR-021)', async () => {
    const { user } = renderActivityCategoriesPage([
      reply(
        'DELETE /api/v1/settings/activity-categories/{categoryId}',
        { body: deleted('activity category', 'ACAT-1002', 'Promotion') },
        { categoryId: 'ACAT-1002' },
      ),
    ]);

    await screen.findByRole('table');
    await chooseAction('Promotion', 'Delete', user);
    const dialog = await screen.findByRole('alertdialog', {
      name: "Delete activity category 'Promotion'?",
    });
    expect(dialog).toHaveTextContent('Its activity types become Uncategorised.');
    await user.click(within(dialog).getByRole('button', { name: 'Delete activity category' }));

    const messages = await findMessages();
    expect(messages).toHaveTextContent("Activity category 'Promotion' deleted");
    expect(within(messages).getByRole('button', { name: 'Undo' })).toBeInTheDocument();
  });

  it('warns about a category in use and offers Set inactive instead (FR-021)', async () => {
    const { user, api } = renderActivityCategoriesPage([
      reply(
        'PATCH /api/v1/settings/activity-categories/{categoryId}',
        { body: { ...ASSET, active: false } },
        { categoryId: 'ACAT-1001' },
      ),
    ]);

    await screen.findByRole('table');
    await chooseAction('Asset', 'Delete', user);

    const dialog = await screen.findByRole('alertdialog', { name: "'Asset' is in use" });
    expect(dialog).toHaveTextContent(
      'Used by 4 activities · Groups 1 activity type — set it inactive instead.',
    );
    await user.click(within(dialog).getByRole('button', { name: 'Set inactive' }));

    expect(await findMessages()).toHaveTextContent("Activity category 'Asset' set inactive");
    expect(api.bodiesOf('PATCH /api/v1/settings/activity-categories/ACAT-1001')).toEqual([
      { active: false, version: 3 },
    ]);
    expect(api.requests.some((request) => request.method === 'DELETE')).toBe(false);
  });

  it('switches to the in-use dialog when the API finds the category in use (FR-021)', async () => {
    const message = "'Promotion' is in use (used by 1 activity) — set it inactive instead.";
    const { user } = renderActivityCategoriesPage([
      reply(
        'DELETE /api/v1/settings/activity-categories/{categoryId}',
        { status: 409, body: inUse(message) },
        { categoryId: 'ACAT-1002' },
      ),
    ]);

    await screen.findByRole('table');
    await chooseAction('Promotion', 'Delete', user);
    await user.click(await screen.findByRole('button', { name: 'Delete activity category' }));

    expect(await screen.findByRole('alertdialog', { name: "'Promotion' is in use" })).toHaveTextContent(
      message,
    );
  });

  it('reactivates an inactive category (FR-018)', async () => {
    const { user, api } = renderActivityCategoriesPage(
      [
        reply(
          'PATCH /api/v1/settings/activity-categories/{categoryId}',
          { body: { ...LEGACY, active: true } },
          { categoryId: 'ACAT-1003' },
        ),
      ],
      `${CATEGORIES_PATH}?status=inactive`,
    );

    await screen.findByRole('table');
    await chooseAction('Legacy', 'Reactivate', user);

    expect(await findMessages()).toHaveTextContent("Activity category 'Legacy' reactivated");
    expect(api.bodiesOf('PATCH /api/v1/settings/activity-categories/ACAT-1003')).toEqual([
      { active: true, version: 1 },
    ]);
  });
});

describe('ActivityCategoriesPage: access and accessibility (FR-023)', () => {
  it('a viewer without settings.manage reads the page with the view-only note (FR-023)', async () => {
    mockApi([
      reply('GET /api/v1/settings/activity-categories', { body: CATEGORIES }),
      reply('GET /api/v1/activity-types', { body: TYPES }),
    ]);
    const { user } = renderWithProviders(<ActivityCategoriesPage />, {
      user: VIEWER,
      route: CATEGORIES_PATH,
    });

    await screen.findByRole('table');
    expect(screen.getByRole('note')).toHaveTextContent(
      'View only — only administrators change Settings. Ask Alex Morgan if something needs to change.',
    );
    expect(screen.queryByRole('button', { name: 'New category' })).not.toBeInTheDocument();
    await user.click(await screen.findByRole('button', { name: 'Actions for Asset' }));
    expect(await screen.findByRole('menuitem', { name: 'View' })).toBeInTheDocument();
    expect(screen.queryByRole('menuitem', { name: 'Edit' })).not.toBeInTheDocument();
    expect(screen.queryByRole('menuitem', { name: 'Delete' })).not.toBeInTheDocument();
  });

  it('has no axe violations (FR-018)', async () => {
    renderActivityCategoriesPage();

    await screen.findByRole('table');
    await expectNoAxeViolations();
  });
});
