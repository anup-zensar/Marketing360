import type { CSSProperties } from 'react';
import type { WidgetLayout, WidgetType } from '../model';
import { WIDGET_TYPES } from '../widget-types';

export const KEYS_HELP_ID = 'canvas-keys-help';
export const NAME_FIELD_ID = 'dashboard-name';
export const WIDGET_MIME = 'application/x-mapp-widget';

export const DETAIL_FIELD_IDS: Record<string, string> = {
  description: 'dashboard-description',
  'filters.dateTo': 'dashboard-filters-endDate',
  'filters.planIds': 'dashboard-filter-plans',
  'filters.teamIds': 'dashboard-filter-teams',
};

export function widgetHandleId(id: string): string {
  return `canvas-widget-${id}`;
}

export function gridArea(layout: WidgetLayout): CSSProperties {
  return {
    gridColumn: `${layout.x + 1} / span ${layout.w}`,
    gridRow: `${layout.y + 1} / span ${layout.h}`,
  };
}

export function draggedType(types: readonly string[]): WidgetType | null {
  const prefix = `${WIDGET_MIME}-`;
  const named = types.find((type) => type.startsWith(prefix))?.slice(prefix.length);
  return WIDGET_TYPES.find((info) => info.type.toLowerCase() === named)?.type ?? null;
}
