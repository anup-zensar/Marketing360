import type { GsapLinkReport, GsapPath, GsapWaitingRow } from './api';
import type { GsapTab } from './gsap-address';

export const TAB_LABELS: Record<GsapTab, string> = {
  swbs: 'SWBS not received',
  waiting: 'Waiting for GSAP',
  orders: 'Purchase orders',
};

/** Waiting in days, with hours under two days (044 FR-006): "1 hour", "30 hours", "3 days". */
export function waitingText(hours: number): string {
  if (hours < 48) {
    return hours === 1 ? '1 hour' : `${hours} hours`;
  }
  return `${Math.floor(hours / 24)} days`;
}

export function workingDaysText(days: number): string {
  return days === 1 ? '1 working day' : `${days} working days`;
}

/** Which plans the amounts cover (044 FR-004): nothing to say when they cover every plan in view. */
export function coverageText(finance: GsapLinkReport['finance']): string | undefined {
  if (!finance || finance.plans === finance.of) {
    return undefined;
  }
  if (finance.plans === 0) {
    return "You can't see financial data";
  }
  return `Covers ${finance.plans} of ${finance.of} plans`;
}

export const ROUTING_LABELS: Record<GsapWaitingRow['routing'], string> = {
  original: 'Original',
  modification: 'Modification',
};

export function pathText(path: GsapPath): string {
  return [path.plan, path.campaign, path.initiative, path.activity].filter(Boolean).join(' › ');
}
