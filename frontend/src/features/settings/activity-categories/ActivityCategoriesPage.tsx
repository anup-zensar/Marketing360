import { Plus, Tags } from 'lucide-react';
import { useState } from 'react';
import { ActionsMenu } from '@/ui/ActionsMenu/ActionsMenu';
import { Button } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { StatusChip } from '@/ui/StatusChip/StatusChip';
import { usePageTitle } from '@/ui/usePageTitle';
import { ListToolbar } from '@/ui/ListToolbar/ListToolbar';
import styles from '@/ui/settings-table.module.css';
import { filterRecords, useListFilters } from '@/ui/useListFilters';
import type { ListFilters } from '@/ui/useListFilters';
import { CanChange, SettingsViewOnlyNote } from '@/ui/CanChange/CanChange';
import { useCanChange } from '@/ui/useCanChange';
import { SettingDeleteDialog } from '../SettingDeleteDialog';
import { useSettingStatus } from '../useSettingStatus';
import { moreActions, typeCount } from './category-actions';
import type { ActivityCategoryActions } from './category-actions';
import { ActivityCategoryDetails } from './ActivityCategoryDetails';
import { ActivityCategoryForm } from './ActivityCategoryForm';
import {
  ACTIVITY_CATEGORY_LISTS,
  setActivityCategoryStatus,
  useDeleteActivityCategory,
} from './api';
import type { ActivityCategory } from './api';
import { useActivityCategories } from '../api';

const CATEGORY_NOUN = { one: 'activity category', many: 'activity categories' };

/** The side-panel form that is open: a new category, or one being changed. */
type OpenForm = { kind: 'new' } | { kind: 'edit'; category: ActivityCategory };

/**
 * Activity categories (feature 047, US3, FR-018 to FR-021): the groups of activity types the
 * reports open on, as the standard CRUD pattern — search and status filter in the address, a
 * table, details in a side panel (`?id=`), New category and Edit in a side-panel form, Set
 * inactive / Reactivate, and Delete (with Undo).
 */
export function ActivityCategoriesPage() {
  usePageTitle('Activity categories');
  const filters = useListFilters();
  const categories = useActivityCategories();
  const deleteCategory = useDeleteActivityCategory();
  const [form, setForm] = useState<OpenForm | null>(null);
  const [deleting, setDeleting] = useState<ActivityCategory | null>(null);
  const status = useSettingStatus({
    kind: 'activityCategory',
    request: setActivityCategoryStatus,
    invalidate: ACTIVITY_CATEGORY_LISTS,
  });

  const actions: ActivityCategoryActions = {
    view: (category) => filters.setOpenId(category.id),
    edit: (category) => setForm({ kind: 'edit', category }),
    setActive: (category, active) =>
      active ? status.reactivate(category) : status.setInactive(category),
    remove: (category) => setDeleting(category),
  };
  const newCategory = () => setForm({ kind: 'new' });
  // The details wait for the list; a category that is not in it any more says so.
  const openCategory =
    filters.openId && categories.data
      ? (categories.data.items.find((category) => category.id === filters.openId) ?? null)
      : undefined;

  return (
    <>
      <PageHeader
        breadcrumb={[{ label: 'Settings', to: '/settings' }, { label: 'Activity categories' }]}
        title="Activity categories"
        description="The groups of activity types the reports open on."
        actions={
          <CanChange>
            <Button variant="primary" icon={<Plus />} onClick={newCategory}>
              New category
            </Button>
          </CanChange>
        }
      />
      <SettingsViewOnlyNote />
      <CategoriesCard
        categories={categories}
        filters={filters}
        actions={actions}
        onNew={newCategory}
      />

      {form && (
        <ActivityCategoryForm
          category={form.kind === 'edit' ? form.category : undefined}
          onClose={() => setForm(null)}
        />
      )}
      {!form && openCategory !== undefined && (
        <ActivityCategoryDetails
          category={openCategory}
          actions={actions}
          onClose={() => filters.setOpenId('')}
        />
      )}
      {deleting && (
        <SettingDeleteDialog
          kind="activityCategory"
          name={deleting.name}
          usage={deleting.usage}
          active={deleting.active}
          deleteText="It moves to the recycle bin; you can restore it. Its activity types become Uncategorised."
          remove={() => deleteCategory.mutateAsync(deleting.id)}
          setInactive={() => status.applyInactive(deleting)}
          invalidate={ACTIVITY_CATEGORY_LISTS}
          onDeleted={() => {
            if (filters.openId === deleting.id) {
              filters.setOpenId('');
            }
          }}
          onClose={() => setDeleting(null)}
        />
      )}
      {status.dialog}
    </>
  );
}

interface CategoriesCardProps {
  categories: ReturnType<typeof useActivityCategories>;
  filters: ListFilters;
  actions: ActivityCategoryActions;
  onNew: () => void;
}

/** The toolbar and the table, or what to show instead: loading, failure, no categories yet. */
function CategoriesCard({ categories, filters, actions, onNew }: CategoriesCardProps) {
  if (!categories.data) {
    return (
      <Card>
        {categories.isError ? (
          <ErrorState error={categories.error} onRetry={() => void categories.refetch()} />
        ) : (
          <Skeleton lines={5} height="18px" />
        )}
      </Card>
    );
  }

  const all = categories.data.items;
  if (all.length === 0) {
    return (
      <Card>
        <EmptyState
          icon={<Tags />}
          title="No activity categories yet — create your first category"
          action={
            <CanChange>
              <Button icon={<Plus />} onClick={onNew}>
                New category
              </Button>
            </CanChange>
          }
        />
      </Card>
    );
  }

  const { shown, counts } = filterRecords(all, filters, (category) => category.name);

  return (
    <Card>
      <ListToolbar
        search={{
          label: 'Search activity categories',
          placeholder: 'Search name',
          value: filters.q,
          onChange: filters.setQ,
        }}
        status={{ value: filters.status, onChange: filters.setStatus, counts }}
        count={{ shown: shown.length, total: all.length, noun: CATEGORY_NOUN }}
      />
      {shown.length === 0 ? (
        <div className={styles.noMatch}>
          <p>Nothing matches these filters</p>
          <Button onClick={filters.clear}>Clear filters</Button>
        </div>
      ) : (
        <div className={styles.tableScroll}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th scope="col">Name</th>
                <th scope="col">Activity types</th>
                <th scope="col">Default</th>
                <th scope="col">Status</th>
                <th scope="col">
                  <span className="visually-hidden">Actions</span>
                </th>
              </tr>
            </thead>
            <tbody>
              {shown.map((category) => (
                <CategoryRow key={category.id} category={category} actions={actions} />
              ))}
            </tbody>
          </table>
        </div>
      )}
    </Card>
  );
}

interface CategoryRowProps {
  category: ActivityCategory;
  actions: ActivityCategoryActions;
}

function CategoryRow({ category, actions }: CategoryRowProps) {
  const canChange = useCanChange();
  return (
    <tr>
      <th scope="row" className={styles.nameCell}>
        <button type="button" className={styles.nameButton} onClick={() => actions.view(category)}>
          {category.name}
        </button>
      </th>
      <td>{typeCount(category)}</td>
      <td className={styles.nowrap}>
        {category.isDefault ? <StatusChip tone="active">Default</StatusChip> : '—'}
      </td>
      <td className={styles.nowrap}>
        <StatusChip tone={category.active ? 'positive' : 'neutral'}>
          {category.active ? 'Active' : 'Inactive'}
        </StatusChip>
      </td>
      <td className={styles.actionsCell}>
        <ActionsMenu
          label={category.name}
          items={[
            { label: 'View', onSelect: () => actions.view(category) },
            ...(canChange
              ? [
                  { label: 'Edit', onSelect: () => actions.edit(category) },
                  ...moreActions(category, actions),
                ]
              : []),
          ]}
        />
      </td>
    </tr>
  );
}
