import fc from 'fast-check';
import { describe, expect, it } from 'vitest';
import { DEFAULT_ADDRESS, readAddress, writeAddress } from './report-address';
import type { ReportAddress } from './report-address';

const id = (prefix: string) =>
  fc.integer({ min: 1, max: 9999 }).map((n) => `${prefix}-${String(n)}`);
const set = <T>(arb: fc.Arbitrary<T>) => fc.uniqueArray(arb, { maxLength: 4 });

const addressArb: fc.Arbitrary<ReportAddress> = fc.record({
  year: fc.option(fc.integer({ min: 2000, max: 2100 }), { nil: null }),
  plans: set(id('P')),
  planTypes: set(fc.constantFrom('operational', 'strategic', 'crossMarket')),
  teams: set(id('T')),
  owners: set(id('U')),
  activityTypes: set(id('AT')),
  activityStatuses: set(fc.constantFrom('created', 'approved', 'complete')),
  attributes: fc.uniqueArray(
    fc.record({
      attributeId: id('SA'),
      values: fc.uniqueArray(fc.constantFrom('SO-1', 'SO-2', 'none', 'yes'), { minLength: 1 }),
    }),
    { maxLength: 3, selector: (a) => a.attributeId },
  ),
  period: fc.constantFrom('month', 'quarter'),
  show: fc.constantFrom(null, 'plan', 'campaign', 'initiative', 'activity'),
  currency: fc.constantFrom(null, 'EUR', 'GBP'),
  tables: set(fc.constantFrom('statusCount', 'attr12Count')),
});

/** The address with every list sorted, as the canonical query writes it. */
function sorted(a: ReportAddress): ReportAddress {
  const s = <T extends string>(values: T[]) => [...values].sort();
  return {
    ...a,
    plans: s(a.plans),
    planTypes: s(a.planTypes),
    teams: s(a.teams),
    owners: s(a.owners),
    activityTypes: s(a.activityTypes),
    activityStatuses: s(a.activityStatuses),
    attributes: [...a.attributes]
      .map((attr) => ({ ...attr, values: s(attr.values) }))
      .sort((x, y) =>
        `${x.attributeId}:${s(x.values).join(',')}`.localeCompare(
          `${y.attributeId}:${s(y.values).join(',')}`,
        ),
      ),
    tables: s(a.tables),
  };
}

describe('report address (039 T020, FR-015, FR-024)', () => {
  it('reads back every address it writes', () => {
    fc.assert(
      fc.property(addressArb, (address) => {
        expect(sorted(readAddress(writeAddress(address)))).toEqual(sorted(address));
      }),
    );
  });

  it("writes the API's canonical query: data-model order, values sorted, defaults left out", () => {
    const query = writeAddress({
      ...DEFAULT_ADDRESS,
      year: 2026,
      plans: ['P-2', 'P-1'],
      attributes: [
        { attributeId: 'SA-2', values: ['SO-2', 'SO-1'] },
        { attributeId: 'SA-1', values: ['none'] },
      ],
      show: 'campaign',
      tables: ['statusCount', 'attr12Count'],
    }).toString();

    expect(query).toBe(
      'year=2026&plan=P-1&plan=P-2&attr=SA-1%3Anone&attr=SA-2%3ASO-1%2CSO-2&show=campaign' +
        '&table=attr12Count&table=statusCount',
    );
  });

  it('drops malformed values and keeps the rest', () => {
    const address = readAddress(
      new URLSearchParams(
        'year=20x6&plan=P-1&plan=bad id&planType=lost&period=year&show=lineItem&currency=eur&table=Bad-key&attr=SA-1',
      ),
    );

    expect(address).toEqual({ ...DEFAULT_ADDRESS, plans: ['P-1'] });
  });
});
