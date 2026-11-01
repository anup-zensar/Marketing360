import { describe, expect, it } from 'vitest';
import type { WidgetDisplay } from '../model';
import { buildOption } from './build-option';
import { tooltipContent } from './tooltip';
import {
  column,
  key,
  money,
  row,
  seriesData,
  THEME,
  twoMeasureData,
  widget,
  widgetData,
} from './widget-fixtures';

function at(value: unknown, ...path: (string | number)[]): unknown {
  let current: unknown = value;
  for (const step of path) {
    if (typeof current !== 'object' || current === null) {
      return undefined;
    }
    current = Reflect.get(current, step);
  }
  return current;
}

function call(value: unknown, ...args: unknown[]): unknown {
  if (typeof value !== 'function') {
    throw new Error('not a function');
  }
  return Reflect.apply(value, undefined, args);
}

describe('051 FR-024: colours and patterns', () => {
  it('uses the Categorical palette by default and the chosen palette otherwise', () => {
    expect(buildOption(widget('bar'), widgetData(), THEME)).toMatchObject({
      color: THEME.palettes.categorical,
    });
    expect(buildOption(widget('line', { palette: 'planning' }), widgetData(), THEME)).toMatchObject(
      { color: THEME.palettes.planning },
    );
  });

  it.each([
    'bar',
    'line',
    'area',
    'combo',
    'pie',
    'donut',
    'scatter',
    'funnel',
    'treemap',
  ] as const)(
    'adds patterns to %s so colour is never the only carrier, and confines the tooltip',
    (type) => {
      const option = buildOption(widget(type), twoMeasureData(), THEME);
      expect(option).toMatchObject({
        aria: { enabled: true, decal: { show: true } },
        tooltip: { confine: true, formatter: tooltipContent },
      });
    },
  );
});

describe('051 FR-025: bar, line and area', () => {
  it('draws one category per group and writes values as the API answered them', () => {
    const option = buildOption(widget('bar'), widgetData(), THEME);
    expect(at(option, 'xAxis')).toMatchObject({
      type: 'category',
      data: ['Event', 'Digital', 'Print'],
    });
    expect(at(option, 'series', 0)).toMatchObject({
      type: 'bar',
      name: 'Sum of Planned cost',
      data: [
        { value: 12500, text: '$12,500.00' },
        { value: 7500, text: '$7,500.00' },
        { value: 5000, text: '$5,000.00' },
      ],
    });
  });

  it('draws one series per breakdown group, leaving gaps where a group has no value', () => {
    const option = buildOption(widget('bar'), seriesData(), THEME);
    expect(at(option, 'series', 0, 'name')).toBe('Mobility');
    expect(at(option, 'series', 1)).toMatchObject({
      name: 'Lubricants',
      data: [{ value: 2500 }, null],
    });
  });

  it.each([
    ['none', undefined],
    ['stacked', 'total'],
  ] as const)('stacking %s', (stack, expected) => {
    const option = buildOption(widget('bar', { stack }), seriesData(), THEME);
    expect(at(option, 'series', 0, 'stack')).toBe(expected);
  });

  it('stacks to 100 % from the shares SP&A One answered', () => {
    const option = buildOption(widget('bar', { stack: 'percent' }), seriesData(), THEME);
    expect(at(option, 'yAxis', 'max')).toBe(100);
    expect(at(option, 'series', 0, 'data')).toMatchObject([
      { value: 80, text: '80.0% ($10,000.00)' },
      { value: 100 },
    ]);
  });

  it('stacks nothing without a breakdown, and stacks areas by default', () => {
    expect(
      at(
        buildOption(widget('bar', { stack: 'percent' }), widgetData(), THEME),
        'series',
        0,
        'stack',
      ),
    ).toBeUndefined();
    const area = buildOption(widget('area'), seriesData(), THEME);
    expect(at(area, 'series', 0)).toMatchObject({ type: 'line', stack: 'total', areaStyle: {} });
  });

  it('turns bars horizontal with the first group on top, keeping each axis title on its side', () => {
    const option = buildOption(
      widget('bar', { orientation: 'horizontal', axisTitles: { x: 'Cost', y: 'Type', y2: '' } }),
      widgetData(),
      THEME,
    );
    expect(at(option, 'xAxis')).toMatchObject({ type: 'value', name: 'Cost' });
    expect(at(option, 'yAxis')).toMatchObject({ type: 'category', inverse: true, name: 'Type' });
  });

  it('shows the legend where chosen, or hides it', () => {
    const right = buildOption(
      widget('bar', { legend: { show: true, position: 'right' } }),
      widgetData(),
      THEME,
    );
    expect(at(right, 'legend')).toMatchObject({ show: true, orient: 'vertical', right: 0 });
    const hidden = buildOption(
      widget('bar', { legend: { show: false, position: 'bottom' } }),
      widgetData(),
      THEME,
    );
    expect(at(hidden, 'legend', 'show')).toBe(false);
  });

  it('shows data labels with the formatted value', () => {
    const option = buildOption(widget('bar', { dataLabels: true }), widgetData(), THEME);
    const label = at(option, 'series', 0, 'label');
    expect(at(label, 'show')).toBe(true);
    expect(call(at(label, 'formatter'), { data: { value: 1, text: '$1.00' } })).toBe('$1.00');
  });

  it('hides gridlines and smooths lines when chosen', () => {
    const option = buildOption(
      widget('line', { gridlines: false, smooth: true }),
      widgetData(),
      THEME,
    );
    expect(at(option, 'yAxis', 'splitLine', 'show')).toBe(false);
    expect(at(option, 'series', 0, 'smooth')).toBe(true);
  });

  it.each([
    [{}, '$12,500.00', '$12.5K'],
    [{ numberFormat: 'compact', decimals: 1 }, '12.5K', '12.5K'],
    [{ numberFormat: 'number', decimals: 0 }, '12,500', '12.5K'],
    [{ numberFormat: 'percent', decimals: 1 }, '12,500.0%', '12.5K%'],
  ] as const)('number format %o writes %s and axis ticks %s', (display, text, tick) => {
    const option = buildOption(widget('bar', display), widgetData(), THEME);
    expect(at(option, 'series', 0, 'data', 0, 'text')).toBe(text);
    expect(call(at(option, 'yAxis', 'axisLabel', 'formatter'), 12500)).toBe(tick);
  });

  it('adds the zoom slider along the categories', () => {
    expect(buildOption(widget('line', { zoom: true }), widgetData(), THEME)).toMatchObject({
      dataZoom: [{ type: 'slider', xAxisIndex: 0 }],
    });
    expect(
      buildOption(widget('bar', { zoom: true, orientation: 'horizontal' }), widgetData(), THEME),
    ).toMatchObject({ dataZoom: [{ type: 'slider', yAxisIndex: 0 }] });
  });

  it('draws the reference line at its value with its label', () => {
    const option = buildOption(
      widget('bar', { referenceLine: { value: '10000', label: 'Target' } }),
      widgetData(),
      THEME,
    );
    const line = at(option, 'series', 0, 'markLine');
    expect(at(line, 'data')).toEqual([{ yAxis: 10000 }]);
    expect(call(at(line, 'label', 'formatter'))).toBe('Target');
  });

  it('writes date buckets in the person’s date format', () => {
    const data = widgetData({
      columns: [column('dimension', 'Month', 'date', 'month'), column('m0', 'Sum of Planned cost')],
      rows: [row(key('2026-03-01'), [money('1.00')]), row(key('2026-04-01'), [money('2.00')])],
    });
    expect(at(buildOption(widget('line'), data, THEME), 'xAxis', 'data')).toEqual([
      'Mar 2026',
      'Apr 2026',
    ]);
    expect(at(buildOption(widget('line'), data, THEME, 'YYYY-MM-DD'), 'xAxis', 'data')).toEqual([
      '2026-03',
      '2026-04',
    ]);
  });
});

describe('051 FR-025: bar and line', () => {
  it('draws the first measure as bars on the left and the others as lines on the right by default', () => {
    const option = buildOption(
      widget('combo', { axisTitles: { x: '', y: 'Cost', y2: 'Activities' } }),
      twoMeasureData(),
      THEME,
    );
    expect(at(option, 'series')).toMatchObject([
      { type: 'bar', yAxisIndex: 0, name: 'Sum of Planned cost' },
      { type: 'line', yAxisIndex: 1, name: 'Count of records' },
    ]);
    expect(at(option, 'yAxis')).toMatchObject([{ name: 'Cost' }, { name: 'Activities' }]);
  });

  it('keeps one value axis when every measure is on the left', () => {
    const display: WidgetDisplay = {
      measureStyles: [
        { as: 'bar', axis: 'left' },
        { as: 'bar', axis: 'left' },
      ],
    };
    const option = buildOption(widget('combo', display), twoMeasureData(), THEME);
    expect(at(option, 'yAxis')).toHaveLength(1);
    expect(at(option, 'series')).toMatchObject([
      { type: 'bar', yAxisIndex: 0 },
      { type: 'bar', yAxisIndex: 0 },
    ]);
  });
});

describe('051 FR-024, FR-025: kinds drawn without an option', () => {
  it.each(['kpi', 'table', 'text'] as const)('%s has no chart option', (type) => {
    expect(buildOption(widget(type), widgetData(), THEME)).toEqual({});
  });
});
