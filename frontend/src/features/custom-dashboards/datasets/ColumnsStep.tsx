import { zodResolver } from '@hookform/resolvers/zod';
import { useState } from 'react';
import { useForm, useWatch } from 'react-hook-form';
import type { Path } from 'react-hook-form';
import { ApiError } from '@/api/errors';
import { useDisplayCurrencies } from '@/features/people';
import { plural } from '@/format/text';
import { Button } from '@/ui/Button/Button';
import { FormField } from '@/ui/FormField/FormField';
import { FormPanel } from '@/ui/FormPanel/FormPanel';
import type { SaveFailure } from '@/ui/FormPanel/problem-of';
import { TextAreaField } from '@/ui/TextAreaField/TextAreaField';
import { TextField } from '@/ui/TextField/TextField';
import { useToast } from '@/ui/Toast/useToast';
import { CHECK_ON_SAVE, useFormProblems } from '@/ui/useFormProblems';
import type { Dataset, DatasetPreview } from '../model';
import { useCreateDataset } from './api';
import type { DatasetUpload } from './api';
import { ColumnTypesForm } from './ColumnTypesForm';
import { nameFromFile } from './files';
import { PreviewRows } from './PreviewRows';
import { datasetSchema, startValues } from './schemas';
import type { DatasetValues } from './schemas';
import styles from './datasets.module.css';

interface ColumnsStepProps {
  upload: DatasetUpload;
  preview: DatasetPreview;
  onBack: () => void;
  onClose: () => void;
  onSaved: (dataset: Dataset) => void;
}

export function ColumnsStep({ upload, preview, onBack, onClose, onSaved }: ColumnsStepProps) {
  const { show } = useToast();
  const create = useCreateDataset();
  const currencies = useDisplayCurrencies();
  const form = useForm<DatasetValues>({
    resolver: zodResolver(datasetSchema),
    defaultValues: startValues(preview, nameFromFile(upload.fileName)),
    ...CHECK_ON_SAVE,
  });
  const fields: Path<DatasetValues>[] = [
    'name',
    'description',
    ...preview.columns.flatMap((_, i) => [`columns.${i}.label`, `columns.${i}.currency`] as const),
  ];
  const { problems, onInvalid, showServerProblems, clearProblems } = useFormProblems(form, fields);
  const [failure, setFailure] = useState<SaveFailure | null>(null);
  const columns = useWatch({ control: form.control, name: 'columns' });

  async function save(values: DatasetValues) {
    clearProblems();
    setFailure(null);
    try {
      const saved = await create.mutateAsync({
        ...upload,
        name: values.name.trim(),
        description: values.description,
        columns: values.columns.map((column, position) => ({
          position,
          label: column.label.trim(),
          kind: column.kind,
          currency: column.kind === 'money' ? column.currency : null,
        })),
      });
      show({ message: `Dataset '${saved.name}' saved` });
      onSaved(saved);
    } catch (error) {
      const apiError = error instanceof ApiError ? error : new ApiError('other');
      if (apiError.kind === 'validation' || apiError.kind === 'nameTaken') {
        showServerProblems(fieldProblems(apiError));
      } else if (apiError.kind !== 'signInRequired' && apiError.kind !== 'sessionExpired') {
        setFailure({ message: apiError.message });
      }
    }
  }

  return (
    <FormPanel
      title="Upload data"
      description="Step 2 of 2: Check the columns"
      isDirty
      saving={create.isPending}
      problems={problems.map((p) => ({
        ...p,
        field: p.field.replace(/^(columns\.\d+)$/, '$1.label'),
      }))}
      failure={failure}
      onSave={() => void form.handleSubmit(save, onInvalid)()}
      onClose={onClose}
      saveLabel="Save dataset"
    >
      <p className={styles.facts}>
        {[
          upload.fileName,
          preview.sheetName && `Sheet '${preview.sheetName}'`,
          plural(preview.rowCount, 'row'),
        ]
          .filter(Boolean)
          .join(' · ')}
      </p>
      <div>
        <Button variant="tertiary" onClick={onBack}>
          Choose another file
        </Button>
      </div>
      {preview.problems.length > 0 && (
        <ul className={styles.notes} aria-label="About this file">
          {preview.problems.map((problem) => (
            <li key={problem}>{problem}</li>
          ))}
        </ul>
      )}
      <FormField
        control={form.control}
        name="name"
        render={(field) => <TextField label="Dataset name" marking="Required" {...field} />}
      />
      <FormField
        control={form.control}
        name="description"
        render={(field) => <TextAreaField label="Description" rows={2} {...field} />}
      />
      <h3 className={styles.stepTitle}>Check the columns</h3>
      <ColumnTypesForm
        control={form.control}
        preview={preview}
        currencies={(currencies.data?.items ?? []).map((currency) => ({
          id: currency.code,
          label: `${currency.code} · ${currency.name}`,
          active: currency.active,
        }))}
        problemsOf={(position) =>
          problems.filter((p) => p.field === `columns.${position}`).map((p) => p.message)
        }
      />
      <PreviewRows preview={preview} labels={columns.map((column) => column.label)} />
    </FormPanel>
  );
}

// The API names fields as in its body ("columns[2].label"); the form names them "columns.2.label".
function fieldProblems(error: ApiError) {
  if (error.fields.length === 0) {
    return [{ field: 'name', message: error.message }];
  }
  return error.fields.map((problem) => ({
    field: problem.field.replace(/\[(\d+)\]/g, '.$1'),
    message: problem.message,
  }));
}
