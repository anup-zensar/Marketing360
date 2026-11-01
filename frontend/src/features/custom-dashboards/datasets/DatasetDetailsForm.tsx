import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import type { Path } from 'react-hook-form';
import { ApiError, conflictMessage, isSessionLost } from '@/api/errors';
import { FormField } from '@/ui/FormField/FormField';
import { FormPanel } from '@/ui/FormPanel/FormPanel';
import type { SaveFailure } from '@/ui/FormPanel/problem-of';
import { TextAreaField } from '@/ui/TextAreaField/TextAreaField';
import { TextField } from '@/ui/TextField/TextField';
import { useToast } from '@/ui/Toast/useToast';
import { CHECK_ON_SAVE, useFormProblems } from '@/ui/useFormProblems';
import type { Dataset } from '../model';
import { useUpdateDataset } from './api';
import type { DatasetUpdate } from './api';
import { detailsSchema } from './schemas';
import type { DetailsValues } from './schemas';
import styles from './datasets.module.css';

const CONFLICT_STEP =
  'Your changes are still here. Cancel and open Edit again to start from theirs.';

interface DatasetDetailsFormProps {
  dataset: Dataset;
  onClose: () => void;
}

export function DatasetDetailsForm({ dataset, onClose }: DatasetDetailsFormProps) {
  const { show } = useToast();
  const update = useUpdateDataset();
  const form = useForm<DetailsValues>({
    resolver: zodResolver(detailsSchema),
    defaultValues: {
      name: dataset.name,
      description: dataset.description,
      labels: dataset.columns.map((column) => column.label),
    },
    ...CHECK_ON_SAVE,
  });
  const fields: Path<DetailsValues>[] = [
    'name',
    'description',
    ...dataset.columns.map((_, i) => `labels.${i}` as const),
  ];
  const { problems, onInvalid, showServerProblems, clearProblems } = useFormProblems(form, fields);
  const [failure, setFailure] = useState<SaveFailure | null>(null);

  async function save(values: DetailsValues) {
    clearProblems();
    setFailure(null);
    try {
      const saved = await update.mutateAsync({ id: dataset.id, body: changes(dataset, values) });
      show({ message: `Dataset '${saved.name}' saved` });
      onClose();
    } catch (error) {
      const apiError = error instanceof ApiError ? error : new ApiError('other');
      if (apiError.kind === 'validation') {
        showServerProblems(apiError.fields.map(toFormField));
      } else if (apiError.kind === 'conflict') {
        setFailure({ message: conflictMessage(apiError, 'dataset'), nextStep: CONFLICT_STEP });
      } else if (!isSessionLost(apiError)) {
        setFailure({ message: apiError.message });
      }
    }
  }

  return (
    <FormPanel
      title="Edit dataset"
      isDirty={form.formState.isDirty}
      saving={update.isPending}
      problems={problems}
      failure={failure}
      onSave={() => void form.handleSubmit(save, onInvalid)()}
      onClose={onClose}
    >
      <FormField
        control={form.control}
        name="name"
        render={(field) => <TextField label="Name" marking="Required" {...field} />}
      />
      <FormField
        control={form.control}
        name="description"
        render={(field) => <TextAreaField label="Description" rows={3} {...field} />}
      />
      <h3 className={styles.stepTitle}>Column names</h3>
      <p className={styles.facts}>Column types are fixed once a dataset is saved.</p>
      {dataset.columns.map((column, index) => (
        <FormField
          key={column.key}
          control={form.control}
          name={`labels.${index}`}
          render={(field) => (
            <TextField label={`Column ${column.position + 1}`} marking="Required" {...field} />
          )}
        />
      ))}
    </FormPanel>
  );
}

function changes(dataset: Dataset, values: DetailsValues): DatasetUpdate {
  return {
    version: dataset.version,
    name: values.name.trim(),
    description: values.description,
    columns: dataset.columns
      .map((column, index) => ({ position: column.position, label: values.labels[index] ?? '' }))
      .filter((column, index) => column.label.trim() !== dataset.columns[index]?.label)
      .map((column) => ({ ...column, label: column.label.trim() })),
  };
}

function toFormField(problem: { field: string; message: string }) {
  const column = /^columns\[(\d+)\]/.exec(problem.field);
  return { field: column ? `labels.${column[1]}` : problem.field, message: problem.message };
}
