import type { DataField, DataSource, Widget, WidgetQuery } from '../model';
import { widgetTypeInfo } from '../widget-types';
import type { WidgetTypeInfo } from '../widget-types';
import { typeName } from './widget-defaults';

export type Found = [path: string, message: string];

type Filter = WidgetQuery['filters'][number];
type Measure = WidgetQuery['measures'][number];
type GroupBy = NonNullable<WidgetQuery['dimension']>;

const DECIMAL = /^-?[0-9]{1,15}(\.[0-9]{1,4})?$/;
const MONEY = /^-?[0-9]{1,15}(\.[0-9]{1,2})?$/;
const DATE = /^\d{4}-\d{2}-\d{2}$/;
const ENTER_NUMBER = 'Enter a number, such as 1250.5.';
const NO_VALUES = ['isEmpty', 'notEmpty'];
const COMPARING = ['gt', 'gte', 'lt', 'lte', 'between'];
const NEEDS_NUMBER: Partial<Record<Measure['aggregation'], string>> = {
  sum: 'Sum needs a number or money field.',
  avg: 'Average needs a number or money field.',
  min: 'Minimum needs a number, money or date field.',
  max: 'Maximum needs a number, money or date field.',
};
const ONE_MEASURE_BROKEN_DOWN = ['bar', 'line', 'area'];

export function tooLong(path: string, value: string, max: number): Found | null {
  return value.length > max ? [path, `Use at most ${max} characters.`] : null;
}

export function queryProblems(widget: Widget, sources: DataSource[] | undefined): Found[] {
  const { query } = widget;
  if (!query) {
    return [];
  }
  if (query.source === '') {
    return [['query.source', 'Choose a source.']];
  }
  const source = sources?.find((each) => each.key === query.source);
  if (sources && !source) {
    return [['query.source', "This data source isn't available. Choose another."]];
  }
  const info = widgetTypeInfo(widget.type);
  const fieldOf = (key: string | null) => source?.fields.find((field) => field.key === key);
  return [
    ...groupProblems('dimension', query.dimension, info.dimension, { info, fieldOf }),
    ...groupProblems('series', query.series, info.series, { info, fieldOf }),
    ...sameGroup(query),
    ...measureCount(query, info),
    ...query.measures.flatMap((measure, index) => measureProblems(measure, index, fieldOf)),
    ...filterCount(query),
    ...query.filters.flatMap((filter, index) => filterProblems(filter, index, fieldOf)),
    ...sortAndLimit(query),
  ];
}

export function displayProblems(widget: Widget): Found[] {
  const { axisTitles, referenceLine, gauge, kpi } = widget.display;
  const found: (Found | null)[] = [
    axisTitles ? tooLong('display.axisTitles.x', axisTitles.x, 40) : null,
    axisTitles ? tooLong('display.axisTitles.y', axisTitles.y, 40) : null,
    axisTitles ? tooLong('display.axisTitles.y2', axisTitles.y2, 40) : null,
    referenceLine ? notDecimal('display.referenceLine.value', referenceLine.value) : null,
    referenceLine ? tooLong('display.referenceLine.label', referenceLine.label, 40) : null,
    gauge ? notDecimal('display.gauge.min', gauge.min) : null,
    gauge?.max ? notDecimal('display.gauge.max', gauge.max) : null,
    gauge?.max && Number(gauge.max) <= Number(gauge.min)
      ? ['display.gauge.max', 'The maximum must be above the minimum.']
      : null,
    kpi?.comparison === 'target' && !kpi.target
      ? ['display.kpi.target', 'Enter the target.']
      : null,
    kpi?.target ? notDecimal('display.kpi.target', kpi.target) : null,
  ];
  return found.filter((each) => each !== null);
}

interface Context {
  info: WidgetTypeInfo;
  fieldOf: (key: string | null) => DataField | undefined;
}

function groupProblems(
  path: 'dimension' | 'series',
  group: GroupBy | null,
  need: WidgetTypeInfo['dimension'],
  { info, fieldOf }: Context,
): Found[] {
  const at = `query.${path}`;
  if (!group) {
    return need === 'required'
      ? [[at, path === 'dimension' ? 'Choose what to group by.' : 'Choose what to break down by.']]
      : [];
  }
  if (need === 'none') {
    return [
      [
        at,
        `A ${typeName(info.type).toLowerCase()} has no ${path === 'dimension' ? 'group' : 'breakdown'}.`,
      ],
    ];
  }
  const field = fieldOf(group.field);
  if (!field) {
    return [];
  }
  if (field.kind === 'date' && !group.dateGroup) {
    return [[`${at}.dateGroup`, 'Choose how to group the dates.']];
  }
  return field.groupable ? [] : [[at, `${field.label} can't be grouped.`]];
}

function sameGroup(query: WidgetQuery): Found[] {
  const { dimension, series } = query;
  const same =
    dimension &&
    series &&
    dimension.field === series.field &&
    dimension.dateGroup === series.dateGroup;
  return same ? [['query.series', 'Break down by another field than the group.']] : [];
}

function measureCount(query: WidgetQuery, info: WidgetTypeInfo): Found[] {
  const count = query.measures.length;
  const [min, max] = info.measures;
  const name = typeName(info.type).toLowerCase();
  if (count === 0) {
    return [['query.measures', 'Choose what to measure.']];
  }
  if (count < min) {
    return [['query.measures', `A ${name} needs at least ${min} measures.`]];
  }
  if (count > max) {
    const most = max === 1 ? 'one measure' : `at most ${max} measures`;
    return [['query.measures', `A ${name} shows ${most}.`]];
  }
  const brokenDown = query.series && count > 1 && ONE_MEASURE_BROKEN_DOWN.includes(info.type);
  return brokenDown ? [['query.measures', `Broken down, a ${name} shows one measure.`]] : [];
}

function measureProblems(measure: Measure, index: number, fieldOf: Context['fieldOf']): Found[] {
  const at = `query.measures[${index}]`;
  const label = tooLong(`${at}.label`, measure.label, 40);
  const labelProblems: Found[] = label ? [label] : [];
  if (measure.field === null) {
    return measure.aggregation === 'count'
      ? labelProblems
      : [[`${at}.field`, 'Choose the field to measure.'], ...labelProblems];
  }
  const field = fieldOf(measure.field);
  const refused = field && !field.aggregations.includes(measure.aggregation);
  const message = NEEDS_NUMBER[measure.aggregation] ?? `This can't be counted.`;
  return refused ? [[`${at}.field`, message], ...labelProblems] : labelProblems;
}

function filterCount(query: WidgetQuery): Found[] {
  return query.filters.length > 10 ? [['query.filters', 'Use at most 10 filters.']] : [];
}

function filterProblems(filter: Filter, index: number, fieldOf: Context['fieldOf']): Found[] {
  const at = `query.filters[${index}]`;
  if (filter.field === '') {
    return [[`${at}.field`, 'Choose a field to filter on.']];
  }
  const field = fieldOf(filter.field);
  if (field && !operatorFits(filter.operator, field)) {
    return [[`${at}.operator`, `This comparison can't be used on ${field.label}.`]];
  }
  const count = valueCountProblem(filter);
  if (count) {
    return [[`${at}.values`, count]];
  }
  const format = field && filter.values.map((value) => valueProblem(value, field)).find(Boolean);
  return format ? [[`${at}.values`, format]] : [];
}

function operatorFits(operator: Filter['operator'], field: DataField): boolean {
  if (operator === 'contains') {
    return field.kind === 'text';
  }
  return !COMPARING.includes(operator) || field.kind !== 'text';
}

function valueCountProblem({ operator, values }: Filter): string | null {
  if (NO_VALUES.includes(operator)) {
    return null;
  }
  const filled = values.filter((value) => value.trim() !== '').length;
  if (operator === 'between') {
    return filled === 2 && values.length === 2 ? null : 'Enter both values.';
  }
  if (operator === 'in') {
    return filled >= 1 && filled <= 50 && filled === values.length ? null : 'Enter 1 to 50 values.';
  }
  return filled === 1 && values.length === 1 ? null : 'Enter a value.';
}

function valueProblem(value: string, field: DataField): string | null {
  switch (field.kind) {
    case 'number':
      return DECIMAL.test(value) ? null : ENTER_NUMBER;
    case 'money':
      return MONEY.test(value) ? null : 'Enter an amount in US dollars, such as 1250.50.';
    case 'date':
      return DATE.test(value) ? null : 'Enter a date as YYYY-MM-DD.';
    case 'text':
      return value.length <= 200 ? null : 'Use at most 200 characters.';
    default:
      return unknownKind(field.kind);
  }
}

function sortAndLimit(query: WidgetQuery): Found[] {
  const found: Found[] = [];
  if (
    query.sort.by === 'measure' &&
    query.sort.measureIndex >= Math.max(1, query.measures.length)
  ) {
    found.push(['query.sort', 'Choose a measure to sort by.']);
  }
  if (!Number.isInteger(query.limit) || query.limit < 1 || query.limit > 500) {
    found.push(['query.limit', 'Show the top 1 to 500 groups.']);
  }
  return found;
}

function notDecimal(path: string, value: string): Found | null {
  return DECIMAL.test(value) ? null : [path, ENTER_NUMBER];
}

function unknownKind(kind: never): never {
  throw new Error(`Unknown field kind: ${String(kind)}`);
}
