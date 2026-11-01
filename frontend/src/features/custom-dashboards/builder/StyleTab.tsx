import { CheckboxField } from '@/ui/CheckboxField/CheckboxField';
import { TextAreaField } from '@/ui/TextAreaField/TextAreaField';
import { TextField } from '@/ui/TextField/TextField';
import type { Widget, WidgetDisplay } from '../model';
import { Select } from './Select';
import { StyleAxes } from './StyleAxes';
import { defaultDataLabels, defaultPalette, PALETTES, shows } from './style-options';
import type { StyleSectionProps } from './style-options';
import { StyleValues } from './StyleValues';
import styles from './Inspector.module.css';

type TextPatch = Partial<Pick<Widget, 'title' | 'subtitle' | 'note'>>;
type Legend = NonNullable<WidgetDisplay['legend']>;

const LEGEND_POSITIONS: { value: Legend['position']; label: string }[] = [
  { value: 'top', label: 'Top' },
  { value: 'bottom', label: 'Bottom' },
  { value: 'left', label: 'Left' },
  { value: 'right', label: 'Right' },
];

export interface StyleTabProps {
  widget: Widget;
  errorAt: (path: string) => string | undefined;
  onDisplay: (display: WidgetDisplay, typing?: string) => void;
  onText: (patch: TextPatch) => void;
}

export function StyleTab({ widget, errorAt, onDisplay, onText }: StyleTabProps) {
  const { type, display } = widget;
  const prefix = `inspector-${widget.id}`;
  const legend: Legend = display.legend ?? { show: true, position: 'bottom' };
  const set = (patch: Partial<WidgetDisplay>, typing?: string) =>
    onDisplay({ ...display, ...patch }, typing);
  const section: StyleSectionProps = { widget, prefix, errorAt, set };

  return (
    <div className={styles.tab}>
      <TextField
        label="Title"
        name={`${prefix}-title`}
        hint="Empty uses the source and measure."
        maxLength={80}
        value={widget.title}
        onChange={(title) => onText({ title })}
        error={errorAt('title')}
      />
      <TextField
        label="Subtitle"
        name={`${prefix}-subtitle`}
        maxLength={120}
        value={widget.subtitle}
        onChange={(subtitle) => onText({ subtitle })}
        error={errorAt('subtitle')}
      />
      {shows(type, 'note') && (
        <TextAreaField
          label="Note"
          name={`${prefix}-note`}
          hint="Plain text, shown as written."
          maxLength={2000}
          rows={6}
          value={widget.note}
          onChange={(note) => onText({ note })}
          error={errorAt('note')}
        />
      )}
      {shows(type, 'palette') && (
        <fieldset className={styles.group}>
          <legend className={styles.rowLegend}>Colours</legend>
          <div className={styles.palettes}>
            {PALETTES.map((palette) => (
              <label key={palette.value} className={styles.paletteChoice}>
                <input
                  type="radio"
                  name={`${prefix}-palette`}
                  value={palette.value}
                  checked={(display.palette ?? defaultPalette(type)) === palette.value}
                  onChange={() => set({ palette: palette.value })}
                />
                {palette.label}
                <span className={styles.swatches} aria-hidden="true">
                  {palette.tokens.map((token) => (
                    <span
                      key={token}
                      className={styles.swatch}
                      style={{ backgroundColor: `var(${token})` }}
                    />
                  ))}
                </span>
              </label>
            ))}
          </div>
        </fieldset>
      )}
      {shows(type, 'legend') && (
        <div className={styles.pair}>
          <CheckboxField
            label="Show the legend"
            name={`${prefix}-legend`}
            checked={legend.show}
            onChange={(show) => set({ legend: { ...legend, show } })}
          />
          <Select
            label="Legend position"
            name={`${prefix}-legend-position`}
            value={legend.position}
            options={LEGEND_POSITIONS}
            onChange={(value) => {
              const position = LEGEND_POSITIONS.find((each) => each.value === value)?.value;
              set({ legend: { ...legend, position: position ?? 'bottom' } });
            }}
          />
        </div>
      )}
      {shows(type, 'dataLabels') && (
        <CheckboxField
          label="Data labels"
          name={`${prefix}-data-labels`}
          checked={display.dataLabels ?? defaultDataLabels(type)}
          onChange={(dataLabels) => set({ dataLabels })}
        />
      )}
      <StyleAxes {...section} />
      <StyleValues {...section} />
    </div>
  );
}
