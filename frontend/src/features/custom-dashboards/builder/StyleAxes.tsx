import { CheckboxField } from '@/ui/CheckboxField/CheckboxField';
import { TextField } from '@/ui/TextField/TextField';
import type { WidgetDisplay } from '../model';
import { Select } from './Select';
import { shows } from './style-options';
import type { StyleSectionProps } from './style-options';
import { defaultMeasureStyle } from './widget-defaults';
import styles from './Inspector.module.css';

type AxisTitles = NonNullable<WidgetDisplay['axisTitles']>;
type MeasureStyle = NonNullable<WidgetDisplay['measureStyles']>[number];

const NO_TITLES: AxisTitles = { x: '', y: '', y2: '' };
const STACKS = [
  { value: 'none', label: 'None' },
  { value: 'stacked', label: 'Stacked' },
  { value: 'percent', label: '100%' },
] as const;
const ORIENTATIONS = [
  { value: 'vertical', label: 'Vertical' },
  { value: 'horizontal', label: 'Horizontal' },
] as const;

export function StyleAxes({ widget, prefix, errorAt, set }: StyleSectionProps) {
  const { type, display, query } = widget;
  const titles = display.axisTitles ?? NO_TITLES;
  const reference = display.referenceLine ?? null;
  const setTitle = (axis: keyof AxisTitles, text: string) =>
    set({ axisTitles: { ...titles, [axis]: text } }, `axisTitles.${axis}`);

  return (
    <>
      {shows(type, 'axisTitles') && (
        <fieldset className={styles.group}>
          <legend className={styles.rowLegend}>Axis titles</legend>
          <TextField
            label="Horizontal"
            name={`${prefix}-axis-x`}
            maxLength={40}
            value={titles.x}
            onChange={(text) => setTitle('x', text)}
            error={errorAt('display.axisTitles.x')}
          />
          <TextField
            label="Vertical"
            name={`${prefix}-axis-y`}
            maxLength={40}
            value={titles.y}
            onChange={(text) => setTitle('y', text)}
            error={errorAt('display.axisTitles.y')}
          />
          {type === 'combo' && (
            <TextField
              label="Right"
              name={`${prefix}-axis-y2`}
              maxLength={40}
              value={titles.y2}
              onChange={(text) => setTitle('y2', text)}
              error={errorAt('display.axisTitles.y2')}
            />
          )}
        </fieldset>
      )}
      {shows(type, 'gridlines') && (
        <CheckboxField
          label="Gridlines"
          name={`${prefix}-gridlines`}
          checked={display.gridlines ?? true}
          onChange={(gridlines) => set({ gridlines })}
        />
      )}
      {shows(type, 'smooth') && (
        <CheckboxField
          label="Smooth lines"
          name={`${prefix}-smooth`}
          checked={display.smooth ?? false}
          onChange={(smooth) => set({ smooth })}
        />
      )}
      {shows(type, 'stack') && query?.series && (
        <Select
          label="Stacking"
          name={`${prefix}-stack`}
          value={display.stack ?? (type === 'area' ? 'stacked' : 'none')}
          options={[...STACKS]}
          onChange={(value) => set({ stack: STACKS.find((each) => each.value === value)?.value })}
        />
      )}
      {shows(type, 'orientation') && (
        <Select
          label="Orientation"
          name={`${prefix}-orientation`}
          value={display.orientation ?? 'vertical'}
          options={[...ORIENTATIONS]}
          onChange={(value) =>
            set({ orientation: ORIENTATIONS.find((each) => each.value === value)?.value })
          }
        />
      )}
      {shows(type, 'zoom') && (
        <CheckboxField
          label="Zoom slider"
          name={`${prefix}-zoom`}
          checked={display.zoom ?? false}
          onChange={(zoom) => set({ zoom })}
        />
      )}
      {shows(type, 'referenceLine') && (
        <fieldset className={styles.group}>
          <legend className={styles.rowLegend}>Reference line</legend>
          <div className={styles.pair}>
            <TextField
              label="Value"
              name={`${prefix}-reference-value`}
              value={reference?.value ?? ''}
              onChange={(value) =>
                set(
                  {
                    referenceLine:
                      value || reference?.label ? { label: '', ...reference, value } : null,
                  },
                  'referenceLine.value',
                )
              }
              error={errorAt('display.referenceLine.value')}
            />
            <TextField
              label="Label"
              name={`${prefix}-reference-label`}
              maxLength={40}
              value={reference?.label ?? ''}
              onChange={(label) =>
                set(
                  {
                    referenceLine:
                      label || reference?.value ? { value: '', ...reference, label } : null,
                  },
                  'referenceLine.label',
                )
              }
              error={errorAt('display.referenceLine.label')}
            />
          </div>
        </fieldset>
      )}
      {shows(type, 'measureStyles') && (
        <MeasureStyles
          prefix={prefix}
          labels={(query?.measures ?? []).map(
            (measure, index) => measure.label || `Measure ${index + 1}`,
          )}
          value={display.measureStyles}
          onChange={(measureStyles) => set({ measureStyles })}
        />
      )}
    </>
  );
}

interface MeasureStylesProps {
  prefix: string;
  labels: string[];
  value: MeasureStyle[] | undefined;
  onChange: (styles: MeasureStyle[]) => void;
}

function MeasureStyles({ prefix, labels, value, onChange }: MeasureStylesProps) {
  const current = labels.map((_, index) => value?.[index] ?? defaultMeasureStyle(index));
  const update = (index: number, patch: Partial<MeasureStyle>) =>
    onChange(current.map((each, at) => (at === index ? { ...each, ...patch } : each)));
  return (
    <>
      {labels.map((label, index) => (
        <fieldset key={`${label}-${index}`} className={styles.row}>
          <legend className={styles.rowLegend}>{label}</legend>
          <div className={styles.pair}>
            <Select
              label="Bars or line"
              name={`${prefix}-style-${index}-as`}
              value={current[index]?.as ?? 'bar'}
              options={[
                { value: 'bar', label: 'Bars' },
                { value: 'line', label: 'Line' },
              ]}
              onChange={(as) => update(index, { as: as === 'line' ? 'line' : 'bar' })}
            />
            <Select
              label="Axis"
              name={`${prefix}-style-${index}-axis`}
              value={current[index]?.axis ?? 'left'}
              options={[
                { value: 'left', label: 'Left' },
                { value: 'right', label: 'Right' },
              ]}
              onChange={(axis) => update(index, { axis: axis === 'right' ? 'right' : 'left' })}
            />
          </div>
        </fieldset>
      ))}
    </>
  );
}
