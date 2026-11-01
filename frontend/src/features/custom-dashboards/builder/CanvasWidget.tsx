import { GripVertical, MoveDiagonal2 } from 'lucide-react';
import type { KeyboardEvent } from 'react';
import { ActionsMenu } from '@/ui/ActionsMenu/ActionsMenu';
import { cx } from '@/ui/cx';
import { WidgetView } from '../charts';
import type { CustomDashboardFilters, Widget, WidgetLayout } from '../model';
import type { BuilderAction } from './builder-reducer';
import { gridArea, KEYS_HELP_ID, widgetCellName, widgetHandleId } from './canvas-dom';
import type { Step } from './grid';
import { usePointerDrag } from './usePointerDrag';
import type { Pitch } from './usePointerDrag';
import { widgetName } from './widget-defaults';
import styles from './Builder.module.css';

const STEP_KEYS: Record<string, Step> = {
  ArrowLeft: 'left',
  ArrowRight: 'right',
  ArrowUp: 'up',
  ArrowDown: 'down',
};

export interface CanvasWidgetProps {
  widget: Widget;
  isSelected: boolean;
  problems: string[];
  filters: CustomDashboardFilters;
  dashboardName: string;
  pitch: () => Pitch;
  onPreview: (layout: WidgetLayout | null) => void;
  onAction: (action: BuilderAction) => void;
  onDeleted: () => void;
}

export function CanvasWidget({
  widget,
  isSelected,
  problems,
  filters,
  dashboardName,
  pitch,
  onPreview,
  onAction,
  onDeleted,
}: CanvasWidgetProps) {
  const { id, layout } = widget;
  const name = widgetName(widget);
  const place = (next: WidgetLayout) => onAction({ type: 'place', id, layout: next });
  const move = usePointerDrag({ layout, resize: false, pitch, onPreview, onDrop: place });
  const resize = usePointerDrag({ layout, resize: true, pitch, onPreview, onDrop: place });

  function remove() {
    onAction({ type: 'delete', id });
    onDeleted();
  }

  function handleKeyDown(event: KeyboardEvent<HTMLButtonElement>, resizeOnly: boolean) {
    const step = STEP_KEYS[event.key];
    const command = event.ctrlKey || event.metaKey;
    if (step) {
      event.preventDefault();
      onAction({ type: 'step', id, step, resize: resizeOnly || event.shiftKey });
    } else if (event.key === 'Delete' || event.key === 'Backspace') {
      event.preventDefault();
      remove();
    } else if (command && event.key.toLowerCase() === 'd') {
      event.preventDefault();
      onAction({ type: 'duplicate', id });
    } else if (event.key === 'Escape' && isSelected) {
      event.stopPropagation();
      onAction({ type: 'select', id: null });
    }
  }

  return (
    <div
      role="group"
      aria-label={widgetCellName(name, layout)}
      className={cx(styles.widget, isSelected && styles.selected)}
      style={gridArea(layout)}
    >
      <div className={styles.widgetHeader}>
        <button
          id={widgetHandleId(id)}
          type="button"
          className={styles.handle}
          aria-describedby={KEYS_HELP_ID}
          aria-pressed={isSelected}
          onFocus={() => !isSelected && onAction({ type: 'select', id })}
          onClick={() => onAction({ type: 'select', id })}
          onKeyDown={(event) => handleKeyDown(event, false)}
          {...move}
        >
          <GripVertical size={16} aria-hidden="true" />
          <span className={styles.handleText}>{name}</span>
        </button>
        <ActionsMenu
          label={`'${name}'`}
          items={[
            { label: 'Duplicate', onSelect: () => onAction({ type: 'duplicate', id }) },
            { label: 'Delete', tone: 'danger', separatorBefore: true, onSelect: remove },
          ]}
        />
      </div>
      {problems.length > 0 && (
        <ul className={styles.problems} role="list">
          {problems.map((problem) => (
            <li key={problem}>{problem}</li>
          ))}
        </ul>
      )}
      <div className={styles.preview}>
        <WidgetView
          widget={widget}
          filters={filters}
          mode="edit"
          dashboardName={dashboardName}
          viewerOwnsDashboard
        />
      </div>
      <button
        type="button"
        className={styles.resize}
        aria-label={`Resize '${name}'`}
        aria-describedby={KEYS_HELP_ID}
        onKeyDown={(event) => handleKeyDown(event, true)}
        {...resize}
      >
        <MoveDiagonal2 size={16} aria-hidden="true" />
      </button>
    </div>
  );
}
