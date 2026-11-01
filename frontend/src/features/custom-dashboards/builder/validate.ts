import { z } from 'zod';
import { MAX_WIDGETS } from '../model';
import type { DataSource, Widget } from '../model';
import type { Draft } from './builder-reducer';
import { blockedBy, overlaps } from './grid';
import { displayProblems, queryProblems, tooLong } from './validate-query';
import type { Found } from './validate-query';
import { widgetName } from './widget-defaults';

export interface Problem {
  field: string;
  message: string;
  widgetId?: string;
}

const detailsSchema = z.object({
  name: z.string().trim().min(1, 'Enter a name.').max(80, 'Use at most 80 characters.'),
  description: z.string().max(500, 'Use at most 500 characters.'),
  filters: z
    .object({
      dateFrom: z.string().nullable(),
      dateTo: z.string().nullable(),
      planIds: z.array(z.string()).max(20, 'Choose at most 20 plans.'),
      teamIds: z.array(z.string()).max(20, 'Choose at most 20 teams.'),
    })
    .refine(
      (filters) => !filters.dateFrom || !filters.dateTo || filters.dateFrom <= filters.dateTo,
      {
        message: 'The end date must be on or after the start date.',
        path: ['dateTo'],
      },
    ),
});

export function validateDraft(draft: Draft, sources: DataSource[] | undefined): Problem[] {
  const widgetCount =
    draft.widgets.length > MAX_WIDGETS
      ? [{ field: 'widgets', message: `A dashboard holds at most ${MAX_WIDGETS} widgets.` }]
      : [];
  return [
    ...detailsProblems(draft),
    ...widgetCount,
    ...draft.widgets.flatMap((_, index) => widgetProblems(draft.widgets, index, sources)),
  ];
}

export function problemsOf(problems: readonly Problem[], widgetId: string): Problem[] {
  return problems.filter((problem) => problem.widgetId === widgetId);
}

export function problemAt(problems: readonly Problem[], field: string): string | undefined {
  return problems.find((problem) => problem.field === field)?.message;
}

function detailsProblems(draft: Draft): Problem[] {
  const result = detailsSchema.safeParse(draft);
  if (result.success) {
    return [];
  }
  return result.error.issues.map((issue) => ({
    field: issue.path.map(String).join('.'),
    message: issue.message,
  }));
}

function widgetProblems(
  widgets: readonly Widget[],
  index: number,
  sources: DataSource[] | undefined,
): Problem[] {
  const widget = widgets[index];
  if (!widget) {
    return [];
  }
  const prefix = `widgets[${index}]`;
  const found: Found[] = [
    ...textProblems(widget),
    ...layoutProblems(widgets, widget),
    ...queryProblems(widget, sources),
    ...displayProblems(widget),
  ];
  return found.map(([path, message]) => ({
    field: `${prefix}.${path}`,
    message,
    widgetId: widget.id,
  }));
}

function textProblems(widget: Widget): Found[] {
  return [
    tooLong('title', widget.title, 80),
    tooLong('subtitle', widget.subtitle, 120),
    tooLong('note', widget.note, 2000),
  ].filter((found) => found !== null);
}

function layoutProblems(widgets: readonly Widget[], widget: Widget): Found[] {
  const blocked = blockedBy(widget.layout, widgets, widget.id);
  if (blocked === 'outside') {
    return [['layout', "This widget goes past the grid's edge."]];
  }
  const other = widgets.find(
    (each) => each.id !== widget.id && overlaps(each.layout, widget.layout),
  );
  return other ? [['layout', `This widget overlaps '${widgetName(other)}'.`]] : [];
}
