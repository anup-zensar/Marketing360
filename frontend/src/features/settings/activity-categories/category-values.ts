import type { ActivityCategory, ActivityCategoryCreate, ActivityCategoryUpdate } from './api';
import type { ActivityCategoryValues } from './schemas';

export function startValues(category: ActivityCategory | undefined): ActivityCategoryValues {
  return {
    name: category?.name ?? '',
    isDefault: category?.isDefault ?? false,
    activityTypeIds: category?.activityTypes.map((type) => type.id) ?? [],
  };
}

export function createBody(values: ActivityCategoryValues): ActivityCategoryCreate {
  return {
    name: values.name,
    isDefault: values.isDefault,
    activityTypeIds: values.activityTypeIds,
  };
}

/** The fields people changed, with the version (D-203); null when nothing changed. */
export function changedFields(
  start: ActivityCategoryValues,
  values: ActivityCategoryValues,
  version: number,
): ActivityCategoryUpdate | null {
  const body: ActivityCategoryUpdate = { version };
  if (values.name !== start.name) {
    body.name = values.name;
  }
  if (values.isDefault !== start.isDefault) {
    body.isDefault = values.isDefault;
  }
  if (!sameIds(values.activityTypeIds, start.activityTypeIds)) {
    // The complete new list of activity types in the category (FR-019).
    body.activityTypeIds = values.activityTypeIds;
  }
  return Object.keys(body).length > 1 ? body : null;
}

function sameIds(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((id) => b.includes(id));
}
