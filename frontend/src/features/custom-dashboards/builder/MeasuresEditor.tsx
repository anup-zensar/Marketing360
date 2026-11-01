import { Plus, X } from 'lucide-react';
import { Button } from '@/ui/Button/Button';
import { ChoiceField } from '@/ui/ChoiceField/ChoiceField';
import { TextField } from '@/ui/TextField/TextField';
import type { DataSource, WidgetQuery } from '../model';
import { AGGREGATION_WORDS, aggregationsFor } from './query-edits';
import { Select } from './Select';
import styles from './Inspector.module.css';

type Measure = WidgetQuery['measures'][number];

const RECORDS = '#records';
const NEW_MEASURE: Measure = { aggregation: 'count', field: null, label: '' };

export interface MeasuresEditorProps {
  idPrefix: string;
  measures: Measure[];
  max: number;
  source: DataSource;
  errorAt: (path: string) => string | undefined;
  onChange: (measures: Measure[], typing?: string) => void;
}

export function MeasuresEditor({
  idPrefix,
  measures,
  max,
  source,
  errorAt,
  onChange,
}: MeasuresEditorProps) {
  const fieldOptions = [
    { id: RECORDS, label: 'Count of records', active: true },
    ...source.fields.map((field) => ({ id: field.key, label: field.label, active: true })),
  ];

  function update(index: number, measure: Measure, typing?: string) {
    onChange(
      measures.map((each, at) => (at === index ? measure : each)),
      typing,
    );
  }

  function chooseField(index: number, measure: Measure, key: string) {
    const field = source.fields.find((each) => each.key === key);
    const allowed = aggregationsFor(field);
    const kept = allowed.includes(measure.aggregation) ? measure.aggregation : 'count';
    const aggregation = allowed.includes('sum') ? 'sum' : kept;
    update(index, { ...measure, field: field ? field.key : null, aggregation });
  }

  return (
    <fieldset className={styles.group} id={`${idPrefix}-measures`} tabIndex={-1}>
      <legend className={styles.legend}>Measures</legend>
      {errorAt('query.measures') && <p className={styles.error}>{errorAt('query.measures')}</p>}
      {measures.map((measure, index) => {
        const field = source.fields.find((each) => each.key === measure.field);
        const at = `query.measures[${index}]`;
        return (
          <fieldset key={index} className={styles.row}>
            <legend className={styles.rowLegend}>Measure {index + 1}</legend>
            <ChoiceField
              label="What to measure"
              name={`${idPrefix}-measure-${index}-field`}
              options={fieldOptions}
              value={measure.field ?? RECORDS}
              onChange={(key) => chooseField(index, measure, key)}
              error={errorAt(`${at}.field`)}
            />
            {field && (
              <Select
                label="Aggregation"
                name={`${idPrefix}-measure-${index}-aggregation`}
                value={measure.aggregation}
                options={aggregationsFor(field).map((value) => ({
                  value,
                  label: AGGREGATION_WORDS[value],
                }))}
                onChange={(value) => {
                  const aggregation = aggregationsFor(field).find((each) => each === value);
                  if (aggregation) {
                    update(index, { ...measure, aggregation });
                  }
                }}
              />
            )}
            <TextField
              label="Label"
              name={`${idPrefix}-measure-${index}-label`}
              hint="Empty uses a name such as 'Sum of Planned cost'."
              maxLength={40}
              value={measure.label}
              onChange={(label) => update(index, { ...measure, label }, `${at}.label`)}
              error={errorAt(`${at}.label`)}
            />
            <Button
              variant="tertiary"
              size="sm"
              icon={<X />}
              onClick={() => onChange(measures.filter((_, at) => at !== index))}
            >
              Remove measure {index + 1}
            </Button>
          </fieldset>
        );
      })}
      {measures.length < max && (
        <Button size="sm" icon={<Plus />} onClick={() => onChange([...measures, NEW_MEASURE])}>
          Add measure
        </Button>
      )}
    </fieldset>
  );
}
