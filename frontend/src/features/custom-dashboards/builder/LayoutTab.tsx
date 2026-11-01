import { useState } from 'react';
import type { Widget, WidgetLayout } from '../model';
import { placeMessage } from './announcements';
import { blockedBy } from './grid';
import { NumberInput } from './NumberInput';
import styles from './Inspector.module.css';

export interface LayoutTabProps {
  widget: Widget;
  widgets: readonly Widget[];
  onPlace: (layout: WidgetLayout) => void;
}

const PARTS = [
  { key: 'x', label: 'Column', offset: 1, min: 1, max: 12 },
  { key: 'y', label: 'Row', offset: 1, min: 1, max: 200 },
  { key: 'w', label: 'Width (columns)', offset: 0, min: 2, max: 12 },
  { key: 'h', label: 'Height (rows)', offset: 0, min: 2, max: 12 },
] as const;

export function LayoutTab({ widget, widgets, onPlace }: LayoutTabProps) {
  const [typed, setTyped] = useState<WidgetLayout | null>(null);
  const shown = typed ?? widget.layout;
  const blocked = typed && blockedBy(typed, widgets, widget.id);

  function change(key: keyof WidgetLayout, value: number, offset: number) {
    const next = { ...shown, [key]: value - offset };
    if (blockedBy(next, widgets, widget.id) === null) {
      setTyped(null);
      onPlace(next);
    } else {
      setTyped(next);
    }
  }

  return (
    <div className={styles.tab}>
      {PARTS.map(({ key, label, offset, min, max }) => (
        <NumberInput
          key={key}
          label={label}
          name={`inspector-${widget.id}-layout-${key}`}
          min={min}
          max={max}
          value={shown[key] + offset}
          onChange={(value) => change(key, value, offset)}
        />
      ))}
      {blocked && (
        <p className={styles.error} role="alert">
          {placeMessage({
            blocked,
            resizing: typed.x === widget.layout.x && typed.y === widget.layout.y,
            layout: typed,
          })}
        </p>
      )}
    </div>
  );
}
