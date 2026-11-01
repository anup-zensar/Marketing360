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
