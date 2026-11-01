import type { ReportChoices, ReportFrame } from './api';
import type { ActivityStatus, PlanType, ReportAddress, Show } from './report-address';

export const PLAN_TYPE_WORDS: Record<PlanType, string> = {
  operational: 'Operational',
  strategic: 'Strategic',
  crossMarket: 'Cross-market',
};

export const STATUS_WORDS: Record<ActivityStatus, string> = {
  created: 'Created',
  submitted: 'Submitted',
  declined: 'Declined',
  approved: 'Approved',
  inProgress: 'In progress',
  complete: 'Complete',
};

export const SHOWS: Show[] = ['plan', 'campaign', 'initiative', 'activity'];

export const SHOW_WORDS: Record<Show, string> = {
  plan: 'Plan',
  campaign: 'Campaign',
  initiative: 'Initiative',
  activity: 'Activity',
};

export const LEVEL_WORDS: Record<string, string> = {
  plan: 'Plan',
  campaign: 'Campaign',
  initiative: 'Initiative',
  activity: 'Activity',
  lineItem: 'Line item',
};

export const NOT_APPLIED = 'Changes not applied yet';

function counted(count: number, one: string, many: string): string {
  return `${count.toLocaleString('en-GB')} ${count === 1 ? one : many}`;
}

/** "In view: 1 plan · 6 campaigns · 2 initiatives · 14 activities · 30 line items" (FR-007). */
export function inViewText(inView: ReportFrame['inView']): string {
  return `In view: ${[
    counted(inView.plans, 'plan', 'plans'),
    counted(inView.campaigns, 'campaign', 'campaigns'),
    counted(inView.initiatives, 'initiative', 'initiatives'),
    counted(inView.activities, 'activity', 'activities'),
    counted(inView.lineItems, 'line item', 'line items'),
  ].join(' · ')}`;
}

export function leftOutText(count: number): string {
  return count === 1
    ? '1 filter value is no longer available and was left out'
    : `${count.toLocaleString('en-GB')} filter values are no longer available and were left out`;
}

export function coverageText(plans: number, of: number): string | null {
  if (plans === of) {
    return null;
  }
  return plans === 0
    ? "You can't see the financial data of these plans"
    : `Covers ${plans.toLocaleString('en-GB')} of ${of.toLocaleString('en-GB')} plans`;
}

export interface Chip {
  key: string;
  label: string;
}

const named = (ids: string[], choices: { id: string; name: string }[]) =>
  ids.map((id) => choices.find((c) => c.id === id)?.name ?? id).join(', ');

/** One chip per restricted filter, naming its values ("Brand: Helix, Not filled in") (FR-014). */
export function chipsOf(address: ReportAddress, choices: ReportChoices): Chip[] {
  const chips: Chip[] = [];
  const add = (key: string, title: string, text: string) => {
    if (text !== '') {
      chips.push({ key, label: `${title}: ${text}` });
    }
  };
  add('plans', 'Plans', named(address.plans, choices.plans));
  add('planTypes', 'Plan type', address.planTypes.map((t) => PLAN_TYPE_WORDS[t]).join(', '));
  add('teams', 'Team', named(address.teams, choices.teams));
  add('owners', 'Owner', named(address.owners, choices.owners));
  add('activityTypes', 'Activity type', named(address.activityTypes, choices.activityTypes));
  add(
    'activityStatuses',
    'Activity status',
    address.activityStatuses.map((s) => STATUS_WORDS[s]).join(', '),
  );
  for (const attr of address.attributes) {
    const group = choices.attributeGroups.find((g) => g.attributeId === attr.attributeId);
    const labels = attr.values.map(
      (v) => group?.values.find((value) => value.id === v)?.label ?? v,
    );
    add(`attr:${attr.attributeId}`, group?.label ?? attr.attributeId, labels.join(', '));
  }
  return chips;
}

/** The address without the filter of the chip key. */
export function withoutChip(address: ReportAddress, key: string): ReportAddress {
  if (key.startsWith('attr:')) {
    const id = key.slice('attr:'.length);
    return { ...address, attributes: address.attributes.filter((a) => a.attributeId !== id) };
  }
  return { ...address, [key]: [] };
}
