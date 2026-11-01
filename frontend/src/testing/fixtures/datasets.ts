/** Datasets of feature 051 for the Datasets pages and the upload, shaped as the API answers them. */
import type { components } from '@/api/generated/schema';
import { AVAILABLE, DANIEL_PERSON, EUROPE_REF, HIDDEN, PRIYA_PERSON } from './shared-dashboards';

type S = components['schemas'];

const OWNER_ACCESS: S['DatasetAccess'] = { edit: AVAILABLE, share: AVAILABLE, delete: AVAILABLE };
export const READER_ACCESS: S['DatasetAccess'] = { edit: HIDDEN, share: HIDDEN, delete: HIDDEN };

export const CURRENCIES: S['DisplayCurrencyPage'] = {
  items: [
    { code: 'USD', name: 'US dollar', active: true },
    { code: 'EUR', name: 'Euro', active: true },
  ],
  total: 2,
};

export function datasetSummary(overrides: Partial<S['DatasetSummary']> = {}): S['DatasetSummary'] {
  return {
    id: 'DS-1001',
    name: 'Event leads 2026',
    description: 'Leads from agency events.',
    owner: PRIYA_PERSON,
    sharedWith: [],
    rowCount: 12,
    columnCount: 4,
    updatedAt: '2026-10-20T08:00:00Z',
    access: OWNER_ACCESS,
    ...overrides,
  };
}

export const SHARED_DATASET = datasetSummary({
  id: 'DS-1002',
  name: 'China retail panel',
  owner: DANIEL_PERSON,
  sharedWith: [EUROPE_REF],
  rowCount: 1250,
  columnCount: 6,
  access: READER_ACCESS,
});

export const COLUMNS: S['DatasetColumn'][] = [
  { key: 'c1', position: 0, label: 'Event', kind: 'text', currency: null },
  { key: 'c2', position: 1, label: 'Date', kind: 'date', currency: null },
  { key: 'c3', position: 2, label: 'Leads', kind: 'number', currency: null },
  { key: 'c4', position: 3, label: 'Agency cost', kind: 'money', currency: 'EUR' },
];

export function dataset(overrides: Partial<S['Dataset']> = {}): S['Dataset'] {
  return {
    id: 'DS-1001',
    name: 'Event leads 2026',
    description: 'Leads from agency events.',
    owner: PRIYA_PERSON,
    isOwner: true,
    sharedWith: [],
    shareableTeams: [EUROPE_REF],
    fileName: 'event-leads.xlsx',
    sheetName: 'Leads',
    rowCount: 12,
    columns: COLUMNS,
    usedBy: [{ id: 'DB-1001', name: 'Lubricants Europe overview' }],
    access: OWNER_ACCESS,
    version: 2,
    createdAt: '2026-10-20T08:00:00Z',
    createdBy: { id: 'U-1002', name: 'Priya Raman' },
    updatedAt: '2026-10-20T08:00:00Z',
    updatedBy: { id: 'U-1002', name: 'Priya Raman' },
    ...overrides,
  };
}

export function preview(overrides: Partial<S['DatasetPreview']> = {}): S['DatasetPreview'] {
  return {
    fileName: 'event-leads.xlsx',
    sheetName: 'Leads',
    rowCount: 12,
    columns: [
      col(0, 'Event', 'text', ['Lubes Expo', 'Auto Show']),
      col(1, 'Date', 'date', ['2026-03-04', '2026-05-12']),
      col(2, 'Leads', 'number', ['120', '85']),
      { ...col(3, 'Cost', 'money', ['12000', '9500']), proposedCurrency: 'USD', emptyCount: 1 },
    ],
    rows: [
      ['Lubes Expo', '2026-03-04', '120', '12000'],
      ['Auto Show', '2026-05-12', '85', ''],
    ],
    problems: [],
    ...overrides,
  };
}

function col(
  position: number,
  label: string,
  proposedKind: S['DataFieldKind'],
  samples: string[],
): S['DatasetPreviewColumn'] {
  return { position, label, proposedKind, proposedCurrency: null, samples, emptyCount: 0 };
}

export function rowPage(offset: number, total: number, count: number): S['DatasetRowPage'] {
  return {
    columns: COLUMNS,
    rows: Array.from({ length: count }, (_, i) => [
      `Event ${offset + i + 1}`,
      '2026-03-04',
      String(10 + i),
      i === 0 ? null : '1200.00',
    ]),
    offset,
    limit: 50,
    total,
  };
}
