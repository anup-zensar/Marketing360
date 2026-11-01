import { CheckboxField } from '@/ui/CheckboxField/CheckboxField';
import { TextField } from '@/ui/TextField/TextField';
import type { WidgetDisplay } from '../model';
import { Select } from './Select';
import { shows } from './style-options';
import type { StyleSectionProps } from './style-options';
import { tableColumnKeys } from './widget-defaults';
import styles from './Inspector.module.css';

type Kpi = NonNullable<WidgetDisplay['kpi']>;

const AUTOMATIC = '';
const FORMATS = [
  { value: 'number', label: 'Number' },
  { value: 'money', label: 'Money' },
  { value: 'percent', label: 'Percent' },
  { value: 'compact', label: 'Compact (1.2M)' },
] as const;
const DECIMALS = ['0', '1', '2', '3', '4'].map((value) => ({ value, label: value }));
const COMPARISONS = [
  { value: 'none', label: 'Nothing' },
  { value: 'target', label: 'Target' },
  { value: 'previousPeriod', label: 'Previous period' },
] as const;
const NO_KPI: Kpi = { comparison: 'none', target: null, higherIsBetter: true };

export function StyleValues({ widget, prefix, errorAt, set }: StyleSectionProps) {
  const { type, display, query } = widget;
  const kpi = display.kpi ?? NO_KPI;
  const gauge = display.gauge ?? { min: '0', max: null };
  const table = display.table ?? { columns: [], totals: false };
  const columns = tableColumnKeys(query);
  const shown = table.columns.length > 0 ? table.columns : columns;

  return (
    <>
      {shows(type, 'numberFormat') && (
        <div className={styles.pair}>
          <Select
            label="Number format"
            name={`${prefix}-format`}
            value={display.numberFormat ?? AUTOMATIC}
            placeholder="Automatic"
            options={[...FORMATS]}
            onChange={(value) =>
              set({ numberFormat: FORMATS.find((each) => each.value === value)?.value })
            }
          />
          <Select
            label="Decimals"
            name={`${prefix}-decimals`}
            value={display.decimals === undefined ? AUTOMATIC : String(display.decimals)}
            placeholder="Automatic"
            options={DECIMALS}
            onChange={(value) => set({ decimals: value === AUTOMATIC ? undefined : Number(value) })}
          />
        </div>
      )}
      {shows(type, 'gauge') && (
        <div className={styles.pair}>
          <TextField
            label="Minimum"
            name={`${prefix}-gauge-min`}
            value={gauge.min}
            onChange={(min) => set({ gauge: { ...gauge, min } }, 'gauge.min')}
            error={errorAt('display.gauge.min')}
          />
          <TextField
            label="Maximum"
            name={`${prefix}-gauge-max`}
            hint="Empty rounds the value up."
            value={gauge.max ?? ''}
            onChange={(max) => set({ gauge: { ...gauge, max: max || null } }, 'gauge.max')}
            error={errorAt('display.gauge.max')}
          />
        </div>
      )}
      {shows(type, 'kpi') && (
        <>
          <Select
            label="Compare with"
            name={`${prefix}-comparison`}
            value={kpi.comparison}
            options={[...COMPARISONS]}
            onChange={(value) => {
              const comparison = COMPARISONS.find((each) => each.value === value)?.value ?? 'none';
              set({ kpi: { ...kpi, comparison } });
            }}
          />
          {kpi.comparison === 'previousPeriod' && (
            <p className={styles.note}>Needs a date range in the dashboard's filters.</p>
          )}
          {kpi.comparison === 'target' && (
            <TextField
              label="Target"
              name={`${prefix}-target`}
              marking="Required"
              value={kpi.target ?? ''}
              onChange={(target) => set({ kpi: { ...kpi, target: target || null } }, 'kpi.target')}
              error={errorAt('display.kpi.target')}
            />
          )}
          <CheckboxField
            label="Higher is better"
            name={`${prefix}-higher`}
            checked={kpi.higherIsBetter}
            onChange={(higherIsBetter) => set({ kpi: { ...kpi, higherIsBetter } })}
          />
        </>
      )}
      {shows(type, 'table') && (
        <>
          <fieldset className={styles.group}>
            <legend className={styles.rowLegend}>Columns</legend>
            {columns.length === 0 && <p className={styles.note}>Choose the data first.</p>}
            {columns.map((key) => (
              <CheckboxField
                key={key}
                label={columnLabel(key, widget)}
                name={`${prefix}-column-${key}`}
                checked={shown.includes(key)}
                onChange={(checked) =>
                  set({
                    table: {
                      ...table,
                      columns: columns.filter((each) =>
                        each === key ? checked : shown.includes(each),
                      ),
                    },
                  })
                }
              />
            ))}
          </fieldset>
          <CheckboxField
            label="Totals row"
            name={`${prefix}-totals`}
            checked={table.totals}
            onChange={(totals) => set({ table: { ...table, totals } })}
          />
        </>
      )}
    </>
  );
}

function columnLabel(key: string, widget: StyleSectionProps['widget']): string {
  if (key === 'dimension') {
    return 'The group';
  }
  if (key === 'series') {
    return 'The breakdown';
  }
  const index = Number(key.slice(1));
  return widget.query?.measures[index]?.label || `Measure ${index + 1}`;
}
