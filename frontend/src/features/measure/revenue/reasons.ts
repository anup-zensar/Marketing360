/** Why a figure is empty (FR-009, 049-C20): the API's stated reason, never inferred and never a sample value. */
const REASON_WORDS: Record<string, string> = {
  noGoal: 'No goal set.',
  notApproved: 'Not approved yet.',
  noSource: 'No active data source.',
  noData: 'No data has arrived yet.',
  attributionNotConfigured: 'Attribution is not configured yet.',
  noPreviousStage: '—',
};

export function reasonText(reason: string): string {
  return REASON_WORDS[reason] ?? '—';
}
