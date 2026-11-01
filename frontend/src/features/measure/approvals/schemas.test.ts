import fc from 'fast-check';
import { describe, expect, it } from 'vitest';
import { RECORD_TYPES, readApprovalsParams, writeApprovalsParams } from './schemas';

describe('the report address of Approvals (041 FR-005)', () => {
  it('keeps the record type through the address and leaves the default out', () => {
    fc.assert(
      fc.property(fc.constantFrom(...RECORD_TYPES), (statusType) => {
        const written = writeApprovalsParams(new URLSearchParams('year=2026'), { statusType });

        expect(readApprovalsParams(written)).toEqual({ statusType });
        expect(written.get('year')).toBe('2026');
        expect(written.has('statusType')).toBe(statusType !== 'activity');
      }),
    );
  });

  it('drops a malformed record type to the default', () => {
    fc.assert(
      fc.property(
        fc.string().filter((text) => !RECORD_TYPES.some((type) => type === text)),
        (text) => {
          const params = new URLSearchParams({ statusType: text });

          expect(readApprovalsParams(params)).toEqual({ statusType: 'activity' });
        },
      ),
    );
  });
});
