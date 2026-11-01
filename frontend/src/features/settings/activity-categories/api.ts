/**
 * What the Activity categories page asks the API (feature 047, US3). The list comes from
 * useActivityCategories() (api/settingsQueries.ts); every change refreshes the lists that show
 * categories — activity types carry the category they are in — and the activity log, recycle bin
 * and Settings counts.
 */
import { useMutation, useQueryClient } from '@tanstack/react-query';
import type { QueryKey } from '@tanstack/react-query';
import { api, call } from '@/api/client';
import type { components } from '@/api/generated/schema';
import { SETTINGS_KEYS } from '../api';

export type ActivityCategory = components['schemas']['ActivityCategory'];
export type ActivityCategoryCreate = components['schemas']['ActivityCategoryCreate'];
export type ActivityCategoryUpdate = components['schemas']['ActivityCategoryUpdate'];

/** The lists a change to an activity category shows in. */
export const ACTIVITY_CATEGORY_LISTS: QueryKey[] = [
  SETTINGS_KEYS.activityCategories,
  SETTINGS_KEYS.activityTypes,
  SETTINGS_KEYS.overview,
  ['activity-log'],
  ['recycle-bin'],
];

/** One activity category as it is stored now; after a conflict the form takes its version from it. */
export function fetchActivityCategory(id: string): Promise<ActivityCategory> {
  return call(
    api.GET('/api/v1/settings/activity-categories/{categoryId}', {
      params: { path: { categoryId: id } },
    }),
  );
}

function patchCategory(id: string, body: ActivityCategoryUpdate): Promise<ActivityCategory> {
  return call(
    api.PATCH('/api/v1/settings/activity-categories/{categoryId}', {
      params: { path: { categoryId: id } },
      body,
    }),
  );
}

/** Sets an activity category inactive or reactivates it (FR-018): the record's own PATCH. */
export function setActivityCategoryStatus(id: string, body: { active: boolean; version: number }) {
  return patchCategory(id, body);
}

export function useCreateActivityCategory() {
  const refresh = useRefreshActivityCategoryLists();
  return useMutation({
    mutationFn: (body: ActivityCategoryCreate) =>
      call(api.POST('/api/v1/settings/activity-categories', { body })),
    onSettled: refresh,
  });
}

/** Changes a category: only the fields sent, with the version the form started from (D-203). */
export function useUpdateActivityCategory() {
  const refresh = useRefreshActivityCategoryLists();
  return useMutation({
    mutationFn: ({ id, body }: { id: string; body: ActivityCategoryUpdate }) =>
      patchCategory(id, body),
    onSettled: refresh,
  });
}

/** Deletes a category (to the recycle bin); its activity types become Uncategorised (FR-021). */
export function useDeleteActivityCategory() {
  const refresh = useRefreshActivityCategoryLists();
  return useMutation({
    mutationFn: (id: string) =>
      call(
        api.DELETE('/api/v1/settings/activity-categories/{categoryId}', {
          params: { path: { categoryId: id } },
        }),
      ),
    onSettled: refresh,
  });
}

function useRefreshActivityCategoryLists() {
  const queryClient = useQueryClient();
  return () =>
    Promise.all(
      ACTIVITY_CATEGORY_LISTS.map((queryKey) => queryClient.invalidateQueries({ queryKey })),
    );
}
