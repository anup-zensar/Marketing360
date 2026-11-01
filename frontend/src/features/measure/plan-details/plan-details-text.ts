import type { PlanDetailsWidget } from './api';

export type Level = 'plan' | 'campaign' | 'initiative' | 'activity';

export const RECORDS: Record<Level, { one: string; many: string }> = {
  plan: { one: 'plan', many: 'Plans' },
  campaign: { one: 'campaign', many: 'Campaigns' },
  initiative: { one: 'initiative', many: 'Initiatives' },
  activity: { one: 'activity', many: 'Activities' },
};

/** The ⓘ tip naming the widget's figure and its definition (039 FR-020; measure-metrics §1). */
export function tipOf(widget: PlanDetailsWidget, level: Level): string {
  const records = RECORDS[level].many.toLowerCase();
  switch (widget.measure) {
    case 'plannedCost':
      return `The planned cost of the ${records} with each value, in the report's currency. A % split divides it by the shares; with several values and no shares, each value gets it all.`;
    case 'filledIn':
      return `How many ${records} have a value. Records where the attribute does not apply are left out of the per cent.`;
    case 'count':
      return `How many ${records} have each value. Not filled in counts the ${records} without one.`;
  }
}

export function countsInEachText(level: Level): string {
  return `${RECORDS[level].many} with several values count in each group; the total counts them once.`;
}

/** "8 activities where it does not apply are not counted" (043-C2). */
export function doesNotApplyText(count: number, level: Level): string | undefined {
  if (count === 0) {
    return undefined;
  }
  const word = count === 1 ? RECORDS[level].one : RECORDS[level].many.toLowerCase();
  return `${count.toLocaleString('en-GB')} ${word} where it does not apply ${count === 1 ? 'is' : 'are'} not counted`;
}

export function noAttributesText(level: Level): string {
  return `No ${RECORDS[level].one} attributes are set up in Settings.`;
}
