import { screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { findMessages } from '@/testing/render';
import {
  ASSET,
  PROMOTION,
  renderActivityCategoriesPage,
} from '@/testing/fixtures/activity-categories';
import { validationFailed, versionConflict } from '@/testing/fixtures/settings';
import { chooseAction } from '@/testing/fixtures/settings-pages';
import { reply } from '@/testing/replies';

const DEFAULT_HINT =
  'The reports open on this category. Marking a new default unmarks the category that had it.';

describe('ActivityCategoriesPage: create and edit (US3, FR-018 to FR-020)', () => {
  it('creates a category with a name (FR-018)', async () => {
    const { user, api } = renderActivityCategoriesPage([
      reply('POST /api/v1/settings/activity-categories', {
        status: 201,
        body: { ...PROMOTION, id: 'ACAT-1004', name: 'Sponsorship' },
      }),
    ]);

    await user.click(screen.getByRole('button', { name: 'New category' }));
    const form = await screen.findByRole('dialog', { name: 'New activity category' });
    await user.type(within(form).getByRole('textbox', { name: /^Name/ }), 'Sponsorship');
    await user.click(within(form).getByRole('button', { name: 'Save' }));

    expect(await findMessages()).toHaveTextContent("Activity category 'Sponsorship' created");
    expect(api.bodiesOf('POST /api/v1/settings/activity-categories')).toEqual([
      { name: 'Sponsorship', isDefault: false, activityTypeIds: [] },
    ]);
  });

  it('shows the API message when the name repeats one ignoring case (FR-018)', async () => {
    const { user } = renderActivityCategoriesPage([
      reply('POST /api/v1/settings/activity-categories', {
        status: 400,
        body: validationFailed([
          { field: 'name', message: "An activity category called 'Asset' already exists." },
        ]),
      }),
    ]);

    await user.click(screen.getByRole('button', { name: 'New category' }));
    const form = await screen.findByRole('dialog', { name: 'New activity category' });
    await user.type(within(form).getByRole('textbox', { name: /^Name/ }), 'asset');
    await user.click(within(form).getByRole('button', { name: 'Save' }));

    expect(
      await within(form).findAllByText("An activity category called 'Asset' already exists."),
    ).toHaveLength(2);
  });

  it('shows every problem at once and saves nothing (FR-018)', async () => {
    const { user } = renderActivityCategoriesPage([
      reply('POST /api/v1/settings/activity-categories', {
        status: 400,
        body: validationFailed([
          { field: 'name', message: 'Enter a name of 1 to 80 characters.' },
          { field: 'activityTypeIds', message: "'Event' is not an activity type any more." },
        ]),
      }),
    ]);

    await user.click(screen.getByRole('button', { name: 'New category' }));
    const form = await screen.findByRole('dialog', { name: 'New activity category' });
    await user.click(within(form).getByRole('button', { name: 'Save' }));

    expect(await within(form).findByText('2 problems need your attention')).toBeInTheDocument();
    expect(within(form).getAllByText('Enter a name of 1 to 80 characters.')).toHaveLength(2);
    expect(within(form).getAllByText("'Event' is not an activity type any more.")).not.toHaveLength(
      0,
    );
  });

  it('moves an activity type that is already in another category (FR-019)', async () => {
    const { user, api } = renderActivityCategoriesPage([
      reply(
        'PATCH /api/v1/settings/activity-categories/{categoryId}',
        { body: { ...PROMOTION, activityTypes: ASSET.activityTypes } },
        { categoryId: 'ACAT-1002' },
      ),
    ]);

    await chooseAction('Promotion', 'Edit', user);
    const form = await screen.findByRole('dialog', { name: 'Edit activity category' });
    const types = await within(form).findByRole('group', { name: /^Activity types/ });
    expect(types).toHaveAccessibleDescription(
      expect.stringContaining('choosing it moves it to this category'),
    );
    // Event is in Asset, and is offered here all the same.
    await user.click(within(types).getByRole('checkbox', { name: 'Event' }));
    await user.click(within(form).getByRole('button', { name: 'Save' }));

    expect(await findMessages()).toHaveTextContent("Activity category 'Promotion' saved");
    expect(api.bodiesOf('PATCH /api/v1/settings/activity-categories/ACAT-1002')).toEqual([
      { activityTypeIds: ['AT-1002'], version: 1 },
    ]);
  });

  it('marks a category the default and says the previous one is unmarked (FR-020)', async () => {
    const { user, api } = renderActivityCategoriesPage([
      reply(
        'PATCH /api/v1/settings/activity-categories/{categoryId}',
        { body: { ...PROMOTION, isDefault: true } },
        { categoryId: 'ACAT-1002' },
      ),
    ]);

    await chooseAction('Promotion', 'Edit', user);
    const form = await screen.findByRole('dialog', { name: 'Edit activity category' });
    const isDefault = within(form).getByRole('checkbox', { name: 'Default category' });
    expect(isDefault).not.toBeChecked();
    expect(isDefault).toHaveAccessibleDescription(expect.stringContaining(DEFAULT_HINT));
    await user.click(isDefault);
    await user.click(within(form).getByRole('button', { name: 'Save' }));

    expect(await findMessages()).toHaveTextContent("Activity category 'Promotion' saved");
    expect(api.bodiesOf('PATCH /api/v1/settings/activity-categories/ACAT-1002')).toEqual([
      { isDefault: true, version: 1 },
    ]);
  });

  it('keeps the typed values after a conflict and saves again over it (FR-018)', async () => {
    let patches = 0;
    const { user, api } = renderActivityCategoriesPage([
      reply(
        'PATCH /api/v1/settings/activity-categories/{categoryId}',
        () => {
          patches += 1;
          return patches === 1
            ? { status: 409, body: versionConflict() }
            : { body: { ...PROMOTION, name: 'Promotions' } };
        },
        { categoryId: 'ACAT-1002' },
      ),
      reply(
        'GET /api/v1/settings/activity-categories/{categoryId}',
        { body: { ...PROMOTION, version: 4 } },
        { categoryId: 'ACAT-1002' },
      ),
    ]);

    await chooseAction('Promotion', 'Edit', user);
    const form = await screen.findByRole('dialog', { name: 'Edit activity category' });
    const name = within(form).getByRole('textbox', { name: /^Name/ });
    await user.clear(name);
    await user.type(name, 'Promotions');
    await user.click(within(form).getByRole('button', { name: 'Save' }));

    expect(await within(form).findByRole('alert')).toHaveTextContent(
      'Not saved — Daniel Okoro changed this activity category on 14 Mar 2026 at 10:42',
    );
    expect(name).toHaveValue('Promotions');
    await waitFor(() =>
      expect(
        api.requests.some(
          (request) =>
            request.method === 'GET' &&
            request.path === '/api/v1/settings/activity-categories/ACAT-1002',
        ),
      ).toBe(true),
    );
    await user.click(within(form).getByRole('button', { name: 'Save' }));

    expect(await findMessages()).toHaveTextContent("Activity category 'Promotions' saved");
    expect(api.bodiesOf('PATCH /api/v1/settings/activity-categories/ACAT-1002')).toEqual([
      { name: 'Promotions', version: 1 },
      { name: 'Promotions', version: 4 },
    ]);
  });
});
