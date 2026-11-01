import type { ActionsMenuItem } from '@/ui/ActionsMenu/ActionsMenu';
import type { ActivityCategory } from './api';

/** What people can do with an activity category, from its row or its details. */
export interface ActivityCategoryActions {
  view: (category: ActivityCategory) => void;
  edit: (category: ActivityCategory) => void;
  setActive: (category: ActivityCategory, active: boolean) => void;
  remove: (category: ActivityCategory) => void;
}

/** Set inactive or Reactivate · Delete (last, after a line — design system §6.1). */
export function moreActions(
  category: ActivityCategory,
  actions: ActivityCategoryActions,
): ActionsMenuItem[] {
  return [
    category.active
      ? { label: 'Set inactive', onSelect: () => actions.setActive(category, false) }
      : { label: 'Reactivate', onSelect: () => actions.setActive(category, true) },
    {
      label: 'Delete',
      tone: 'danger',
      separatorBefore: true,
      onSelect: () => actions.remove(category),
    },
  ];
}

/** "2 activity types", or "None yet" when the category groups nothing. */
export function typeCount(category: ActivityCategory): string {
  const count = category.activityTypes.length;
  if (count === 0) {
    return 'None yet';
  }
  return count === 1 ? '1 activity type' : `${count} activity types`;
}
