import type { DataField, DataSource, WidgetQuery, WidgetType } from '../model';
import { defaultSort } from './widget-defaults';

type GroupBy = NonNullable<WidgetQuery['dimension']>;
type Filter = WidgetQuery['filters'][number];
type Measure = WidgetQuery['measures'][number];
export type Operator = Filter['operator'];
export type Aggregation = Measure['aggregation'];

export interface SourceChange {
  query: WidgetQuery;
  cleared: string | null;
}

export const OPERATOR_WORDS: Record<Operator, string> = {
  eq: 'is',
  ne: 'is not',
  in: 'is one of',
  contains: 'contains',
  gt: 'greater than',
  gte: 'at least',
  lt: 'less than',
  lte: 'at most',
  between: 'between',
  isEmpty: 'is empty',
  notEmpty: 'is not empty',
};

export const AGGREGATION_WORDS: Record<Aggregation, string> = {
  count: 'Count',
  countDistinct: 'Count distinct',
  sum: 'Sum',
  avg: 'Average',
  min: 'Minimum',
  max: 'Maximum',
};

const ALL_KINDS: Operator[] = ['eq', 'ne', 'in', 'isEmpty', 'notEmpty'];
const COMPARING: Operator[] = ['gt', 'gte', 'lt', 'lte', 'between'];

export function operatorsFor(field: DataField | undefined): Operator[] {
  if (!field) {
    return ALL_KINDS;
  }
  return field.kind === 'text'
    ? ['eq', 'ne', 'in', 'contains', 'isEmpty', 'notEmpty']
    : [...ALL_KINDS.slice(0, 3), ...COMPARING, 'isEmpty', 'notEmpty'];
}

export function valuesFor(operator: Operator, values: string[]): string[] {
  if (operator === 'isEmpty' || operator === 'notEmpty') {
    return [];
  }
  if (operator === 'between') {
    return [values[0] ?? '', values[1] ?? ''];
  }
  return operator === 'in' ? values : [values[0] ?? ''];
}

export function changeSource(query: WidgetQuery, source: DataSource): SourceChange {
  const fieldOf = (key: string | null) => source.fields.find((field) => field.key === key);
  const dimension = keptGroup(query.dimension, fieldOf);
  const series = keptGroup(query.series, fieldOf);
  const measures = query.measures.filter((measure) => measureFits(measure, fieldOf));
  const filters = query.filters.filter((filter) => filterFits(filter, fieldOf));
  const sortFits = query.sort.by === 'dimension' || query.sort.measureIndex < measures.length;
  const cleared = [
    query.dimension && !dimension ? 'Group by' : null,
    query.series && !series ? 'Break down by' : null,
    counted(query.measures.length - measures.length, 'measure'),
    counted(query.filters.length - filters.length, 'filter'),
  ].filter((part) => part !== null);
  return {
    query: {
      ...query,
      source: source.key,
      dimension,
      series,
      measures,
      filters,
      sort: sortFits ? query.sort : { by: 'measure', measureIndex: 0, direction: 'desc' },
    },
    cleared:
      cleared.length > 0
        ? `Cleared what doesn't fit ${source.label}: ${cleared.join(', ')}.`
        : null,
  };
}

export function withDimension(
  type: WidgetType,
  query: WidgetQuery,
  dimension: GroupBy | null,
  field: DataField | undefined,
): WidgetQuery {
  const isDate = field?.kind === 'date';
  const untouched = sameSort(query.sort, defaultSort(type, !isDate));
  return {
    ...query,
    dimension,
    sort: untouched ? defaultSort(type, isDate) : query.sort,
    ...(dimension === null && { groupRest: false }),
  };
}

export function groupOf(field: DataField | undefined): GroupBy | null {
  if (!field) {
    return null;
  }
  return { field: field.key, dateGroup: field.kind === 'date' ? 'month' : null };
}

export function aggregationsFor(field: DataField | undefined): Aggregation[] {
  return field ? field.aggregations : ['count'];
}

function sameSort(a: WidgetQuery['sort'], b: WidgetQuery['sort']): boolean {
  return a.by === b.by && a.direction === b.direction && a.measureIndex === b.measureIndex;
}

function keptGroup(
  group: GroupBy | null,
  fieldOf: (key: string | null) => DataField | undefined,
): GroupBy | null {
  const field = group && fieldOf(group.field);
  if (!group || !field?.groupable) {
    return null;
  }
  const isDate = field.kind === 'date';
  return isDate === (group.dateGroup !== null) ? group : null;
}

function measureFits(
  measure: Measure,
  fieldOf: (key: string | null) => DataField | undefined,
): boolean {
  if (measure.field === null) {
    return true;
  }
  return fieldOf(measure.field)?.aggregations.includes(measure.aggregation) ?? false;
}

function filterFits(
  filter: Filter,
  fieldOf: (key: string | null) => DataField | undefined,
): boolean {
  const field = fieldOf(filter.field);
  return field !== undefined && operatorsFor(field).includes(filter.operator);
}

function counted(count: number, noun: string): string | null {
  if (count <= 0) {
    return null;
  }
  return count === 1 ? `1 ${noun}` : `${count} ${noun}s`;
}
