import { z } from 'zod';

export const RECORD_TYPES = ['activity', 'purchaseRequest', 'modifiedPurchaseRequest'] as const;
export type RecordType = (typeof RECORD_TYPES)[number];

export interface ApprovalsParams {
  statusType: RecordType;
}

const statusTypeSchema = z.enum(RECORD_TYPES);

export function readApprovalsParams(params: URLSearchParams): ApprovalsParams {
  const parsed = statusTypeSchema.safeParse(params.get('statusType') ?? undefined);
  return { statusType: parsed.success ? parsed.data : 'activity' };
}

export function writeApprovalsParams(
  params: URLSearchParams,
  next: ApprovalsParams,
): URLSearchParams {
  const written = new URLSearchParams(params);
  if (next.statusType === 'activity') {
    written.delete('statusType');
  } else {
    written.set('statusType', next.statusType);
  }
  return written;
}
