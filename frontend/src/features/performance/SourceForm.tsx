import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { PersonField } from '@/features/people';
import { ChoiceField } from '@/ui/ChoiceField/ChoiceField';
import { FormField } from '@/ui/FormField/FormField';
import { FormPanel } from '@/ui/FormPanel/FormPanel';
import { TextField } from '@/ui/TextField/TextField';
import { useToast } from '@/ui/Toast/useToast';
import { CHECK_ON_SAVE, useFormProblems } from '@/ui/useFormProblems';
import type { SaveFailure } from '@/ui/FormPanel/problem-of';
import { asApiError, CONFLICT_NEXT_STEP, conflictMessage } from '@/api/errors';
import styles from '@/ui/settings-table.module.css';
import { useCreatePerformanceSource, useUpdatePerformanceSource } from './api';
import type { PerformanceSource, PerformanceSourceInput } from './api';
import { CADENCE_WORDS, CATEGORY_WORDS, DELIVERY_WORDS } from './performance-text';
import { sourceSchema } from './schemas';
import type { SourceValues } from './schemas';

const FIELDS = [
  'name',
  'description',
  'ownerId',
  'category',
  'deliveryRoute',
  'cadence',
  'integrationId',
] as const;

function choicesOf(words: Record<string, string>) {
  return Object.entries(words).map(([id, label]) => ({ id, label, active: true }));
}

/** The chosen option as its literal type: an id not in `known` keeps the current value. */
function narrowed<T extends string>(known: readonly T[], value: string, keep: T): T {
  return known.find((option) => option === value) ?? keep;
}

function startValues(source?: PerformanceSource): SourceValues {
  return {
    name: source?.name ?? '',
    description: source?.description ?? '',
    ownerId: source?.ownerId ?? '',
    category: source?.category ?? 'webAnalytics',
    deliveryRoute: source?.deliveryRoute ?? 'upload',
    cadence: source?.cadence ?? 'monthly',
    integrationId: source?.integrationId ?? '',
  };
}

export interface SourceFormProps {
  /** The source to change; without one, the form adds a new data source. */
  source?: PerformanceSource;
  onClose: () => void;
}

/**
 * New data source and Edit data source (046 FR-001) in a side panel: name, owner, category, how
 * its data arrives and how often, and the source's own identifier for integration-ID matching.
 */
export function SourceForm({ source, onClose }: SourceFormProps) {
  const { show } = useToast();
  const createSource = useCreatePerformanceSource();
  const updateSource = useUpdatePerformanceSource(source?.id ?? '');
  const form = useForm<SourceValues>({
    resolver: zodResolver(sourceSchema),
    defaultValues: startValues(source),
    ...CHECK_ON_SAVE,
  });
  const { problems, onInvalid, showServerProblems, clearProblems } = useFormProblems(form, FIELDS);
  const [saving, setSaving] = useState(false);
  const [failure, setFailure] = useState<SaveFailure | null>(null);

  async function save(values: SourceValues) {
    clearProblems();
    setSaving(true);
    try {
      const body: PerformanceSourceInput = { ...values, version: source?.version };
      if (source) {
        const saved = await updateSource.mutateAsync(body);
        show({ message: `Data source '${saved.name}' saved` });
      } else {
        const created = await createSource.mutateAsync(body);
        show({ message: `Data source '${created.name}' created` });
      }
      onClose();
    } catch (error) {
      const failed = asApiError(error);
      if (failed.kind === 'validation') {
        showServerProblems(failed.fields);
      } else if (failed.kind === 'conflict') {
        setFailure({
          message: conflictMessage(failed, 'data source'),
          nextStep: CONFLICT_NEXT_STEP,
        });
      } else {
        setFailure({ message: failed.message });
      }
    } finally {
      setSaving(false);
    }
  }

  const submit = form.handleSubmit(save, onInvalid);

  return (
    <FormPanel
      title={source ? 'Edit data source' : 'New data source'}
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
          render={(field) => <TextField label="Name" marking="Required" {...field} />}
        />
        <FormField
          control={form.control}
          name="description"
          render={(field) => <TextField label="Description" {...field} />}
        />
        <FormField
          control={form.control}
          name="ownerId"
          render={(field) => <PersonField label="Owner" marking="Required" {...field} />}
        />
        <FormField
          control={form.control}
          name="category"
          render={(field) => (
            <ChoiceField
              label="Category"
              options={choicesOf(CATEGORY_WORDS)}
              {...field}
              onChange={(value) =>
                field.onChange(narrowed(sourceSchema.shape.category.options, value, field.value))
              }
            />
          )}
        />
        <FormField
          control={form.control}
          name="deliveryRoute"
          render={(field) => (
            <ChoiceField
              label="Delivery"
              hint="How rows arrive: a file upload, or a pull from the simulated feed."
              options={choicesOf(DELIVERY_WORDS)}
              {...field}
              onChange={(value) =>
                field.onChange(
                  narrowed(sourceSchema.shape.deliveryRoute.options, value, field.value),
                )
              }
            />
          )}
        />
        <FormField
          control={form.control}
          name="cadence"
          render={(field) => (
            <ChoiceField
              label="Cadence"
              options={choicesOf(CADENCE_WORDS)}
              {...field}
              onChange={(value) =>
                field.onChange(narrowed(sourceSchema.shape.cadence.options, value, field.value))
              }
            />
          )}
        />
        <FormField
          control={form.control}
          name="integrationId"
          render={(field) => (
            <TextField
              label="Integration ID"
              hint="The source's own identifier for its rows' integration-ID key; leave empty without one."
              {...field}
            />
          )}
        />
      </div>
    </FormPanel>
  );
}
