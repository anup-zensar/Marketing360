import { describe, expect, it } from 'vitest';
import { NO_FILTERS } from '../model';
import type { Widget, WidgetQuery } from '../model';
import { BAR_QUERY, BAR_WIDGET, NOTE_WIDGET, SOURCES } from './builder-fixtures';
import type { Draft } from './builder-reducer';
import { problemAt, problemsOf, validateDraft } from './validate';
import { newWidget } from './widget-defaults';

function draft(widgets: Widget[], overrides: Partial<Draft> = {}): Draft {
  return { name: 'Overview', description: '', filters: NO_FILTERS, widgets, ...overrides };
}

function bar(query: Partial<WidgetQuery>, extra: Partial<Widget> = {}): Widget {
  return { ...BAR_WIDGET, query: { ...BAR_QUERY, ...query }, ...extra };
}

function messages(widgets: Widget[], overrides: Partial<Draft> = {}) {
  return validateDraft(draft(widgets, overrides), SOURCES).map(({ field, message }) => [
    field,
    message,
  ]);
}

describe('051 FR-033, FR-008: the dashboard checked on Save, every problem at once', () => {
  it('finds nothing wrong with a complete dashboard', () => {
    expect(validateDraft(draft([BAR_WIDGET, NOTE_WIDGET]), SOURCES)).toEqual([]);
  });

  it('lists the name, the description, the dates and a widget together', () => {
    const incomplete = newWidget('pie', 'w3', { x: 0, y: 4, w: 6, h: 4 });
    expect(
      messages([incomplete], {
        name: '  ',
        description: 'x'.repeat(501),
        filters: { ...NO_FILTERS, dateFrom: '2026-07-01', dateTo: '2026-06-30' },
      }),
    ).toEqual([
      ['name', 'Enter a name.'],
      ['description', 'Use at most 500 characters.'],
      ['filters.dateTo', 'The end date must be on or after the start date.'],
      ['widgets[0].query.source', 'Choose a source.'],
    ]);
  });

  it('refuses a name over 80 characters', () => {
    expect(messages([], { name: 'x'.repeat(81) })).toEqual([
      ['name', 'Use at most 80 characters.'],
    ]);
  });

  it('ties widget problems to their widget, for the canvas', () => {
    const problems = validateDraft(draft([bar({ measures: [] }), NOTE_WIDGET]), SOURCES);
    expect(problemsOf(problems, 'w1').map((problem) => problem.message)).toEqual([
      'Choose what to measure.',
    ]);
    expect(problemsOf(problems, 'w2')).toEqual([]);
    expect(problemAt(problems, 'widgets[0].query.measures')).toBe('Choose what to measure.');
  });
});

describe('051 FR-015: what each widget type needs', () => {
  it.each([
    ['bar without a group', bar({ dimension: null }), 'Choose what to group by.'],
    [
      'heatmap without a breakdown',
      { ...bar({}), type: 'heatmap' as const },
      'Choose what to break down by.',
    ],
    [
      'pie with two measures',
      {
        ...bar({
          measures: [...BAR_QUERY.measures, { aggregation: 'count', field: null, label: '' }],
        }),
        type: 'pie' as const,
      },
      'A pie chart shows one measure.',
    ],
    [
      'combo with one measure',
      { ...bar({}), type: 'combo' as const },
      'A bar and line chart needs at least 2 measures.',
    ],
    ['key figure with a group', { ...bar({}), type: 'kpi' as const }, 'A key figure has no group.'],
    [
      'bar broken down with two measures',
      bar({
        series: { field: 'status', dateGroup: null },
        measures: [...BAR_QUERY.measures, { aggregation: 'count', field: null, label: '' }],
      }),
      'Broken down, a bar chart shows one measure.',
    ],
  ])('%s', (_case, widget, message) => {
    expect(validateDraft(draft([widget]), SOURCES).map((problem) => problem.message)).toContain(
      message,
    );
  });

  it('needs a date group for a date field and another field for the breakdown', () => {
    expect(
      messages([
        bar({
          dimension: { field: 'startDate', dateGroup: null },
          series: { field: 'startDate', dateGroup: null },
        }),
      ]),
    ).toEqual([
      ['widgets[0].query.dimension.dateGroup', 'Choose how to group the dates.'],
      ['widgets[0].query.series.dateGroup', 'Choose how to group the dates.'],
      ['widgets[0].query.series', 'Break down by another field than the group.'],
    ]);
  });

  it('refuses a sum of a text field and a field missing for an aggregation', () => {
    expect(
      messages([
        bar({
          measures: [
            { aggregation: 'sum', field: 'status', label: '' },
            { aggregation: 'avg', field: null, label: 'x'.repeat(41) },
          ],
        }),
      ]),
    ).toEqual([
      ['widgets[0].query.measures[0].field', 'Sum needs a number or money field.'],
      ['widgets[0].query.measures[1].field', 'Choose the field to measure.'],
      ['widgets[0].query.measures[1].label', 'Use at most 40 characters.'],
    ]);
  });

  it('refuses a source that is not in the data catalogue', () => {
    expect(messages([bar({ source: 'DS-9999' })])).toEqual([
      ['widgets[0].query.source', "This data source isn't available. Choose another."],
    ]);
  });

  it('checks what the API checks when the data catalogue has not loaded yet', () => {
    const problems = validateDraft(draft([bar({ source: 'DS-9999', measures: [] })]), undefined);
    expect(problems.map((problem) => problem.message)).toEqual(['Choose what to measure.']);
  });
});

describe('051 FR-019, FR-020: filters, sort and top N', () => {
  it.each([
    [{ field: '', operator: 'eq', values: ['x'] }, 'field', 'Choose a field to filter on.'],
    [
      { field: 'status', operator: 'gt', values: ['x'] },
      'operator',
      "This comparison can't be used on Status.",
    ],
    [{ field: 'status', operator: 'eq', values: [] }, 'values', 'Enter a value.'],
    [{ field: 'plannedCost', operator: 'between', values: ['1'] }, 'values', 'Enter both values.'],
    [{ field: 'status', operator: 'in', values: [] }, 'values', 'Enter 1 to 50 values.'],
    [
      { field: 'plannedCost', operator: 'gte', values: ['1.234'] },
      'values',
      'Enter an amount in US dollars, such as 1250.50.',
    ],
    [
      { field: 'startDate', operator: 'lt', values: ['1/7/2026'] },
      'values',
      'Enter a date as YYYY-MM-DD.',
    ],
  ] as const)('filter %o: %s', (filter, part, message) => {
    expect(messages([bar({ filters: [{ ...filter, values: [...filter.values] }] })])).toEqual([
      [`widgets[0].query.filters[0].${part}`, message],
    ]);
  });

  it('takes is empty without a value and is one of with several', () => {
    const filters: WidgetQuery['filters'] = [
      { field: 'status', operator: 'isEmpty', values: [] },
      { field: 'status', operator: 'in', values: ['Planned', 'Done'] },
    ];
    expect(messages([bar({ filters })])).toEqual([]);
  });

  it('needs the measure it sorts by and a top N of 1 to 500', () => {
    expect(
      messages([bar({ sort: { by: 'measure', measureIndex: 2, direction: 'asc' }, limit: 501 })]),
    ).toEqual([
      ['widgets[0].query.sort', 'Choose a measure to sort by.'],
      ['widgets[0].query.limit', 'Show the top 1 to 500 groups.'],
    ]);
  });
});

describe('051 FR-013, FR-025: layout and display options', () => {
  it('names the widget another one overlaps', () => {
    const over = { ...NOTE_WIDGET, layout: { x: 4, y: 0, w: 4, h: 2 } };
    expect(messages([BAR_WIDGET, over])).toContainEqual([
      'widgets[1].layout',
      "This widget overlaps 'Planned cost by type'.",
    ]);
  });

  it('checks the target, gauge range, reference line and lengths', () => {
    const display: Widget['display'] = {
      kpi: { comparison: 'target', target: null, higherIsBetter: true },
      gauge: { min: '10', max: '5' },
      referenceLine: { value: 'ten', label: '' },
      axisTitles: { x: 'x'.repeat(41), y: '', y2: '' },
    };
    expect(messages([bar({}, { display, title: 'x'.repeat(81) })])).toEqual([
      ['widgets[0].title', 'Use at most 80 characters.'],
      ['widgets[0].display.axisTitles.x', 'Use at most 40 characters.'],
      ['widgets[0].display.referenceLine.value', 'Enter a number, such as 1250.5.'],
      ['widgets[0].display.gauge.max', 'The maximum must be above the minimum.'],
      ['widgets[0].display.kpi.target', 'Enter the target.'],
    ]);
  });
});
