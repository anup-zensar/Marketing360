import { useRef, useState } from 'react';
import type { DragEvent } from 'react';
import { cx } from '@/ui/cx';
import { GRID_COLUMNS, GRID_ROWS } from '../model';
import type { WidgetLayout } from '../model';
import { widgetTypeInfo } from '../widget-types';
import type { BuilderAction, Draft } from './builder-reducer';
import { draggedType, gridArea, KEYS_HELP_ID } from './canvas-dom';
import { CanvasWidget } from './CanvasWidget';
import { fits, placeDrop, rowsUsed } from './grid';
import type { Cell } from './grid';
import type { Pitch } from './usePointerDrag';
import type { Problem } from './validate';
import { problemsOf } from './validate';
import styles from './Builder.module.css';

const ROW_HEIGHT = 80;
const GAP = 16;
const SPARE_ROWS = 4;
const MIN_ROWS = 8;

export interface CanvasProps {
  draft: Draft;
  selectedId: string | null;
  problems: readonly Problem[];
  announcement: { text: string; count: number };
  onAction: (action: BuilderAction) => void;
}

interface Preview {
  layout: WidgetLayout;
  id: string | null;
}

export function Canvas({ draft, selectedId, problems, announcement, onAction }: CanvasProps) {
  const region = useRef<HTMLElement>(null);
  const grid = useRef<HTMLDivElement>(null);
  const [preview, setPreview] = useState<Preview | null>(null);
  const rows = Math.min(GRID_ROWS, Math.max(MIN_ROWS, rowsUsed(draft.widgets) + SPARE_ROWS));

  function pitch(): Pitch {
    const width = grid.current?.getBoundingClientRect().width ?? 0;
    return { col: Math.max(1, (width + GAP) / GRID_COLUMNS), row: ROW_HEIGHT + GAP };
  }

  function cellAt(event: DragEvent<HTMLElement>): Cell {
    const rect = grid.current?.getBoundingClientRect();
    const { col, row } = pitch();
    return {
      x: Math.floor((event.clientX - (rect?.left ?? 0)) / col),
      y: Math.floor((event.clientY - (rect?.top ?? 0)) / row),
    };
  }

  function handleDragOver(event: DragEvent<HTMLElement>) {
    const type = draggedType(event.dataTransfer.types);
    if (!type) {
      return;
    }
    event.preventDefault();
    event.dataTransfer.dropEffect = 'copy';
    const layout = placeDrop(widgetTypeInfo(type).size, cellAt(event), draft.widgets);
    setPreview(layout ? { layout, id: null } : null);
  }

  function handleDrop(event: DragEvent<HTMLElement>) {
    const type = draggedType(event.dataTransfer.types);
    setPreview(null);
    if (type) {
      event.preventDefault();
      onAction({ type: 'add', widgetType: type, at: cellAt(event) });
    }
  }

  const blocked = preview !== null && !fits(preview.layout, draft.widgets, preview.id ?? undefined);

  return (
    <section
      ref={region}
      tabIndex={-1}
      aria-labelledby="canvas-title"
      className={styles.canvas}
      onDragOver={handleDragOver}
      onDragLeave={() => setPreview(null)}
      onDrop={handleDrop}
    >
      <h2 id="canvas-title" className="visually-hidden">
        Canvas
      </h2>
      <p id={KEYS_HELP_ID} className="visually-hidden">
        Arrow keys move the widget one cell; Shift with the arrow keys resizes it; Delete removes
        it.
      </p>
      <div
        ref={grid}
        className={styles.grid}
        style={{ gridTemplateRows: `repeat(${rows}, ${ROW_HEIGHT}px)` }}
      >
        {draft.widgets.length === 0 && (
          <p className={styles.empty}>
            Add a widget: click a type in the palette, or drag it here.
          </p>
        )}
        {draft.widgets.map((widget) => (
          <CanvasWidget
            key={widget.id}
            widget={widget}
            isSelected={widget.id === selectedId}
            problems={problemsOf(problems, widget.id).map((problem) => problem.message)}
            filters={draft.filters}
            dashboardName={draft.name}
            pitch={pitch}
            onPreview={(layout) => setPreview(layout && { layout, id: widget.id })}
            onAction={onAction}
            onDeleted={() => region.current?.focus()}
          />
        ))}
        {preview && (
          <div
            aria-hidden="true"
            className={cx(styles.drop, blocked && styles.dropBlocked)}
            style={gridArea(preview.layout)}
          />
        )}
      </div>
      <div role="status" className="visually-hidden">
        <span key={announcement.count}>{announcement.text}</span>
      </div>
    </section>
  );
}
