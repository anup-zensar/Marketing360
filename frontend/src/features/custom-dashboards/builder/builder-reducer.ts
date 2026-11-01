import { stableHash } from '../keys';
import { MAX_WIDGETS, NO_FILTERS } from '../model';
import type {
  CustomDashboard,
  CustomDashboardFilters,
  Widget,
  WidgetDisplay,
  WidgetLayout,
  WidgetQuery,
  WidgetType,
} from '../model';
import { widgetTypeInfo } from '../widget-types';
import { blockedBy, firstFreePlace, placeDrop, stepped } from './grid';
import type { Cell, Step } from './grid';
import { placeMessage } from './announcements';
import { fitDisplay, newWidget, nextWidgetId, typeName, widgetName } from './widget-defaults';

export const HISTORY_LIMIT = 50;

export interface Draft {
  name: string;
  description: string;
  filters: CustomDashboardFilters;
  widgets: Widget[];
}

interface Change {
  draft: Draft;
  label: string;
}

export interface BuilderState {
  draft: Draft;
  saved: Draft;
  version: number;
  selectedId: string | null;
  past: Change[];
  future: Change[];
  typingKey: string | null;
  announcement: { text: string; count: number };
}

type TextPatch = Partial<Pick<Widget, 'title' | 'subtitle' | 'note'>>;

export type BuilderAction =
  | { type: 'load'; dashboard: CustomDashboard | null }
  | { type: 'saved'; dashboard: CustomDashboard }
  | { type: 'followVersion'; version: number }
  | { type: 'select'; id: string | null }
  | { type: 'rename'; name: string }
  | { type: 'describe'; description: string }
  | { type: 'setFilters'; filters: CustomDashboardFilters }
  | { type: 'add'; widgetType: WidgetType; at?: Cell }
  | { type: 'duplicate'; id: string }
  | { type: 'delete'; id: string }
  | { type: 'place'; id: string; layout: WidgetLayout }
  | { type: 'step'; id: string; step: Step; resize: boolean }
  | { type: 'setQuery'; id: string; query: WidgetQuery; typing?: string }
  | { type: 'setDisplay'; id: string; display: WidgetDisplay; typing?: string }
  | { type: 'setText'; id: string; patch: TextPatch }
  | { type: 'undo' }
  | { type: 'redo' };

const EMPTY_DRAFT: Draft = { name: '', description: '', filters: NO_FILTERS, widgets: [] };

export function initialState(dashboard: CustomDashboard | null): BuilderState {
  const draft = dashboard ? draftOf(dashboard) : EMPTY_DRAFT;
  return {
    draft,
    saved: draft,
    version: dashboard?.version ?? 0,
    selectedId: null,
    past: [],
    future: [],
    typingKey: null,
    announcement: { text: '', count: 0 },
  };
}

export function isDirty(state: BuilderState): boolean {
  return stableHash(state.draft) !== stableHash(state.saved);
}

export function selectedWidget(state: BuilderState): Widget | null {
  return state.draft.widgets.find((widget) => widget.id === state.selectedId) ?? null;
}

export function builderReducer(state: BuilderState, action: BuilderAction): BuilderState {
  switch (action.type) {
    case 'load':
      return initialState(action.dashboard);
    case 'saved':
      return savedState(state, action.dashboard);
    case 'followVersion':
      return { ...state, version: action.version };
    case 'select':
      return { ...state, selectedId: action.id, typingKey: null };
    case 'rename':
      return change(state, { ...state.draft, name: action.name }, 'renamed the dashboard', 'name');
    case 'describe':
      return change(
        state,
        { ...state.draft, description: action.description },
        'changed the description',
        'description',
      );
    case 'setFilters':
      return change(state, { ...state.draft, filters: action.filters }, 'changed the filters');
    case 'add':
      return add(state, action.widgetType, action.at);
    case 'duplicate':
      return duplicate(state, action.id);
    case 'delete':
      return remove(state, action.id);
    case 'place':
      return place(state, action.id, action.layout);
    case 'step':
      return step(state, action);
    case 'setQuery':
      return edit(state, action.id, (widget) => fitDisplay({ ...widget, query: action.query }), {
        typing: action.typing && `query:${action.id}:${action.typing}`,
      });
    case 'setDisplay':
      return edit(state, action.id, (widget) => ({ ...widget, display: action.display }), {
        typing: action.typing && `display:${action.id}:${action.typing}`,
      });
    case 'setText':
      return edit(state, action.id, (widget) => ({ ...widget, ...action.patch }), {
        typing: `text:${action.id}:${Object.keys(action.patch).join()}`,
      });
    case 'undo':
      return travel(state, 'undo');
    case 'redo':
      return travel(state, 'redo');
    default:
      return unknownAction(action);
  }
}

function draftOf(dashboard: CustomDashboard): Draft {
  const { name, description, filters, widgets } = dashboard;
  return { name, description, filters, widgets };
}

function savedState(state: BuilderState, dashboard: CustomDashboard): BuilderState {
  const draft = draftOf(dashboard);
  const selectedId = draft.widgets.some((widget) => widget.id === state.selectedId)
    ? state.selectedId
    : null;
  return { ...state, draft, saved: draft, version: dashboard.version, selectedId, typingKey: null };
}

function announce(state: BuilderState, text: string): BuilderState {
  return { ...state, announcement: { text, count: state.announcement.count + 1 } };
}

function change(state: BuilderState, draft: Draft, label: string, typing?: string): BuilderState {
  const typingKey = typing ?? null;
  const merges = typingKey !== null && typingKey === state.typingKey;
  const past = merges
    ? state.past
    : [...state.past, { draft: state.draft, label }].slice(-HISTORY_LIMIT);
  return { ...state, draft, past, future: [], typingKey };
}

function withWidgets(state: BuilderState, widgets: Widget[]): Draft {
  return { ...state.draft, widgets };
}

function add(state: BuilderState, type: WidgetType, at?: Cell): BuilderState {
  const { widgets } = state.draft;
  const size = widgetTypeInfo(type).size;
  const id = nextWidgetId(widgets);
  const layout = at ? placeDrop(size, at, widgets) : firstFreePlace(size, widgets);
  if (widgets.length >= MAX_WIDGETS || !id) {
    return announce(state, `A dashboard holds at most ${MAX_WIDGETS} widgets.`);
  }
  if (!layout) {
    return announce(state, "There's no free place for it on the grid.");
  }
  const added = newWidget(type, id, layout);
  const next = change(state, withWidgets(state, [...widgets, added]), `added ${typeName(type)}`);
  return announce({ ...next, selectedId: id }, `${typeName(type)} added.`);
}

function duplicate(state: BuilderState, id: string): BuilderState {
  const { widgets } = state.draft;
  const original = widgets.find((widget) => widget.id === id);
  const newId = nextWidgetId(widgets);
  const layout =
    original && firstFreePlace({ w: original.layout.w, h: original.layout.h }, widgets);
  if (!original || !newId || widgets.length >= MAX_WIDGETS || !layout) {
    return announce(state, "Can't duplicate: there's no room for another widget.");
  }
  const copy = { ...original, id: newId, layout };
  const label = `duplicated '${widgetName(original)}'`;
  const next = change(state, withWidgets(state, [...widgets, copy]), label);
  return announce({ ...next, selectedId: newId }, `'${widgetName(original)}' duplicated.`);
}

function remove(state: BuilderState, id: string): BuilderState {
  const widget = state.draft.widgets.find((each) => each.id === id);
  if (!widget) {
    return state;
  }
  const rest = state.draft.widgets.filter((each) => each.id !== id);
  const next = change(state, withWidgets(state, rest), `deleted '${widgetName(widget)}'`);
  const selectedId = state.selectedId === id ? null : state.selectedId;
  return announce({ ...next, selectedId }, `'${widgetName(widget)}' deleted.`);
}

function place(state: BuilderState, id: string, layout: WidgetLayout): BuilderState {
  const widget = state.draft.widgets.find((each) => each.id === id);
  if (!widget || stableHash(widget.layout) === stableHash(layout)) {
    return state;
  }
  const resizing = widget.layout.x === layout.x && widget.layout.y === layout.y;
  const blocked = blockedBy(layout, state.draft.widgets, id);
  if (blocked) {
    return announce(state, placeMessage({ blocked, resizing, layout }));
  }
  const label = `${resizing ? 'resized' : 'moved'} '${widgetName(widget)}'`;
  const widgets = state.draft.widgets.map((each) => (each.id === id ? { ...each, layout } : each));
  return announce(
    change(state, withWidgets(state, widgets), label),
    placeMessage({ blocked, resizing, layout }),
  );
}

function step(state: BuilderState, action: Extract<BuilderAction, { type: 'step' }>): BuilderState {
  const widget = state.draft.widgets.find((each) => each.id === action.id);
  return widget
    ? place(state, widget.id, stepped(widget.layout, action.step, action.resize))
    : state;
}

function edit(
  state: BuilderState,
  id: string,
  update: (widget: Widget) => Widget,
  { typing }: { typing?: string | undefined },
): BuilderState {
  const widget = state.draft.widgets.find((each) => each.id === id);
  if (!widget) {
    return state;
  }
  const widgets = state.draft.widgets.map((each) => (each.id === id ? update(each) : each));
  return change(state, withWidgets(state, widgets), `changed '${widgetName(widget)}'`, typing);
}

function travel(state: BuilderState, way: 'undo' | 'redo'): BuilderState {
  const from = way === 'undo' ? state.past : state.future;
  const last = from.at(-1);
  if (!last) {
    return announce(state, way === 'undo' ? 'Nothing to undo.' : 'Nothing to redo.');
  }
  const kept = from.slice(0, -1);
  const back = [
    ...(way === 'undo' ? state.future : state.past),
    { draft: state.draft, label: last.label },
  ];
  const selectedId = last.draft.widgets.some((widget) => widget.id === state.selectedId)
    ? state.selectedId
    : null;
  const moved =
    way === 'undo'
      ? { past: kept, future: back }
      : { past: back.slice(-HISTORY_LIMIT), future: kept };
  return announce(
    { ...state, ...moved, draft: last.draft, selectedId, typingKey: null },
    `${way === 'undo' ? 'Undone' : 'Redone'}: ${last.label}.`,
  );
}

function unknownAction(action: never): never {
  throw new Error(`Unknown builder action: ${JSON.stringify(action)}`);
}
