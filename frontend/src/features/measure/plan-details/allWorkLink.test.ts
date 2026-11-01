import { describe, expect, it } from 'vitest';
import { DEFAULT_ADDRESS } from '../shared/report-address';
import { allWorkLink } from './allWorkLink';

describe('allWorkLink (043 FR-014)', () => {
  const address = {
    ...DEFAULT_ADDRESS,
    year: 2026,
    teams: ['T-1001'],
    activityTypes: ['AT-1001'],
    attributes: [
      { attributeId: 'SA-12', values: ['GB01'] },
      { attributeId: 'SA-13', values: ['yes'] },
    ],
    period: 'quarter' as const,
    currency: 'EUR',
  };

  it("opens All work with the report's filters and the value in place of its attribute's", () => {
    const url = new URL(allWorkLink(address, 'attr', 'none', 'SA-12'), 'https://mapp.test');

    expect(url.pathname).toBe('/all-work');
    expect(url.searchParams.getAll('attr')).toEqual(['SA-13:yes', 'SA-12:none']);
    expect(url.searchParams.get('team')).toBe('T-1001');
    expect(url.searchParams.get('period')).toBeNull();
    expect(url.searchParams.get('currency')).toBeNull();
  });

  it('replaces the activity type or status filter with the value', () => {
    const url = new URL(allWorkLink(address, 'activityType', 'AT-2002'), 'https://mapp.test');

    expect(url.searchParams.getAll('activityType')).toEqual(['AT-2002']);
    expect(url.searchParams.getAll('attr')).toEqual(['SA-12:GB01', 'SA-13:yes']);
  });
});
