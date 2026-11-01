import { describe, expect, it } from 'vitest';
import { describeWidget, shownGroupsNote } from './describe';
import { column, key, money, row, seriesData, widget, widgetData } from './widget-fixtures';

describe('051 FR-038: a text summary for every widget', () => {
  it('names the chart, its measures, its group and how many groups it shows', () => {
    expect(describeWidget(widget('bar'), widgetData())).toBe(
      'Bar chart: Planned cost by type. Sum of Planned cost by activity type; 3 groups.',
    );
  });

  it('names the breakdown', () => {
    expect(describeWidget(widget('combo', {}, { title: '' }), seriesData())).toBe(
      'Bar and line chart. Sum of Planned cost by activity type, broken down by plan; 2 groups.',
    );
  });

  it('says when only the top groups are shown, and that the rest are Other', () => {
    const data = widgetData({
      rows: [
        row(key('Event'), [money('3.00')]),
        row({ value: null, label: 'Other', kind: 'other' }, [money('2.00')]),
      ],
      groupCount: 37,
    });
    expect(describeWidget(widget('pie'), data)).toBe(
      'Pie chart: Planned cost by type. Sum of Planned cost by activity type; showing 1 of 37 groups, the rest as Other.',
    );
    expect(shownGroupsNote(data)).toBe('Showing 1 of 37 groups, the rest as Other');
    expect(shownGroupsNote(widgetData())).toBeNull();
  });

  it('reads out the value of a key figure or gauge', () => {
    const data = widgetData({ columns: [column('m0', 'Sum of Planned cost')], rows: [] });
    expect(describeWidget(widget('kpi', {}, { title: 'Planned cost' }), data)).toBe(
      'Key figure: Planned cost. Sum of Planned cost: $25,000.00.',
    );
  });

  it('names a text note', () => {
    expect(describeWidget(widget('text', {}, { title: 'About' }), widgetData())).toBe(
      'Text note: About.',
    );
  });
});
