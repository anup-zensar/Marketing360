import fc from 'fast-check';
import { describe, expect, it } from 'vitest';
import {
  DEFAULT_PARAMS,
  PAGE_SIZES,
  RECORD_TYPES,
  SORT_KEYS,
  readApprovalsParams,
  writeApprovalsParams,
} from './schemas';
import type { ApprovalsParams } from './schemas';

const word = fc.stringMatching(/^[A-Za-z0-9-]{1,12}$/);
const params: fc.Arbitrary<ApprovalsParams> = fc.record({
  statusType: fc.constantFrom(...RECORD_TYPES),
  recordTypes: fc.uniqueArray(word, { maxLength: 3 }),
  approvers: fc.uniqueArray(word, { maxLength: 3 }),
  outcomes: fc.uniqueArray(word, { maxLength: 3 }),
  q: fc.stringMatching(/^[a-z0-9]{0,20}$/),
  sort: fc.constantFrom(...SORT_KEYS),
  order: fc.constantFrom('asc' as const, 'desc' as const),
  page: fc.integer({ min: 1, max: 500 }),
  pageSize: fc.constantFrom(...PAGE_SIZES),
});

describe('the report address of Approvals (041 FR-005, FR-015)', () => {
  it('keeps every parameter through the address and the frame’s parameters beside them', () => {
    fc.assert(
      fc.property(params, (next) => {
        const written = writeApprovalsParams(new URLSearchParams('year=2026'), next);

        expect(readApprovalsParams(written)).toEqual(next);
        expect(written.get('year')).toBe('2026');
      }),
    );
  });

  it('leaves the defaults out of the address', () => {
    expect(writeApprovalsParams(new URLSearchParams(), DEFAULT_PARAMS).toString()).toBe('');
  });

  it('drops malformed values to the defaults', () => {
    const bad = new URLSearchParams(
      'statusType=x&sort=nope&order=up&page=0&pageSize=7&recordType=a%20b&outcome=%3Cscript%3E',
    );

    expect(readApprovalsParams(bad)).toEqual(DEFAULT_PARAMS);
  });
});
