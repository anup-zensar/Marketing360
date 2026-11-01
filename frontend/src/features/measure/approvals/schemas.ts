import { z } from 'zod';

export const RECORD_TYPES = ['activity', 'purchaseRequest', 'modifiedPurchaseRequest'] as const;
export type RecordType = (typeof RECORD_TYPES)[number];

export const SORT_KEYS = [
  'itemId',
  'itemName',
  'recordType',
  'approver',
  'decidedOn',
  'outcome',
  'submittedOn',
  'level',
  'purchaseOrder',
  'amount',
  'owner',
  'comments',
] as const;
export type SortKey = (typeof SORT_KEYS)[number];

export const PAGE_SIZES = [5, 10, 25, 50, 100] as const;
export type PageSize = (typeof PAGE_SIZES)[number];

export interface ApprovalsParams {
  statusType: RecordType;
  recordTypes: string[];
  approvers: string[];
  outcomes: string[];
  q: string;
  sort: SortKey;
  order: 'asc' | 'desc';
  page: number;
  pageSize: PageSize;
}

export const DEFAULT_PARAMS: ApprovalsParams = {
  statusType: 'activity',
  recordTypes: [],
  approvers: [],
  outcomes: [],
  q: '',
  sort: 'submittedOn',
  order: 'desc',
  page: 1,
  pageSize: 50,
};

const wordSchema = z.string().regex(/^[A-Za-z0-9-]{1,40}$/);

function words(params: URLSearchParams, name: string): string[] {
  const kept: string[] = [];
  for (const raw of params.getAll(name)) {
    const parsed = wordSchema.safeParse(raw);
    if (parsed.success && !kept.includes(parsed.data)) {
      kept.push(parsed.data);
    }
  }
  return kept;
}

function pick<T>(params: URLSearchParams, name: string, schema: z.ZodType<T>, fallback: T): T {
  const parsed = schema.safeParse(params.get(name) ?? undefined);
  return parsed.success ? parsed.data : fallback;
}

export function readApprovalsParams(params: URLSearchParams): ApprovalsParams {
  return {
    statusType: pick(params, 'statusType', z.enum(RECORD_TYPES), DEFAULT_PARAMS.statusType),
    recordTypes: words(params, 'recordType'),
    approvers: words(params, 'approver'),
    outcomes: words(params, 'outcome'),
    q: pick(params, 'q', z.string().trim().min(1).max(200), ''),
    sort: pick(params, 'sort', z.enum(SORT_KEYS), DEFAULT_PARAMS.sort),
    order: pick(params, 'order', z.enum(['asc', 'desc']), DEFAULT_PARAMS.order),
    page: pick(params, 'page', z.coerce.number().int().min(1), 1),
    pageSize:
      PAGE_SIZES.find((n) => n === Number(params.get('pageSize'))) ?? DEFAULT_PARAMS.pageSize,
  };
}

const OWN_NAMES = [
  'statusType',
  'recordType',
  'approver',
  'outcome',
  'q',
  'sort',
  'order',
  'page',
  'pageSize',
];

/** The address with the report's own parameters set; a default is left out, so the address stays short. */
export function writeApprovalsParams(
  params: URLSearchParams,
  next: ApprovalsParams,
): URLSearchParams {
  const written = new URLSearchParams(params);
  for (const name of OWN_NAMES) {
    written.delete(name);
  }
  const set = (name: string, value: string, fallback: string) => {
    if (value !== fallback) {
      written.set(name, value);
    }
  };
  set('statusType', next.statusType, DEFAULT_PARAMS.statusType);
  set('q', next.q, '');
  set('sort', next.sort, DEFAULT_PARAMS.sort);
  set('order', next.order, DEFAULT_PARAMS.order);
  set('page', String(next.page), '1');
  set('pageSize', String(next.pageSize), String(DEFAULT_PARAMS.pageSize));
  next.recordTypes.forEach((value) => written.append('recordType', value));
  next.approvers.forEach((value) => written.append('approver', value));
  next.outcomes.forEach((value) => written.append('outcome', value));
  return written;
}
