import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { CheckboxField } from '@/ui/CheckboxField/CheckboxField';
import { FormField } from '@/ui/FormField/FormField';
import { FormPanel } from '@/ui/FormPanel/FormPanel';
import { TextField } from '@/ui/TextField/TextField';
import { useToast } from '@/ui/Toast/useToast';
import { CHECK_ON_SAVE, useFormProblems } from '@/ui/useFormProblems';
import type { SaveFailure } from '@/ui/FormPanel/problem-of';
import { useFormatters } from '@/format/useFormatters';
import styles from '@/ui/settings-table.module.css';
import { saveFailureOf } from '../save-failure';
import { ActivityTypesField } from './ActivityTypesField';
import { changedFields, createBody, startValues } from './category-values';
import {
  fetchActivityCategory,
  useCreateActivityCategory,
  useUpdateActivityCategory,
} from './api';
import type { ActivityCategory } from './api';
import { activityCategorySchema } from './schemas';
import type { ActivityCategoryValues } from './schemas';

const FIELDS = ['name', 'isDefault', 'activityTypeIds'] as const;

const DEFAULT_HINT =
  'The reports open on this category. Marking a new default unmarks the category that had it.';

export interface ActivityCategoryFormProps {
  /** The activity category to change; without one, the form adds a new category. */
  category?: ActivityCategory;
  onClose: () => void;
}

/**
 * New activity category and Edit activity category (FR-018 to FR-020) in a side panel: the name,
 * whether it is the default the reports open on, and the activity types it groups. Every problem
 * shows at once; a conflict keeps what was typed and names who saved first.
 */
export function ActivityCategoryForm({ category, onClose }: ActivityCategoryFormProps) {
  const { show } = useToast();
  const { dateFormat } = useFormatters();
  const createCategory = useCreateActivityCategory();
  const updateCategory = useUpdateActivityCategory();
  // The edit-form rule: the category is copied into the form once, when the panel opens.
  const [start] = useState(() => startValues(category));
  const form = useForm<ActivityCategoryValues>({
    resolver: zodResolver(activityCategorySchema),
    defaultValues: start,
    ...CHECK_ON_SAVE,
  });
  const { problems, onInvalid, showServerProblems, clearProblems } = useFormProblems(form, FIELDS);
  const [version, setVersion] = useState(category?.version ?? 0);
  const [saving, setSaving] = useState(false);
  const [failure, setFailure] = useState<SaveFailure | null>(null);

  async function save(values: ActivityCategoryValues) {
    clearProblems();
    setSaving(true);
    try {
      if (!category) {
        const created = await createCategory.mutateAsync(createBody(values));
        show({ message: `Activity category '${created.name}' created` });
        onClose();
        return;
      }
      const body = changedFields(start, values, version);
      if (!body) {
        onClose();
        return;
      }
      const updated = await updateCategory.mutateAsync({ id: category.id, body });
      show({ message: `Activity category '${updated.name}' saved` });
      onClose();
    } catch (error) {
      await showFailure(error);
    } finally {
      setSaving(false);
    }
  }

  async function showFailure(error: unknown) {
    const view = saveFailureOf(error, 'activityCategory', dateFormat);
    showServerProblems(view.problems);
    setFailure(view.failure);
    if (view.conflict && category) {
      // Typed values stay; the next Save replaces the other person's change on purpose.
      const latest = await fetchActivityCategory(category.id).catch(() => undefined);
      if (latest) {
        setVersion(latest.version);
      }
    }
  }

  const submit = form.handleSubmit(save, onInvalid);

  return (
    <FormPanel
      title={category ? 'Edit activity category' : 'New activity category'}
      isDirty={form.formState.isDirty}
      saving={saving}
      problems={problems}
      failure={failure}
      onSave={() => void submit()}
      onClose={onClose}
    >
      <div className={styles.formFields}>
        <FormField
          control={form.control}
          name="name"
          render={(field) => (
            <TextField label="Name" marking="Required" maxLength={80} {...field} />
          )}
        />
        <FormField
          control={form.control}
          name="isDefault"
          render={(field) => (
            <CheckboxField
              label="Default category"
              name={field.name}
              checked={field.value}
              onChange={field.onChange}
              hint={DEFAULT_HINT}
              error={field.error}
            />
          )}
        />
        <FormField
          control={form.control}
          name="activityTypeIds"
          render={(field) => <ActivityTypesField category={category} {...field} />}
        />
      </div>
    </FormPanel>
  );
}
