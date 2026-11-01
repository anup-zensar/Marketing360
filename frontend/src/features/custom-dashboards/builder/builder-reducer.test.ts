import { describe, expect, it } from 'vitest';
import { BAR_QUERY, BAR_WIDGET, dashboard } from './builder-fixtures';
import {
  builderReducer,
  HISTORY_LIMIT,
  initialState,
  isDirty,
  selectedWidget,
} from './builder-reducer';
import type { BuilderAction, BuilderState } from './builder-reducer';

function run(state: BuilderState, ...actions: BuilderAction[]): BuilderState {
  return actions.reduce(builderReducer, state);
}

const empty = () => initialState(null);
const loaded = () => initialState(dashboard());

describe('051 FR-027: adding widgets from the palette', () => {
  it('places a clicked type in the first free place with its default size, selected', () => {
    const state = run(
      empty(),
      { type: 'add', widgetType: 'bar' },
      { type: 'add', widgetType: 'kpi' },
    );
    expect(state.draft.widgets.map((widget) => [widget.id, widget.layout])).toEqual([
      ['w1', { x: 0, y: 0, w: 6, h: 4 }],
      ['w2', { x: 6, y: 0, w: 3, h: 2 }],
    ]);
    expect(state.selectedId).toBe('w2');
    expect(state.announcement.text).toBe('Key figure added.');
  });

  it('places a dropped type where it was dropped', () => {
    const state = run(empty(), { type: 'add', widgetType: 'line', at: { x: 3, y: 5 } });
    expect(state.draft.widgets[0]?.layout).toEqual({ x: 3, y: 5, w: 6, h: 4 });
    expect(state.announcement.text).toBe('Line chart added.');
  });

  it('gives a text note no query and a chart an empty query to complete', () => {
    const state = run(
      empty(),
      { type: 'add', widgetType: 'text' },
      { type: 'add', widgetType: 'pie' },
    );
    expect(state.draft.widgets[0]?.query).toBeNull();
    expect(state.draft.widgets[1]?.query).toMatchObject({ source: '', measures: [], limit: 50 });
  });

  it('refuses a 31st widget and says why', () => {
    const thirty = Array.from({ length: 30 }, (): BuilderAction => ({
      type: 'add',
      widgetType: 'kpi',
    }));
    const state = run(empty(), ...thirty, { type: 'add', widgetType: 'kpi' });
    expect(state.draft.widgets).toHaveLength(30);
    expect(state.announcement.text).toBe('A dashboard holds at most 30 widgets.');
  });
});

describe('051 FR-028: moving and resizing on the grid', () => {
  it('moves one cell with an arrow key and announces where', () => {
    const state = run(loaded(), { type: 'step', id: 'w2', step: 'down', resize: false });
    expect(state.draft.widgets[1]?.layout).toEqual({ x: 6, y: 1, w: 4, h: 2 });
    expect(state.announcement.text).toBe('Moved to column 7, row 2.');
  });

  it('resizes one cell with Shift and an arrow key', () => {
    const state = run(loaded(), { type: 'step', id: 'w2', step: 'right', resize: true });
    expect(state.draft.widgets[1]?.layout).toEqual({ x: 6, y: 0, w: 5, h: 2 });
    expect(state.announcement.text).toBe('Resized to 5 columns by 2 rows.');
  });

  it('never moves over another widget, and says so', () => {
    const state = run(loaded(), { type: 'step', id: 'w2', step: 'left', resize: false });
    expect(state.draft.widgets[1]?.layout).toEqual(dashboard().widgets[1]?.layout);
    expect(state.announcement.text).toBe("Can't move there: another widget is in the way.");
    expect(state.past).toHaveLength(0);
  });

  it('never moves past the edge of the grid', () => {
    const state = run(loaded(), { type: 'step', id: 'w1', step: 'up', resize: false });
    expect(state.announcement.text).toBe("Can't move there: that's the edge of the grid.");
  });

  it('places a widget dragged with the pointer or set in the Layout tab', () => {
    const state = run(loaded(), { type: 'place', id: 'w1', layout: { x: 0, y: 4, w: 8, h: 4 } });
    expect(state.draft.widgets[0]?.layout).toEqual({ x: 0, y: 4, w: 8, h: 4 });
  });
});

describe('051 FR-029: duplicating and deleting', () => {
  it('duplicates a widget into the next free place, selected', () => {
    const state = run(loaded(), { type: 'duplicate', id: 'w1' });
    expect(state.draft.widgets[2]).toMatchObject({
      id: 'w3',
      title: BAR_WIDGET.title,
      query: BAR_WIDGET.query,
      layout: { x: 6, y: 2, w: 6, h: 4 },
    });
    expect(state.selectedId).toBe('w3');
  });

  it('deletes a widget and clears the selection', () => {
    const state = run(loaded(), { type: 'select', id: 'w1' }, { type: 'delete', id: 'w1' });
    expect(state.draft.widgets.map((widget) => widget.id)).toEqual(['w2']);
    expect(selectedWidget(state)).toBeNull();
    expect(state.announcement.text).toBe("'Planned cost by type' deleted.");
  });
});

describe('051 FR-030: undo and redo of the editor', () => {
  it('undoes the last change and says what was undone', () => {
    const state = run(
      loaded(),
      { type: 'step', id: 'w1', step: 'down', resize: false },
      { type: 'undo' },
    );
    expect(state.draft).toEqual(loaded().draft);
    expect(state.announcement.text).toBe("Undone: moved 'Planned cost by type'.");
    expect(isDirty(state)).toBe(false);
  });

  it('redoes what was undone, until a new change clears the redo list', () => {
    const moved = run(loaded(), { type: 'step', id: 'w1', step: 'down', resize: false });
    const redone = run(moved, { type: 'undo' }, { type: 'redo' });
    expect(redone.draft).toEqual(moved.draft);
    expect(redone.announcement.text).toBe("Redone: moved 'Planned cost by type'.");

    const branched = run(
      moved,
      { type: 'undo' },
      { type: 'rename', name: 'Other' },
      { type: 'redo' },
    );
    expect(branched.draft.name).toBe('Other');
    expect(branched.announcement.text).toBe('Nothing to redo.');
  });

  it(`keeps the last ${HISTORY_LIMIT} steps at least`, () => {
    const steps = Array.from({ length: 60 }, (_, index): BuilderAction => ({
      type: 'setFilters',
      filters: { dateFrom: null, dateTo: null, planIds: [`P-${index}`], teamIds: [] },
    }));
    const undos = Array.from({ length: 60 }, (): BuilderAction => ({ type: 'undo' }));
    const state = run(empty(), ...steps, ...undos);
    expect(state.draft.filters.planIds).toEqual(['P-9']);
    expect(state.announcement.text).toBe('Nothing to undo.');
  });

  it('undoes typing in one field as one step', () => {
    const state = run(
      empty(),
      { type: 'rename', name: 'L' },
      { type: 'rename', name: 'Lu' },
      { type: 'rename', name: 'Lub' },
      { type: 'undo' },
    );
    expect(state.draft.name).toBe('');
  });

  it('clears the selection when undo removes the selected widget', () => {
    const state = run(empty(), { type: 'add', widgetType: 'bar' }, { type: 'undo' });
    expect(state.selectedId).toBeNull();
    expect(state.announcement.text).toBe('Undone: added Bar chart.');
  });
});

describe('051 FR-012, FR-034: unsaved changes and saving', () => {
  it('is dirty after a change and clean again once saved', () => {
    const changed = run(loaded(), { type: 'rename', name: 'Europe' });
    expect(isDirty(changed)).toBe(true);
    const saved = run(changed, {
      type: 'saved',
      dashboard: dashboard({ name: 'Europe', version: 4 }),
    });
    expect(isDirty(saved)).toBe(false);
    expect(saved.version).toBe(4);
  });

  it('keeps the version it started from until told to follow a newer one', () => {
    expect(loaded().version).toBe(3);
    expect(run(loaded(), { type: 'followVersion', version: 5 }).version).toBe(5);
  });

  it('keeps combo measure styles and table columns in step with the measures', () => {
    const state = run(empty(), { type: 'add', widgetType: 'combo' });
    const id = state.draft.widgets[0]?.id ?? '';
    const styled = run(
      state,
      { type: 'setDisplay', id, display: { measureStyles: [{ as: 'line', axis: 'left' }] } },
      {
        type: 'setQuery',
        id,
        query: {
          ...BAR_QUERY,
          measures: [
            { aggregation: 'count', field: null, label: '' },
            { aggregation: 'sum', field: 'plannedCost', label: '' },
          ],
        },
      },
    );
    expect(styled.draft.widgets[0]?.display.measureStyles).toEqual([
      { as: 'line', axis: 'left' },
      { as: 'line', axis: 'right' },
    ]);
  });
});
