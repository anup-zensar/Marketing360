import {
  formatAmount,
  parseDecimalInput,
  parseMoneyInput,
  toCents,
  toHundredths,
} from '@/format/decimal';
import type { Model, ModelCreate, ModelUpdate } from './api';
import type { ModelValues } from './schemas';

export type ModelFormMode =
  { kind: 'new' } | { kind: 'duplicate'; model: Model } | { kind: 'edit'; model: Model };

export function startValues(mode: ModelFormMode): ModelValues {
  if (mode.kind === 'new') {
    return {
      name: '',
      inqToMqlRate: '',
      mqlToDealRate: '',
      averageDealSize: '',
      activityTypeIds: [],
      inqToMqlVelocityMonths: '0',
      mqlToDealVelocityMonths: '0',
      allowedGoalStages: [...ALL_GOAL_STAGES],
    };
  }
  const { model } = mode;
  return {
    name: mode.kind === 'duplicate' ? `${model.name} (copy)` : model.name,
    inqToMqlRate: model.inqToMqlRate,
    mqlToDealRate: model.mqlToDealRate,
    averageDealSize: formatAmount(model.averageDealSize.amount),
    // A new model can allow active types only.
    activityTypeIds: model.activityTypes
      .filter((type) => mode.kind === 'edit' || type.active)
      .map((type) => type.id),
    inqToMqlVelocityMonths: String(model.inqToMqlVelocityMonths),
    mqlToDealVelocityMonths: String(model.mqlToDealVelocityMonths),
    allowedGoalStages: [...model.allowedGoalStages],
  };
}

/** The stages a goal may be set against, in funnel order (047 FR-002). */
export const ALL_GOAL_STAGES = ['inq', 'mql', 'revenue'] as const;

/** A typed velocity as the API takes it: whole months, or NaN left for the API to refuse. */
function velocity(text: string): number {
  return Number.parseInt(text.trim(), 10);
}

export function createBody(values: ModelValues): ModelCreate {
  return {
    name: values.name,
    inqToMqlRate: rateText(values.inqToMqlRate),
    mqlToDealRate: rateText(values.mqlToDealRate),
    averageDealSize: dealSizeInput(values.averageDealSize),
    activityTypeIds: values.activityTypeIds,
    inqToMqlVelocityMonths: velocity(values.inqToMqlVelocityMonths),
    mqlToDealVelocityMonths: velocity(values.mqlToDealVelocityMonths),
    allowedGoalStages: values.allowedGoalStages,
  };
}

/** The fields people changed, with the version (D-203); null when nothing changed. */
export function changedFields(
  model: Model,
  start: ModelValues,
  values: ModelValues,
  version: number,
): ModelUpdate | null {
  const body: ModelUpdate = { version };
  if (values.name !== start.name) {
    body.name = values.name;
  }
  if (!sameRate(values.inqToMqlRate, model.inqToMqlRate)) {
    body.inqToMqlRate = rateText(values.inqToMqlRate);
  }
  if (!sameRate(values.mqlToDealRate, model.mqlToDealRate)) {
    body.mqlToDealRate = rateText(values.mqlToDealRate);
  }
  if (!sameAmount(values.averageDealSize, model.averageDealSize.amount)) {
    body.averageDealSize = dealSizeInput(values.averageDealSize);
  }
  if (!sameIds(values.activityTypeIds, start.activityTypeIds)) {
    body.activityTypeIds = values.activityTypeIds;
  }
  if (values.inqToMqlVelocityMonths !== start.inqToMqlVelocityMonths) {
    body.inqToMqlVelocityMonths = velocity(values.inqToMqlVelocityMonths);
  }
  if (values.mqlToDealVelocityMonths !== start.mqlToDealVelocityMonths) {
    body.mqlToDealVelocityMonths = velocity(values.mqlToDealVelocityMonths);
  }
  if (!sameIds(values.allowedGoalStages, start.allowedGoalStages)) {
    body.allowedGoalStages = values.allowedGoalStages;
  }
  return Object.keys(body).length > 1 ? body : null;
}

/** Changes that may affect plans and activities, checked with …/impact first (FR-011, FR-012). */
export function affectsOthers(body: ModelUpdate): boolean {
  return (
    body.inqToMqlRate !== undefined ||
    body.mqlToDealRate !== undefined ||
    body.averageDealSize !== undefined ||
    body.activityTypeIds !== undefined ||
    body.allowedGoalStages !== undefined
  );
}

function sameIds(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((id) => b.includes(id));
}

/** A typed rate as the API takes it: "2,5" stays as typed, so the API says what is wrong. */
function rateText(text: string): string {
  return parseDecimalInput(text, 2) ?? text.trim();
}

/** A typed amount as the API takes it; null when nothing is typed. */
function dealSizeInput(text: string): ModelCreate['averageDealSize'] {
  if (text.trim() === '') {
    return null;
  }
  return { amount: parseMoneyInput(text)?.amount ?? text.trim(), currency: 'USD' };
}

function sameRate(typed: string, saved: string): boolean {
  const rate = parseDecimalInput(typed, 2);
  return rate !== null && toHundredths(rate) === toHundredths(saved);
}

function sameAmount(typed: string, saved: string): boolean {
  const amount = parseMoneyInput(typed);
  return amount !== null && toCents(amount.amount) === toCents(saved);
}
