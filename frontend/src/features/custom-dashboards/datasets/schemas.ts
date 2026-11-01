import { z } from 'zod';
import type { DatasetPreview } from '../model';

export const KIND_OPTIONS = [
  { id: 'text', label: 'Text', active: true },
  { id: 'number', label: 'Number', active: true },
  { id: 'date', label: 'Date', active: true },
  { id: 'money', label: 'Money', active: true },
] as const;

const KIND = z.enum(['text', 'number', 'date', 'money']);

export function kindOf(value: string): z.infer<typeof KIND> {
  const parsed = KIND.safeParse(value);
  return parsed.success ? parsed.data : 'text';
}

const COLUMN = z
  .object({
    label: z.string().trim().min(1, 'Enter a column name.').max(60, 'Use at most 60 characters.'),
    kind: KIND,
    currency: z.string().nullable(),
  })
  .superRefine((column, context) => {
    if (column.kind === 'money' && !column.currency) {
      context.addIssue({ code: 'custom', path: ['currency'], message: 'Choose the currency.' });
    }
  });

export const datasetSchema = z
  .object({
    name: z.string().trim().min(1, 'Enter a name.').max(80, 'Use at most 80 characters.'),
    description: z.string().max(500, 'Use at most 500 characters.'),
    columns: z.array(COLUMN),
  })
  .superRefine(({ columns }, context) => {
    const seen = new Set<string>();
    columns.forEach((column, index) => {
      const key = column.label.trim().toLowerCase();
      if (key && seen.has(key)) {
        context.addIssue({
          code: 'custom',
          path: ['columns', index, 'label'],
          message: 'Another column has this name.',
        });
      }
      seen.add(key);
    });
  });

export type DatasetValues = z.infer<typeof datasetSchema>;

export function startValues(preview: DatasetPreview, name: string): DatasetValues {
  return {
    name,
    description: '',
    columns: preview.columns.map((column) => ({
      label: column.label,
      kind: column.proposedKind,
      currency: column.proposedCurrency,
    })),
  };
}

export const detailsSchema = z.object({
  name: z.string().trim().min(1, 'Enter a name.').max(80, 'Use at most 80 characters.'),
  description: z.string().max(500, 'Use at most 500 characters.'),
  labels: z.array(
    z.string().trim().min(1, 'Enter a column name.').max(60, 'Use at most 60 characters.'),
  ),
});

export type DetailsValues = z.infer<typeof detailsSchema>;
