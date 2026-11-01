import { describe, expect, it } from 'vitest';
import { buildOption } from './build-option';
import { niceMax } from './chart-data';
import {
  column,
  key,
  money,
  num,
  row,
  seriesData,
  THEME,
  twoMeasureData,
  widget,
  widgetData,
} from './widget-fixtures';

function at(value: unknown, ...path: (string | number)[]): unknown {
  return path.reduce<unknown>(
    (current, step) =>
      typeof current === 'object' && current !== null ? Reflect.get(current, step) : undefined,
    value,
  );
}

function call(value: unknown, ...args: unknown[]): unknown {
  if (typeof value !== 'function') {
    throw new Error('not a function');
  }
  return Reflect.apply(value, undefined, args);
}

describe('051 FR-024, FR-025: pie and donut', () => {
  it('draws one slice per group, labelled with SP&A One’s share by default', () => {
    const option = buildOption(widget('pie'), widgetData(), THEME);
    const series = at(option, 'series', 0);
    expect(series).toMatchObject({
      type: 'pie',
      radius: ['0%', '70%'],
      data: [
        { name: 'Event', value: 12500, text: '$12,500.00 (50.0%)' },
        { name: 'Digital', value: 7500 },
        { name: 'Print', value: 5000 },
      ],
    });
    expect(at(series, 'label', 'show')).toBe(true);
    expect(
      call(at(series, 'label', 'formatter'), { name: 'Event', data: { share: '50.0%' } }),
    ).toBe('Event: 50.0%');
  });

  it('leaves a hole in a donut and hides labels when asked', () => {
    const option = buildOption(widget('donut', { dataLabels: false }), widgetData(), THEME);
    expect(at(option, 'series', 0)).toMatchObject({
      radius: ['45%', '70%'],
      label: { show: false },
    });
  });
});

describe('051 FR-025: funnel and gauge', () => {
  it('keeps the API’s order in a funnel and turns it horizontal when chosen', () => {
    const option = buildOption(
      widget('funnel', { orientation: 'horizontal' }),
      widgetData(),
      THEME,
    );
    expect(at(option, 'series', 0)).toMatchObject({
      type: 'funnel',
      sort: 'none',
      orient: 'horizontal',
      data: [{ name: 'Event' }, { name: 'Digital' }, { name: 'Print' }],
    });
  });

  it('shows the total on a gauge up to a round maximum by default', () => {
    const data = widgetData({ columns: [column('m0', 'Sum of Planned cost')], rows: [] });
    const series = at(buildOption(widget('gauge'), data, THEME), 'series', 0);
    expect(series).toMatchObject({ type: 'gauge', min: 0, max: 25000 });
    expect(call(at(series, 'detail', 'formatter'))).toBe('$25,000.00');
  });

  it('uses the chosen gauge minimum and maximum', () => {
    const option = buildOption(
      widget('gauge', { gauge: { min: '1000', max: '40000' } }),
      widgetData(),
      THEME,
    );
    expect(at(option, 'series', 0)).toMatchObject({ min: 1000, max: 40000 });
  });

  it.each([
    [0, 1],
    [7, 10],
    [12, 20],
    [25000, 25000],
    [31000, 50000],
  ])('rounds %d up to %d', (value, expected) => {
    expect(niceMax(value)).toBe(expected);
  });
});

describe('051 FR-025: scatter and heatmap', () => {
  it('places each group at its two measures and names it', () => {
    const option = buildOption(widget('scatter'), twoMeasureData(), THEME);
    expect(at(option, 'series', 0, 'data', 0)).toMatchObject({
      name: 'Event',
      value: [12500, 5],
      lines: [
        ['Sum of Planned cost', '$12,500.00'],
        ['Count of records', '5'],
      ],
    });
    expect(at(option, 'xAxis', 'name')).toBe('Sum of Planned cost');
  });

  it('sizes points by a third measure, largest at the full size', () => {
    const data = widgetData({
      columns: [
        column('dimension', 'Plan', 'text'),
        column('m0', 'Budget'),
        column('m1', 'Planned cost'),
        column('m2', 'Activities', 'number'),
      ],
      rows: [
        row(key('A'), [money('1.00'), money('2.00'), num('4')]),
        row(key('B'), [money('3.00'), money('4.00'), num('1')]),
      ],
    });
    const points = at(buildOption(widget('scatter'), data, THEME), 'series', 0, 'data');
    expect(points).toMatchObject([{ symbolSize: 36 }, { symbolSize: 21 }]);
  });

  it('fills heatmap cells from five bins of the Sequential palette, each one colour', () => {
    const option = buildOption(widget('heatmap'), seriesData(), THEME);
    expect(at(option, 'series', 0, 'data', 0)).toMatchObject({ value: [0, 0, 10000] });
    const pieces = at(option, 'visualMap', 'pieces');
    expect(Array.isArray(pieces) && pieces.map((piece: unknown) => at(piece, 'color'))).toEqual(
      THEME.palettes.sequential,
    );
    expect(at(option, 'visualMap', 'pieces', 0, 'label')).toBe('$2.5K – $4K');
  });

  it('hides the heatmap legend when the legend is off', () => {
    const option = buildOption(
      widget('heatmap', { legend: { show: false, position: 'bottom' } }),
      seriesData(),
      THEME,
    );
    expect(at(option, 'visualMap', 'show')).toBe(false);
  });
});

describe('051 FR-025: treemap and radar', () => {
  it('draws one box per group, values in labels only when chosen', () => {
    const option = buildOption(widget('treemap', { dataLabels: true }), widgetData(), THEME);
    const series = at(option, 'series', 0);
    expect(at(series, 'data')).toMatchObject([
      { name: 'Event', value: 12500, itemStyle: { color: 'cat-1' } },
      { name: 'Digital', itemStyle: { color: 'cat-2' } },
      { name: 'Print' },
    ]);
    const label = at(series, 'label', 'formatter');
    expect(call(label, { name: 'Event', data: { text: '$12,500.00' } })).toBe('Event\n$12,500.00');
  });

  it('splits a treemap box by the breakdown without adding values up', () => {
    const option = buildOption(widget('treemap'), seriesData(), THEME);
    expect(at(option, 'series', 0, 'data', 0)).toEqual(
      expect.objectContaining({
        name: 'Event',
        children: [
          { name: 'Mobility', value: 10000, text: '$10,000.00' },
          { name: 'Lubricants', value: 2500, text: '$2,500.00' },
        ],
      }),
    );
    expect(at(option, 'series', 0, 'data', 0, 'value')).toBeUndefined();
  });

  it('draws the groups as spokes and one shape per breakdown group', () => {
    const option = buildOption(widget('radar'), seriesData(), THEME);
    expect(at(option, 'radar', 'indicator')).toEqual([
      { name: 'Event', max: 10000 },
      { name: 'Digital', max: 10000 },
    ]);
    expect(at(option, 'series', 0, 'data')).toMatchObject([
      { name: 'Mobility', value: [10000, 7500] },
      { name: 'Lubricants', value: [2500, 0] },
    ]);
  });
});
