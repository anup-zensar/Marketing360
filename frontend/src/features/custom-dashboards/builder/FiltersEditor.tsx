import { Plus, X } from 'lucide-react';
import { Button } from '@/ui/Button/Button';
import { ChoiceField } from '@/ui/ChoiceField/ChoiceField';
import { TextAreaField } from '@/ui/TextAreaField/TextAreaField';
import { TextField } from '@/ui/TextField/TextField';
import type { DataSource, WidgetQuery } from '../model';
import { OPERATOR_WORDS, operatorsFor, valuesFor } from './query-edits';
import { Select } from './Select';
import styles from './Inspector.module.css';

type Filter = WidgetQuery['filters'][number];

const MAX_FILTERS = 10;
const NEW_FILTER: Filter = { field: '', operator: 'eq', values: [''] };

export interface FiltersEditorProps {
  idPrefix: string;
  filters: Filter[];
  source: DataSource;
  errorAt: (path: string) => string | undefined;
  onChange: (filters: Filter[], typing?: string) => void;
}

export function FiltersEditor({
  idPrefix,
  filters,
  source,
  errorAt,
  onChange,
}: FiltersEditorProps) {
  const fieldOptions = source.fields.map((field) => ({
    id: field.key,
    label: field.label,
    active: true,
  }));

  function update(index: number, filter: Filter, typing?: string) {
    onChange(
      filters.map((each, at) => (at === index ? filter : each)),
      typing,
    );
  }

  return (
    <fieldset className={styles.group} id={`${idPrefix}-filters`} tabIndex={-1}>
      <legend className={styles.legend}>Filters</legend>
      {filters.length === 0 && <p className={styles.note}>Every record of the source counts.</p>}
      {filters.map((filter, index) => {
        const field = source.fields.find((each) => each.key === filter.field);
        const operators = operatorsFor(field);
        const at = `query.filters[${index}]`;
        const name = `${idPrefix}-filter-${index}`;
        return (
          <fieldset key={index} className={styles.row}>
            <legend className={styles.rowLegend}>Filter {index + 1}</legend>
            <ChoiceField
              label="Field"
              name={`${name}-field`}
              options={fieldOptions}
              value={filter.field}
              onChange={(key) => {
                const next = source.fields.find((each) => each.key === key);
                const operator = operatorsFor(next).includes(filter.operator)
                  ? filter.operator
                  : 'eq';
                update(index, { field: key, operator, values: valuesFor(operator, filter.values) });
              }}
              error={errorAt(`${at}.field`)}
            />
            <Select
              label="Comparison"
              name={`${name}-operator`}
              value={filter.operator}
              options={operators.map((value) => ({ value, label: OPERATOR_WORDS[value] }))}
              onChange={(value) => {
                const operator = operators.find((each) => each === value);
                if (operator) {
                  update(index, {
                    ...filter,
                    operator,
                    values: valuesFor(operator, filter.values),
                  });
                }
              }}
              error={errorAt(`${at}.operator`)}
            />
            <FilterValues
              name={name}
              filter={filter}
              error={errorAt(`${at}.values`)}
              onChange={(values) => update(index, { ...filter, values }, `${at}.values`)}
            />
            <Button
              variant="tertiary"
              size="sm"
              icon={<X />}
              onClick={() => onChange(filters.filter((_, at) => at !== index))}
            >
              Remove filter {index + 1}
            </Button>
          </fieldset>
        );
      })}
      {filters.length < MAX_FILTERS && (
        <Button size="sm" icon={<Plus />} onClick={() => onChange([...filters, NEW_FILTER])}>
          Add filter
        </Button>
      )}
    </fieldset>
  );
}

interface FilterValuesProps {
  name: string;
  filter: Filter;
  error: string | undefined;
  onChange: (values: string[]) => void;
}

function FilterValues({ name, filter, error, onChange }: FilterValuesProps) {
  const { operator, values } = filter;
  if (operator === 'isEmpty' || operator === 'notEmpty') {
    return null;
  }
  if (operator === 'in') {
    return (
      <TextAreaField
        label="Values"
        hint="One value per line."
        name={`${name}-values`}
        rows={3}
        value={values.join('\n')}
        onChange={(text) => onChange(text.split('\n'))}
        error={error}
      />
    );
  }
  if (operator === 'between') {
    return (
      <div className={styles.pair}>
        <TextField
          label="From"
          name={`${name}-from`}
          value={values[0] ?? ''}
          onChange={(from) => onChange([from, values[1] ?? ''])}
          error={error}
        />
        <TextField
          label="To"
          name={`${name}-to`}
          value={values[1] ?? ''}
          onChange={(to) => onChange([values[0] ?? '', to])}
        />
      </div>
    );
  }
  return (
    <TextField
      label="Value"
      name={`${name}-value`}
      hint="Dates as YYYY-MM-DD; amounts in US dollars."
      value={values[0] ?? ''}
      onChange={(value) => onChange([value])}
      error={error}
    />
  );
}
