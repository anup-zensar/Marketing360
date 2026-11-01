import styles from './charts.module.css';

/** Extra fields the option builders put on each data item, read back by tooltips and labels. */
export interface ItemText {
  text?: string;
  lines?: [label: string, text: string][];
}

/** Text from data never becomes HTML: tooltips are DOM nodes filled with textContent (ADR-0020). */
export function tooltipContent(params: unknown): HTMLElement {
  const items = (Array.isArray(params) ? params : [params]).filter(isRecord);
  const box = element('div', styles.tooltip);
  const first = items[0];
  const title = first ? stringOf(first.axisValueLabel) || stringOf(first.name) : '';
  if (title) {
    box.append(element('p', styles.tooltipTitle, title));
  }
  const list = element('ul', styles.tooltipList);
  for (const item of items) {
    for (const [label, text] of linesOf(item)) {
      const row = element('li', styles.tooltipRow);
      const marker = element('span', styles.tooltipMarker);
      marker.style.backgroundColor = stringOf(item.color);
      row.append(marker, element('span', styles.tooltipLabel, label), element('span', '', text));
      list.append(row);
    }
  }
  box.append(list);
  return box;
}

/** The text a data label shows: the item's formatted value. */
export function labelText(params: unknown): string {
  if (!isRecord(params)) {
    return '';
  }
  const data = params.data;
  if (isRecord(data) && typeof data.text === 'string') {
    return data.text;
  }
  return stringOf(params.value);
}

function linesOf(item: Record<string, unknown>): [string, string][] {
  const data = item.data;
  if (isRecord(data) && Array.isArray(data.lines)) {
    return data.lines.filter(isLine);
  }
  return [[stringOf(item.seriesName), labelText(item)]];
}

function isLine(value: unknown): value is [string, string] {
  return (
    Array.isArray(value) &&
    value.length === 2 &&
    typeof value[0] === 'string' &&
    typeof value[1] === 'string'
  );
}

function element<K extends keyof HTMLElementTagNameMap>(
  tag: K,
  className: string | undefined,
  text?: string,
): HTMLElementTagNameMap[K] {
  const node = document.createElement(tag);
  if (className) {
    node.className = className;
  }
  if (text !== undefined) {
    node.textContent = text;
  }
  return node;
}

function stringOf(value: unknown): string {
  return typeof value === 'string' || typeof value === 'number' ? String(value) : '';
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}
