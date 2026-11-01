import { joinWithAnd, midSentence, plural } from '@/format/text';
import type { Widget, WidgetData } from '../model';
import { widgetTypeInfo } from '../widget-types';
import { columnOf, keyId, measureColumn, uniqueKeys } from './chart-data';
import { formatValue } from './format';

/** The text summary assistive technology reads instead of the drawn chart. */
export function describeWidget(widget: Widget, data: WidgetData): string {
  const kind = kindOf(widget);
  const named = widget.title ? `${kind}: ${widget.title}.` : `${kind}.`;
  if (widget.type === 'text') {
    return named;
  }
  const measures = data.columns.filter((column) => /^m\d+$/.test(column.key));
  const measureNames = joinWithAnd(measures.map((column) => column.label));
  const dimension = columnOf(data, 'dimension');
  if (!dimension) {
    const total = data.totals[0];
    const value = total ? formatValue(total, measureColumn(data, 0), widget.display) : '—';
    return `${named} ${measureNames}: ${value}.`;
  }
  const series = columnOf(data, 'series');
  const by = `${measureNames} by ${midSentence(dimension.label)}`;
  const brokenDown = series ? `, broken down by ${midSentence(series.label)}` : '';
  return `${named} ${by}${brokenDown}; ${groupsText(data)}.`;
}

function kindOf(widget: Widget): string {
  const info = widgetTypeInfo(widget.type);
  return info.group === 'charts' ? `${info.label} chart` : info.label;
}

function groupsText(data: WidgetData): string {
  const shown = uniqueKeys(data.rows.map((row) => row.dimension)).filter(
    (key) => key?.kind !== 'other',
  ).length;
  const all = plural(data.groupCount, 'group');
  if (shown >= data.groupCount) {
    return all;
  }
  const rest = data.rows.some((row) => row.dimension?.kind === 'other')
    ? ', the rest as Other'
    : '';
  return `showing ${shown} of ${all}${rest}`;
}

export function shownGroupsNote(data: WidgetData): string | null {
  const text = groupsText(data);
  return text.startsWith('showing') ? `S${text.slice(1)}` : null;
}

export function rowKey(row: WidgetData['rows'][number]): string {
  return `${keyId(row.dimension)}|${keyId(row.series)}`;
}
