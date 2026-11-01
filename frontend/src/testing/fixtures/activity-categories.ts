import { createElement } from 'react';
import type { HttpHandler } from 'msw';
import type { components } from '@/api/generated/schema';
import { ActivityCategoriesPage } from '@/features/settings/activity-categories/ActivityCategoriesPage';
import { mockApi } from '../mock-api';
import { renderWithProviders } from '../render';
import { reply } from '../replies';
import { EVENT, PODCAST, PRINT, TYPES } from './activity-types';
import { MADE, NOT_USED } from './settings';
import { testUser } from './users';

type ActivityCategory = components['schemas']['ActivityCategory'];

export const CATEGORIES_PATH = '/settings/activity-categories';

function ref(type: { id: string; name: string; active: boolean }) {
  return { id: type.id, name: type.name, active: type.active };
}

/** Asset: the default category, groups Event, in use. */
export const ASSET: ActivityCategory = {
  id: 'ACAT-1001',
  name: 'Asset',
  isDefault: true,
  activityTypes: [ref(EVENT)],
  active: true,
  usage: {
    inUse: true,
    summary: 'Used by 4 activities · Groups 1 activity type',
    parts: [
      { kind: 'activity', text: 'Used by 4 activities', count: 4, inRecycleBin: 0 },
      { kind: 'activityType', text: 'Groups 1 activity type', count: 1, inRecycleBin: 0 },
    ],
    records: [],
  },
  version: 3,
  ...MADE,
};

/** Promotion: not the default, groups nothing yet, unused. */
export const PROMOTION: ActivityCategory = {
  ...ASSET,
  id: 'ACAT-1002',
  name: 'Promotion',
  isDefault: false,
  activityTypes: [],
  usage: NOT_USED,
  version: 1,
};

/** Legacy: inactive and unused, holding the inactive Print advert. */
export const LEGACY: ActivityCategory = {
  ...PROMOTION,
  id: 'ACAT-1003',
  name: 'Legacy',
  activityTypes: [ref(PRINT)],
  active: false,
};

export const CATEGORIES = { items: [ASSET, PROMOTION, LEGACY], total: 3 };

export const UNCATEGORISED_TYPE = PODCAST;

/** The Activity categories page for a signed-in person, with the categories and types answered. */
export function renderActivityCategoriesPage(
  answers: HttpHandler[] = [],
  route: string = CATEGORIES_PATH,
) {
  const api = mockApi([
    reply('GET /api/v1/settings/activity-categories', { body: CATEGORIES }),
    reply('GET /api/v1/activity-types', { body: TYPES }),
    ...answers,
  ]);
  return {
    ...renderWithProviders(createElement(ActivityCategoriesPage), {
      route,
      user: testUser(),
      path: CATEGORIES_PATH,
    }),
    api,
  };
}
