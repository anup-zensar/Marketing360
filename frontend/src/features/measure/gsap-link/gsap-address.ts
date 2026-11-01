import { useSearchParams } from 'react-router';

export type GsapTab = 'swbs' | 'waiting' | 'orders';

export const GSAP_TABS: GsapTab[] = ['swbs', 'waiting', 'orders'];

/** Rows per page, 5 to 500 (044 FR-015). */
export const PAGE_SIZES = [5, 10, 20, 50, 100, 200, 500] as const;
const DEFAULT_PAGE_SIZE = 20;

/** The columns each tab sorts by: the contract's GsapLinkSort. */
const SORT_KEYS: Record<GsapTab, string[]> = {
  swbs: ['lineItem', 'plan', 'owner', 'companyCode', 'approvedAt', 'waiting'],
  waiting: [
    'purchaseRequest',
    'vendor',
    'owner',
    'plan',
    'routedAt',
    'gsapPrNumber',
    'amount',
    'waiting',
  ],
  orders: [
    'lineItem',
    'plan',
    'purchaseRequest',
    'gsapPrNumber',
    'poNumber',
    'poDate',
    'vendor',
    'poAmount',
    'invoiced',
    'paid',
  ],
};

export interface GsapTable {
  tab: GsapTab;
  sort: string;
  page: number;
  pageSize: number;
}

/** Longest waiting first on the first two tabs; line item, then PO number, on Purchase orders. */
export function defaultSort(tab: GsapTab): string {
  return tab === 'orders' ? 'lineItem' : '-waiting';
}

function validSort(tab: GsapTab, sort: string | null): sort is string {
  return sort !== null && SORT_KEYS[tab].includes(sort.replace(/^-/, ''));
}

function setOrDrop(params: URLSearchParams, name: string, value: string, fallback: string) {
  if (value === fallback) {
    params.delete(name);
  } else {
    params.set(name, value);
  }
}

/** The tab, its sort and its page in the address (FR-001, FR-015); a new tab, sort or page size starts at page 1. */
export function useGsapTable() {
  const [params, setParams] = useSearchParams();
  const tab = GSAP_TABS.find((t) => t === params.get('tab')) ?? 'swbs';
  const sort = params.get('sort');
  const page = Number(params.get('page'));
  const size = Number(params.get('pageSize'));
  const table: GsapTable = {
    tab,
    sort: validSort(tab, sort) ? sort : defaultSort(tab),
    page: Number.isInteger(page) && page > 1 ? page : 1,
    pageSize: PAGE_SIZES.find((s) => s === size) ?? DEFAULT_PAGE_SIZE,
  };
  const change = (patch: Partial<GsapTable>) => {
    const next = { ...table, page: 1, ...patch };
    if (patch.tab !== undefined && patch.tab !== table.tab) {
      next.sort = defaultSort(patch.tab);
    }
    const written = new URLSearchParams(params);
    setOrDrop(written, 'tab', next.tab, 'swbs');
    setOrDrop(written, 'sort', next.sort, defaultSort(next.tab));
    setOrDrop(written, 'page', String(next.page), '1');
    setOrDrop(written, 'pageSize', String(next.pageSize), String(DEFAULT_PAGE_SIZE));
    setParams(written);
  };
  return { table, change };
}
