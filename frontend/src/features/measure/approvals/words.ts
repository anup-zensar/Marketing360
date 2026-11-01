import { RECORD_TYPES } from './schemas';
import type { RecordType } from './schemas';

export const NOTHING = '—';

export const RECORD_WORDS: Record<RecordType, string> = {
  activity: 'Activities',
  purchaseRequest: 'Purchase requests',
  modifiedPurchaseRequest: 'Modified purchase requests',
};

export const OUTCOME_WORDS: Record<string, string> = {
  waiting: 'Waiting',
  approved: 'Approved',
  declined: 'Declined',
  recalled: 'Recalled',
  replaced: 'Replaced by a new submission',
  autoApproved: 'Approved automatically',
  notNeeded: 'Not needed',
};

export function recordWord(type: string): string {
  const known = RECORD_TYPES.find((t) => t === type);
  return known ? RECORD_WORDS[known] : type;
}
