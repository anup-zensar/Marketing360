import { describe, expect, it } from 'vitest';
import { stableHash, widgetDataKeys } from './keys';
import type { WidgetDataRequest } from './model';
import { WIDGET_TYPES, widgetTypeInfo } from './widget-types';

const RULES_2_3 = [
  ['bar', 'required', 'optional', 1, 4],
  ['line', 'required', 'optional', 1, 4],
  ['area', 'required', 'optional', 1, 4],
  ['combo', 'required', 'none', 2, 4],
  ['pie', 'required', 'none', 1, 1],
  ['donut', 'required', 'none', 1, 1],
  ['scatter', 'required', 'optional', 2, 3],
  ['funnel', 'required', 'none', 1, 1],
  ['gauge', 'none', 'none', 1, 1],
  ['heatmap', 'required', 'required', 1, 1],
  ['treemap', 'required', 'optional', 1, 1],
  ['radar', 'required', 'optional', 1, 1],
  ['kpi', 'none', 'none', 1, 1],
  ['table', 'optional', 'optional', 1, 8],
  ['text', 'none', 'none', 0, 0],
] as const;

describe('widget types', () => {
  it.each(RULES_2_3)(
    'FR-015: %s needs a group %s, a breakdown %s and %i to %i measures',
    (type, dimension, series, min, max) => {
      const info = widgetTypeInfo(type);
      expect([info.dimension, info.series, info.measures]).toEqual([dimension, series, [min, max]]);
    },
  );

  it('FR-026: the palette lists every type once, charts first, in the order of the words on screen', () => {
    expect(WIDGET_TYPES.map((t) => t.label)).toEqual([
      'Bar',
      'Line',
      'Area',
      'Bar and line',
      'Pie',
      'Donut',
      'Scatter',
      'Funnel',
      'Gauge',
      'Heatmap',
      'Treemap',
      'Radar',
      'Key figure',
      'Table',
      'Text note',
    ]);
    expect(WIDGET_TYPES.map((t) => t.group)).toEqual([
      ...Array<string>(12).fill('charts'),
      'figures',
      'figures',
      'text',
    ]);
  });

  it('FR-013: a new widget takes its default size', () => {
    const sizes = Object.fromEntries(WIDGET_TYPES.map((t) => [t.type, `${t.size.w}x${t.size.h}`]));
    expect(sizes).toMatchObject({
      bar: '6x4',
      heatmap: '6x4',
      gauge: '4x4',
      kpi: '3x2',
      table: '6x5',
      text: '4x2',
    });
  });
});

describe('widget data keys', () => {
  const request: WidgetDataRequest = {
    type: 'kpi',
    query: {
      source: 'activities',
      dimension: null,
      series: null,
      measures: [{ aggregation: 'count', field: null, label: '' }],
      filters: [],
      sort: { by: 'measure', measureIndex: 0, direction: 'desc' },
      limit: 50,
      groupRest: false,
    },
    filters: null,
    comparison: null,
  };

  it('FR-036: equal requests share one key whatever their key order, and name their source', () => {
    const reordered: WidgetDataRequest = {
      comparison: null,
      filters: null,
      query: request.query,
      type: 'kpi',
    };
    expect(widgetDataKeys.one(reordered)).toEqual(widgetDataKeys.one(request));
    expect(widgetDataKeys.one(request).slice(0, 2)).toEqual(['widget-data', 'activities']);
    expect(stableHash({ b: 1, a: [2, { d: 3, c: 4 }] })).toBe('{"a":[2,{"c":4,"d":3}],"b":1}');
  });
});
