import type { WidgetLayout } from '../model';
import type { Blocked } from './grid';

interface Placement {
  blocked: Blocked | null;
  resizing: boolean;
  layout: WidgetLayout;
}

export function placeMessage({ blocked, resizing, layout }: Placement): string {
  if (blocked === 'overlap') {
    return resizing
      ? "Can't resize: another widget is in the way."
      : "Can't move there: another widget is in the way.";
  }
  if (blocked === 'outside') {
    return resizing
      ? "Can't resize: a widget is 2 to 12 columns wide and 2 to 12 rows high, inside the grid."
      : "Can't move there: that's the edge of the grid.";
  }
  return resizing
    ? `Resized to ${layout.w} columns by ${layout.h} rows.`
    : `Moved to column ${layout.x + 1}, row ${layout.y + 1}.`;
}
