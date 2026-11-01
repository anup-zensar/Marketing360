import { CircleAlert } from 'lucide-react';
import { useWatch } from 'react-hook-form';
import type { Control } from 'react-hook-form';
import { plural } from '@/format/text';
import { ChoiceField } from '@/ui/ChoiceField/ChoiceField';
import type { ChoiceOption } from '@/ui/ChoiceField/ChoiceField';
import { FormField } from '@/ui/FormField/FormField';
import { TextField } from '@/ui/TextField/TextField';
import type { DatasetPreview } from '../model';
import { KIND_OPTIONS, kindOf } from './schemas';
import type { DatasetValues } from './schemas';
import styles from './datasets.module.css';

interface ColumnTypesFormProps {
  control: Control<DatasetValues>;
  preview: DatasetPreview;
  currencies: ChoiceOption[];
  problemsOf: (position: number) => string[];
}

export function ColumnTypesForm({
  control,
  preview,
  currencies,
  problemsOf,
}: ColumnTypesFormProps) {
  return (
    <ol role="list" className={styles.columns} aria-label="Columns">
      {preview.columns.map((column, index) => (
        <li key={column.position}>
          <fieldset className={styles.column}>
            <legend className={styles.legend}>Column {column.position + 1}</legend>
            <FormField
              control={control}
              name={`columns.${index}.label`}
              render={(field) => <TextField label="Name" marking="Required" {...field} />}
            />
            <FormField
              control={control}
              name={`columns.${index}.kind`}
              render={({ onChange, ...field }) => (
                <ChoiceField
                  label="Type"
                  options={[...KIND_OPTIONS]}
                  onChange={(kind) => onChange(kindOf(kind))}
                  {...field}
                />
              )}
            />
            <CurrencyField control={control} index={index} currencies={currencies} />
            <p className={styles.samples}>
              {column.samples.length > 0
                ? `Samples: ${column.samples.join(' · ')}`
                : 'No values in the first rows'}
              {column.emptyCount > 0 && ` (${plural(column.emptyCount, 'empty cell')})`}
            </p>
            {problemsOf(column.position).map((message) => (
              <p key={message} className={styles.problem}>
                <CircleAlert aria-hidden="true" />
                {message}
              </p>
            ))}
          </fieldset>
        </li>
      ))}
    </ol>
  );
}

interface CurrencyFieldProps {
  control: Control<DatasetValues>;
  index: number;
  currencies: ChoiceOption[];
}

function CurrencyField({ control, index, currencies }: CurrencyFieldProps) {
  const kind = useWatch({ control, name: `columns.${index}.kind` });
  if (kind !== 'money') {
    return null;
  }
  return (
    <FormField
      control={control}
      name={`columns.${index}.currency`}
      render={({ value, onChange, ...field }) => (
        <ChoiceField
          label="Currency"
          marking="Required"
          hint="Amounts are shown in US dollars at the current rate."
          options={currencies}
          value={value ?? ''}
          onChange={(code) => onChange(code || null)}
          {...field}
        />
      )}
    />
  );
}
