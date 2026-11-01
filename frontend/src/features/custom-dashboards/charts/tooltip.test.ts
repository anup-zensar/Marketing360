import { describe, expect, it } from 'vitest';
import { labelText, tooltipContent } from './tooltip';

const SCRIPT = '<img src=x onerror="alert(1)">';

describe('051 FR-038: tooltips', () => {
  it('lists each series of an axis with its formatted value', () => {
    const content = tooltipContent([
      {
        axisValueLabel: 'Event',
        seriesName: 'Budget',
        color: 'red',
        data: { value: 1, text: '$1.00' },
      },
      {
        axisValueLabel: 'Event',
        seriesName: 'Planned cost',
        color: 'blue',
        data: { value: 2, text: '$2.00' },
      },
    ]);
    expect(content.querySelector('p')?.textContent).toBe('Event');
    expect([...content.querySelectorAll('li')].map((li) => li.textContent)).toEqual([
      'Budget$1.00',
      'Planned cost$2.00',
    ]);
  });

  it('shows the lines an item carries, as scatter points and heatmap cells do', () => {
    const content = tooltipContent({
      name: 'Event',
      data: {
        lines: [
          ['Budget', '$1.00'],
          ['Activities', '3'],
        ],
      },
    });
    expect([...content.querySelectorAll('li')].map((li) => li.textContent)).toEqual([
      'Budget$1.00',
      'Activities3',
    ]);
  });

  it('never turns text from data into HTML', () => {
    const content = tooltipContent({ name: SCRIPT, seriesName: SCRIPT, data: { text: SCRIPT } });
    expect(content.querySelector('img')).toBeNull();
    expect(content.textContent).toBe(SCRIPT + SCRIPT + SCRIPT);
  });

  it('labels an item with its text, or its raw value without one', () => {
    expect(labelText({ data: { text: '$1.00' }, value: 1 })).toBe('$1.00');
    expect(labelText({ value: 3 })).toBe('3');
    expect(labelText(null)).toBe('');
  });
});
