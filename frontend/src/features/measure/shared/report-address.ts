import { z } from 'zod';
import type { operations } from '@/api/generated/schema';

/** The query every report operation takes: the report filter of 039 data-model §2. */
export type ReportQuery = Omit<
  NonNullable<operations['getReportFrame']['parameters']['query']>,
  'report'
>;

export type PlanType = 'operational' | 'strategic' | 'crossMarket';
export type ActivityStatus =
  'created' | 'submitted' | 'declined' | 'approved' | 'inProgress' | 'complete';
export type Period = 'month' | 'quarter';
export type Show = 'plan' | 'campaign' | 'initiative' | 'activity';

export interface AttributeChoice {
  attributeId: string;
  values: string[];
}

/** A report's state in the page address (FR-015): an empty list restricts nothing. */
export interface ReportAddress {
  year: number | null;
  plans: string[];
  planTypes: PlanType[];
  teams: string[];
  owners: string[];
  activityTypes: string[];
  activityStatuses: ActivityStatus[];
  attributes: AttributeChoice[];
  period: Period;
  show: Show | null;
  currency: string | null;
  tables: string[];
}

/** The parameters the frame owns; any other belongs to the report. */
export const FRAME_PARAMS: readonly string[] = [
  'year',
  'plan',
  'planType',
  'team',
  'owner',
  'activityType',
  'activityStatus',
  'attr',
  'period',
  'show',
  'currency',
  'table',
];

export const DEFAULT_ADDRESS: ReportAddress = {
  year: null,
  plans: [],
  planTypes: [],
  teams: [],
  owners: [],
  activityTypes: [],
  activityStatuses: [],
  attributes: [],
  period: 'month',
  show: null,
  currency: null,
  tables: [],
};

const idSchema = z.string().regex(/^[A-Z]+-[0-9A-Z]{1,20}$/);
const planTypeSchema = z.enum(['operational', 'strategic', 'crossMarket']);
const statusSchema = z.enum([
  'created',
  'submitted',
  'declined',
  'approved',
  'inProgress',
  'complete',
]);
const widgetKeySchema = z.string().regex(/^[a-z][A-Za-z0-9]{0,39}$/);
const attrSchema = z
  .string()
  .regex(/^[A-Z]+-[0-9A-Z]{1,20}:[A-Za-z0-9-]+(,[A-Za-z0-9-]+)*$/)
  .transform((text) => {
    const [attributeId = '', rest = ''] = text.split(':');
    return { attributeId, values: rest.split(',') };
  });

/** The values of a parameter the schema accepts, unique, in their order; malformed ones are dropped (FR-015). */
function listOf<T>(params: URLSearchParams, name: string, schema: z.ZodType<T>): T[] {
  const kept: T[] = [];
  for (const raw of params.getAll(name)) {
    const parsed = schema.safeParse(raw);
    if (parsed.success && !kept.includes(parsed.data)) {
      kept.push(parsed.data);
    }
  }
  return kept;
}

function oneOf<T>(params: URLSearchParams, name: string, schema: z.ZodType<T>): T | null {
  const parsed = schema.safeParse(params.get(name) ?? undefined);
  return parsed.success ? parsed.data : null;
}

export function readAddress(params: URLSearchParams): ReportAddress {
  return {
    year: oneOf(params, 'year', z.coerce.number().int().min(2000).max(2100)),
    plans: listOf(params, 'plan', idSchema),
    planTypes: listOf(params, 'planType', planTypeSchema),
    teams: listOf(params, 'team', idSchema),
    owners: listOf(params, 'owner', idSchema),
    activityTypes: listOf(params, 'activityType', idSchema),
    activityStatuses: listOf(params, 'activityStatus', statusSchema),
    attributes: mergeAttributes(
      params.getAll('attr').flatMap((raw) => {
        const parsed = attrSchema.safeParse(raw);
        return parsed.success ? [parsed.data] : [];
      }),
    ),
    period: oneOf(params, 'period', z.enum(['month', 'quarter'])) ?? 'month',
    show: oneOf(params, 'show', z.enum(['plan', 'campaign', 'initiative', 'activity'])),
    currency: oneOf(params, 'currency', z.string().regex(/^[A-Z]{3}$/)),
    tables: listOf(params, 'table', widgetKeySchema),
  };
}

function mergeAttributes(choices: AttributeChoice[]): AttributeChoice[] {
  const merged: AttributeChoice[] = [];
  for (const choice of choices) {
    const same = merged.find((m) => m.attributeId === choice.attributeId);
    if (same) {
      same.values = [...new Set([...same.values, ...choice.values])];
    } else {
      merged.push({ attributeId: choice.attributeId, values: [...new Set(choice.values)] });
    }
  }
  return merged;
}

const sorted = (values: readonly string[]) => [...values].sort();

/** The address in the API's canonical form (039 data-model §2): its order, values sorted, defaults left out. */
export function writeAddress(address: ReportAddress): URLSearchParams {
  const params = new URLSearchParams();
  const add = (name: string, values: readonly string[]) => {
    for (const value of sorted(values)) {
      params.append(name, value);
    }
  };
  if (address.year !== null) {
    params.append('year', String(address.year));
  }
  add('plan', address.plans);
  add('planType', address.planTypes);
  add('team', address.teams);
  add('owner', address.owners);
  add('activityType', address.activityTypes);
  add('activityStatus', address.activityStatuses);
  add(
    'attr',
    address.attributes.map((a) => `${a.attributeId}:${sorted(a.values).join(',')}`),
  );
  if (address.period !== 'month') {
    params.append('period', address.period);
  }
  add('show', address.show === null ? [] : [address.show]);
  add('currency', address.currency === null ? [] : [address.currency]);
  add('table', address.tables);
  return params;
}

/** The address as the report operations' query; the defaults are left for the API. */
export function toQuery(address: ReportAddress): ReportQuery {
  return {
    ...(address.year !== null && { year: address.year }),
    ...(address.plans.length > 0 && { plan: address.plans }),
    ...(address.planTypes.length > 0 && { planType: address.planTypes }),
    ...(address.teams.length > 0 && { team: address.teams }),
    ...(address.owners.length > 0 && { owner: address.owners }),
    ...(address.activityTypes.length > 0 && { activityType: address.activityTypes }),
    ...(address.activityStatuses.length > 0 && { activityStatus: address.activityStatuses }),
    ...(address.attributes.length > 0 && {
      attr: address.attributes.map((a) => `${a.attributeId}:${a.values.join(',')}`),
    }),
    period: address.period,
    ...(address.show !== null && { show: address.show }),
    ...(address.currency !== null && { currency: address.currency }),
  };
}

/** Whether any filter restricts the records, the year aside (FR-014's Reset filters). */
export function isRestricted(address: ReportAddress): boolean {
  return (
    address.plans.length +
      address.planTypes.length +
      address.teams.length +
      address.owners.length +
      address.activityTypes.length +
      address.activityStatuses.length +
      address.attributes.length >
    0
  );
}

/** Every filter but the year back to its default; period, show, currency and tables stay (FR-014). */
export function resetFilters(address: ReportAddress): ReportAddress {
  return {
    ...DEFAULT_ADDRESS,
    year: address.year,
    period: address.period,
    show: address.show,
    currency: address.currency,
    tables: address.tables,
  };
}
