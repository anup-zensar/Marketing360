import { act, fireEvent, screen, within } from '@testing-library/react';
import { useReducer } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { reply } from '@/testing/replies';
import { testUser } from '@/testing/fixtures/users';
import { widgetData } from '../charts/widget-fixtures';
import type { CustomDashboard } from '../model';
import { dashboard } from './builder-fixtures';
import { builderReducer, initialState } from './builder-reducer';
import { Canvas } from './Canvas';
import { WIDGET_MIME } from './canvas-dom';
import { WidgetPalette } from './WidgetPalette';

const GRID_WIDTH = 1184;

vi.mock('echarts/core', async (importOriginal) => ({
  ...(await importOriginal<object>()),
  init: () => ({ setOption: () => {}, resize: () => {}, dispose: () => {} }),
}));

function Harness({ start }: { start: CustomDashboard | null }) {
  const [state, dispatch] = useReducer(builderReducer, start, initialState);
  return (
    <>
      <WidgetPalette onAdd={(widgetType) => dispatch({ type: 'add', widgetType })} />
      <Canvas
        draft={state.draft}
        selectedId={state.selectedId}
        problems={[]}
        announcement={state.announcement}
        onAction={dispatch}
      />
    </>
  );
}

function renderCanvas(start: CustomDashboard | null = dashboard()) {
  mockApi([reply('POST /api/v1/data-sources/query', { body: widgetData() })]);
  return renderWithProviders(<Harness start={start} />, { user: testUser() });
}

function cell(name: string): RegExp {
  return new RegExp(`^Widget: ${name}, column \\d+, row \\d+$`);
}

function widget(name: string) {
  return screen.getByRole('group', { name: cell(name) });
}

function handle(name: string) {
  return within(widget(name)).getByRole('button', { name });
}

function announced() {
  const outsideWidgets = screen
    .getAllByRole('status')
    .filter((status) => !status.closest('[role="group"]'));
  return outsideWidgets[0]?.textContent;
}

beforeEach(() => {
  Element.prototype.setPointerCapture = () => {};
  vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue(
    DOMRect.fromRect({ x: 0, y: 0, width: GRID_WIDTH, height: 800 }),
  );
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe('051 FR-028: moving and resizing widgets with the keyboard', () => {
  it('moves the focused widget one cell with an arrow key and announces it', async () => {
    const { user } = renderCanvas();
    handle('About').focus();
    await user.keyboard('{ArrowDown}');

    expect(widget('About')).toHaveStyle({ gridColumn: '7 / span 4', gridRow: '2 / span 2' });
    expect(announced()).toBe('Moved to column 7, row 2.');
  });

  it('names each widget once, with its place on the grid (SC-006)', () => {
    renderCanvas();
    expect(
      screen.getByRole('group', { name: 'Widget: About, column 7, row 1' }),
    ).toBeInTheDocument();
    expect(screen.getAllByRole('region', { name: 'Planned cost by type' })).toHaveLength(1);
  });

  it('resizes one cell with Shift and an arrow key', async () => {
    const { user } = renderCanvas();
    handle('About').focus();
    await user.keyboard('{Shift>}{ArrowRight}{/Shift}');

    expect(widget('About')).toHaveStyle({ gridColumn: '7 / span 5' });
    expect(announced()).toBe('Resized to 5 columns by 2 rows.');
  });

  it('resizes with the arrow keys on the resize handle', async () => {
    const { user } = renderCanvas();
    screen.getByRole('button', { name: "Resize 'About'" }).focus();
    await user.keyboard('{ArrowDown}');

    expect(widget('About')).toHaveStyle({ gridRow: '1 / span 3' });
  });

  it('never moves over another widget and says why', async () => {
    const { user } = renderCanvas();
    handle('Planned cost by type').focus();
    await user.keyboard('{ArrowRight}');

    expect(widget('Planned cost by type')).toHaveStyle({ gridColumn: '1 / span 6' });
    expect(announced()).toBe("Can't move there: another widget is in the way.");
  });

  it('selects a widget when it gets the focus and deselects it with Escape', async () => {
    const { user } = renderCanvas();
    act(() => handle('Planned cost by type').focus());
    expect(handle('Planned cost by type')).toHaveAttribute('aria-pressed', 'true');

    await user.keyboard('{Escape}');
    expect(handle('Planned cost by type')).toHaveAttribute('aria-pressed', 'false');
  });
});

describe('051 FR-029: duplicating and deleting from the keyboard and the menu', () => {
  it('duplicates with Ctrl+D into the next free place', async () => {
    const { user } = renderCanvas();
    handle('About').focus();
    await user.keyboard('{Control>}d{/Control}');

    expect(screen.getAllByRole('group', { name: cell('About') })).toHaveLength(2);
    expect(announced()).toBe("'About' duplicated.");
  });

  it('deletes with the Delete key and puts the focus on the canvas', async () => {
    const { user } = renderCanvas();
    handle('About').focus();
    await user.keyboard('{Delete}');

    expect(screen.queryByRole('group', { name: cell('About') })).not.toBeInTheDocument();
    expect(screen.getByRole('region', { name: 'Canvas' })).toHaveFocus();
    expect(announced()).toBe("'About' deleted.");
  });

  it('duplicates and deletes from the widget menu', async () => {
    const { user } = renderCanvas();
    await user.click(screen.getByRole('button', { name: "Actions for 'About'" }));
    await user.click(await screen.findByRole('menuitem', { name: 'Delete' }));

    expect(screen.queryByRole('group', { name: cell('About') })).not.toBeInTheDocument();
  });
});

describe('051 FR-028: moving and resizing with the pointer', () => {
  it('moves a widget dragged by its title to the cell under the pointer', () => {
    renderCanvas();
    const grip = handle('About');
    fireEvent.pointerDown(grip, { pointerId: 1, button: 0, clientX: 0, clientY: 0 });
    fireEvent.pointerMove(grip, { pointerId: 1, clientX: 200, clientY: 192 });
    fireEvent.pointerUp(grip, { pointerId: 1, clientX: 200, clientY: 192 });

    expect(widget('About')).toHaveStyle({ gridColumn: '9 / span 4', gridRow: '3 / span 2' });
    expect(announced()).toBe('Moved to column 9, row 3.');
  });

  it('resizes a widget dragged by its corner handle', () => {
    renderCanvas();
    const corner = screen.getByRole('button', { name: "Resize 'About'" });
    fireEvent.pointerDown(corner, { pointerId: 1, button: 0, clientX: 0, clientY: 0 });
    fireEvent.pointerUp(corner, { pointerId: 1, clientX: 100, clientY: 96 });

    expect(widget('About')).toHaveStyle({ gridColumn: '7 / span 5', gridRow: '1 / span 3' });
    expect(announced()).toBe('Resized to 5 columns by 3 rows.');
  });

  it('leaves a widget in place when dropped on another one', () => {
    renderCanvas();
    const grip = handle('About');
    fireEvent.pointerDown(grip, { pointerId: 1, button: 0, clientX: 0, clientY: 0 });
    fireEvent.pointerUp(grip, { pointerId: 1, clientX: -300, clientY: 0 });

    expect(widget('About')).toHaveStyle({ gridColumn: '7 / span 4' });
    expect(announced()).toBe("Can't move there: another widget is in the way.");
  });
});

describe('051 FR-027: adding from the palette', () => {
  it('adds a clicked type in the first free place, says so and selects it', async () => {
    const { user } = renderCanvas(null);
    expect(screen.getByText(/Add a widget: click a type in the palette/)).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: 'Add bar' }));

    expect(widget('Bar chart')).toHaveStyle({ gridColumn: '1 / span 6', gridRow: '1 / span 4' });
    expect(handle('Bar chart')).toHaveAttribute('aria-pressed', 'true');
    expect(announced()).toBe('Bar chart added.');
  });

  it('adds a dragged type where it is dropped', () => {
    vi.stubGlobal('DragEvent', class extends MouseEvent {});
    renderCanvas(null);
    const types = [WIDGET_MIME, `${WIDGET_MIME}-kpi`];
    const dataTransfer = { types, getData: () => 'kpi', dropEffect: 'none' };
    const canvas = screen.getByRole('region', { name: 'Canvas' });

    fireEvent.dragOver(canvas, { dataTransfer, clientX: 450, clientY: 200 });
    fireEvent.drop(canvas, { dataTransfer, clientX: 450, clientY: 200 });

    expect(widget('Key figure')).toHaveStyle({ gridColumn: '5 / span 3', gridRow: '3 / span 2' });
  });

  it('has no accessibility violations (SC-006)', async () => {
    renderCanvas();
    await expectNoAxeViolations();
  });
});
