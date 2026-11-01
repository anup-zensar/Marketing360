import type { components } from '@/api/generated/schema';

type ApprovalsReport = components['schemas']['ApprovalsReport'];

const OUTCOMES = [
  'waiting',
  'approved',
  'declined',
  'recalled',
  'replaced',
  'autoApproved',
] as const;

/** Activities in 2026: eight rounds, one of them waiting since 14 March, 1.3 days on average. */
export function approvalsReport(overrides: Partial<ApprovalsReport> = {}): ApprovalsReport {
  const counts = [1, 2, 1, 2, 1, 1];
  const shares = [13, 25, 13, 25, 12, 12];
  return {
    statusMix: {
      recordType: 'activity',
      recordTypes: ['activity', 'purchaseRequest'],
      total: 8,
      waiting: { count: 1, oldestSince: '2026-03-14T09:00:00Z' },
      items: OUTCOMES.map((outcome, i) => ({
        outcome,
        count: counts[i] ?? 0,
        sharePercent: shares[i] ?? 0,
      })),
    },
    average: {
      days: '1.3',
      decisions: 3,
      approvers: [
        { person: { id: 'U-1001', name: 'Bob Ruiz' }, days: '1.5', decisions: 2 },
        { person: { id: 'U-1002', name: 'Erin Poe' }, days: '1.0', decisions: 1 },
      ],
    },
    grid: {
      items: [],
      total: 0,
      page: 1,
      pageSize: 50,
      choices: { recordTypes: [], outcomes: [], approvers: [] },
    },
    financeCovers: { plans: 1, of: 2 },
    ...overrides,
  };
}

export function noApprovalsReport(): ApprovalsReport {
  const empty = approvalsReport();
  return approvalsReport({
    statusMix: {
      ...empty.statusMix,
      recordTypes: [],
      total: 0,
      waiting: { count: 0, oldestSince: null },
      items: empty.statusMix.items.map((item) => ({ ...item, count: 0, sharePercent: 0 })),
    },
    average: { days: null, decisions: 0, approvers: [] },
    financeCovers: { plans: 0, of: 0 },
  });
}
