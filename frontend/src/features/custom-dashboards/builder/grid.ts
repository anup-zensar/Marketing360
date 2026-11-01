import { GRID_COLUMNS, GRID_ROWS } from '../model';
import type { WidgetLayout } from '../model';

export const MIN_SIZE = 2;
export const MAX_SIZE = 12;

export interface Placed {
  id: string;
  layout: WidgetLayout;
}

export interface Cell {
  x: number;
  y: number;
}

export interface Size {
  w: number;
  h: number;
}

export type Step = 'left' | 'right' | 'up' | 'down';

export type Blocked = 'outside' | 'overlap';

const MOVES: Record<Step, Cell> = {
  left: { x: -1, y: 0 },
  right: { x: 1, y: 0 },
  up: { x: 0, y: -1 },
  down: { x: 0, y: 1 },
};

export function overlaps(a: WidgetLayout, b: WidgetLayout): boolean {
  return a.x < b.x + b.w && b.x < a.x + a.w && a.y < b.y + b.h && b.y < a.y + a.h;
}

export function isInside(layout: WidgetLayout): boolean {
  const { x, y, w, h } = layout;
  const whole = [x, y, w, h].every(Number.isInteger);
  const sized = w >= MIN_SIZE && w <= MAX_SIZE && h >= MIN_SIZE && h <= MAX_SIZE;
  return whole && sized && x >= 0 && y >= 0 && x + w <= GRID_COLUMNS && y + h <= GRID_ROWS;
}

export function blockedBy(
  layout: WidgetLayout,
  widgets: readonly Placed[],
  ignoreId?: string,
): Blocked | null {
  if (!isInside(layout)) {
    return 'outside';
  }
  const others = widgets.filter((widget) => widget.id !== ignoreId);
  return others.some((widget) => overlaps(widget.layout, layout)) ? 'overlap' : null;
}

export function fits(layout: WidgetLayout, widgets: readonly Placed[], ignoreId?: string): boolean {
  return blockedBy(layout, widgets, ignoreId) === null;
}

export function firstFreePlace(size: Size, widgets: readonly Placed[]): WidgetLayout | null {
  for (let y = 0; y + size.h <= GRID_ROWS; y += 1) {
    for (let x = 0; x + size.w <= GRID_COLUMNS; x += 1) {
      const layout = { x, y, ...size };
      if (fits(layout, widgets)) {
        return layout;
      }
    }
  }
  return null;
}

export function placeDrop(size: Size, cell: Cell, widgets: readonly Placed[]): WidgetLayout | null {
  const dropped = {
    x: clamp(cell.x, 0, GRID_COLUMNS - size.w),
    y: clamp(cell.y, 0, GRID_ROWS - size.h),
    ...size,
  };
  return fits(dropped, widgets) ? dropped : firstFreePlace(size, widgets);
}

export function stepped(layout: WidgetLayout, step: Step, resize: boolean): WidgetLayout {
  const { x, y } = MOVES[step];
  return resize
    ? { ...layout, w: layout.w + x, h: layout.h + y }
    : { ...layout, x: layout.x + x, y: layout.y + y };
}

export function shifted(layout: WidgetLayout, cells: Cell, resize: boolean): WidgetLayout {
  if (resize) {
    return {
      ...layout,
      w: clamp(layout.w + cells.x, MIN_SIZE, Math.min(MAX_SIZE, GRID_COLUMNS - layout.x)),
      h: clamp(layout.h + cells.y, MIN_SIZE, Math.min(MAX_SIZE, GRID_ROWS - layout.y)),
    };
  }
  return {
    ...layout,
    x: clamp(layout.x + cells.x, 0, GRID_COLUMNS - layout.w),
    y: clamp(layout.y + cells.y, 0, GRID_ROWS - layout.h),
  };
}

export function rowsUsed(widgets: readonly Placed[]): number {
  return Math.max(0, ...widgets.map((widget) => widget.layout.y + widget.layout.h));
}

function clamp(value: number, low: number, high: number): number {
  if (!Number.isFinite(value)) {
    return low;
  }
  return Math.min(Math.max(Math.round(value), low), high);
}
