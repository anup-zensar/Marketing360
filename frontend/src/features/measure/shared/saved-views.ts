import { useCallback } from 'react';
import { z } from 'zod';
import { ApiError } from '@/api/errors';
import type { FieldProblem } from '@/ui/FormErrorSummary/FormErrorSummary';
import { useToast } from '@/ui/Toast/useToast';
import type { ReportView } from './api';

const NAME_MAX = 60;
const VIEW_GONE = 'This view no longer exists — it may have been deleted in another tab.';

const viewName = z
  .string()
  .refine((name) => name.trim() !== '', 'Enter a name for the view.')
  .refine((name) => [...name.trim()].length <= NAME_MAX, 'Use 60 characters or fewer.');

export const saveViewSchema = z.object({ name: viewName, isDefault: z.boolean() });
export type SaveViewValues = z.infer<typeof saveViewSchema>;

export function viewNamed(views: ReportView[], name: string, except?: string) {
  const wanted = name.trim().toLocaleLowerCase();
  return views.find((view) => view.id !== except && view.name.toLocaleLowerCase() === wanted);
}

export function renameViewSchema(views: ReportView[], viewId: string) {
  return z.object({
    name: viewName.superRefine((name, context) => {
      if (viewNamed(views, name, viewId)) {
        context.addIssue({
          code: 'custom',
          message: `You already have a view called '${name.trim()}'.`,
        });
      }
    }),
  });
}

/** Two addresses show the same report whatever their order; the page of a table is not part of a view. */
export function sameQuery(a: string, b: string): boolean {
  const normal = (query: string) => {
    const params = new URLSearchParams(query);
    params.delete('page');
    params.sort();
    return params.toString();
  };
  return normal(a) === normal(b);
}

/** A refused change: the name's problems for the form; anything else, a stale version too, as a message. */
export function useViewRefusal(): (error: unknown, nameField: string) => FieldProblem[] {
  const { show } = useToast();
  return useCallback(
    (error: unknown, nameField: string) => {
      if (error instanceof ApiError && error.kind === 'validation' && error.fields.length > 0) {
        return error.fields.map((f) => ({
          field: f.field === 'name' ? nameField : f.field,
          message: f.message,
        }));
      }
      const failure = error instanceof ApiError ? error : new ApiError('other');
      if (failure.kind !== 'signInRequired' && failure.kind !== 'sessionExpired') {
        show({ message: failure.kind === 'notFound' ? VIEW_GONE : failure.message });
      }
      return [];
    },
    [show],
  );
}
