import { SettingsCount } from '@/ui/SettingsCount/SettingsCount';
import { useActivityCategories } from '../api';

/** "2 active" on the Activity categories card of the Settings page (FR-018). */
export function ActivityCategoryCount() {
  const categories = useActivityCategories();
  return (
    <SettingsCount
      loading={categories.isPending}
      count={categories.data?.items.filter((category) => category.active).length}
    />
  );
}
