import fc from 'fast-check';
import { describe, expect, it } from 'vitest';
import { GRID_COLUMNS, GRID_ROWS } from '../model';
import type { WidgetLayout } from '../model';
import {
  blockedBy,
  firstFreePlace,
  fits,
  isInside,
  overlaps,
  placeDrop,
  rowsUsed,
  shifted,
  stepped,
} from './grid';
import type { Placed, Step } from './grid';

const size = fc.record({ w: fc.integer({ min: 2, max: 12 }), h: fc.integer({ min: 2, max: 12 }) });
const cell = fc.record({
  x: fc.integer({ min: -5, max: 20 }),
  y: fc.integer({ min: -5, max: 220 }),
});
const steps = fc.array(
  fc.record({
    step: fc.constantFrom<Step>('left', 'right', 'up', 'down'),
    resize: fc.boolean(),
  }),
  { maxLength: 40 },
);

function at(id: string, layout: WidgetLayout): Placed {
  return { id, layout };
}

function layoutIsValid(widgets: readonly Placed[]): boolean {
  return widgets.every(
    (widget, index) =>
      isInside(widget.layout) &&
      widgets.slice(index + 1).every((other) => !overlaps(widget.layout, other.layout)),
  );
}

function addAll(sizes: { w: number; h: number }[]): Placed[] {
  return sizes.reduce<Placed[]>((placed, next, index) => {
    const layout = firstFreePlace(next, placed);
    return layout ? [...placed, at(`w${index + 1}`, layout)] : placed;
  }, []);
}

describe('051 FR-013, FR-027: placing widgets on the 12-column grid', () => {
  it('puts a new widget in the first free place, left to right then down', () => {
    const placed = [at('w1', { x: 0, y: 0, w: 6, h: 4 })];
    expect(firstFreePlace({ w: 6, h: 4 }, placed)).toEqual({ x: 6, y: 0, w: 6, h: 4 });
    expect(firstFreePlace({ w: 7, h: 4 }, placed)).toEqual({ x: 0, y: 4, w: 7, h: 4 });
  });

  it('places a drop where it lands when free, kept inside the edges', () => {
    expect(placeDrop({ w: 6, h: 4 }, { x: 10, y: 3 }, [])).toEqual({ x: 6, y: 3, w: 6, h: 4 });
  });

  it('places a drop on another widget in the first free place instead', () => {
    const placed = [at('w1', { x: 0, y: 0, w: 6, h: 4 })];
    expect(placeDrop({ w: 4, h: 2 }, { x: 2, y: 1 }, placed)).toEqual({ x: 6, y: 0, w: 4, h: 2 });
  });

  it('finds no place once the grid is full to the 200th row', () => {
    const placed = Array.from({ length: 20 }, (_, index) =>
      at(`w${index + 1}`, { x: 0, y: index * 10, w: 12, h: 10 }),
    );
    expect(firstFreePlace({ w: 2, h: 2 }, placed)).toBeNull();
  });

  it('names why a place is refused: past an edge or over another widget (FR-028)', () => {
    const placed = [at('w1', { x: 0, y: 0, w: 6, h: 4 }), at('w2', { x: 6, y: 0, w: 6, h: 4 })];
    expect(blockedBy({ x: 7, y: 0, w: 6, h: 4 }, placed, 'w2')).toBe('outside');
    expect(blockedBy({ x: 5, y: 0, w: 6, h: 4 }, placed, 'w2')).toBe('overlap');
    expect(blockedBy({ x: 6, y: 4, w: 6, h: 4 }, placed, 'w2')).toBeNull();
  });

  it.each([
    [{ x: 0, y: 0, w: 1, h: 4 }, false],
    [{ x: 0, y: 0, w: 13, h: 4 }, false],
    [{ x: 0, y: 0, w: 4, h: 13 }, false],
    [{ x: -1, y: 0, w: 4, h: 4 }, false],
    [{ x: 0, y: 197, w: 4, h: 4 }, false],
    [{ x: 0.5, y: 0, w: 4, h: 4 }, false],
    [{ x: 10, y: 196, w: 2, h: 4 }, true],
  ])('keeps %o inside the grid: %s', (layout, inside) => {
    expect(isInside(layout)).toBe(inside);
  });

  it('steps one cell with the arrow keys and grows one cell with Shift (FR-028)', () => {
    const layout = { x: 2, y: 2, w: 4, h: 3 };
    expect(stepped(layout, 'right', false)).toEqual({ x: 3, y: 2, w: 4, h: 3 });
    expect(stepped(layout, 'up', false)).toEqual({ x: 2, y: 1, w: 4, h: 3 });
    expect(stepped(layout, 'right', true)).toEqual({ x: 2, y: 2, w: 5, h: 3 });
    expect(stepped(layout, 'up', true)).toEqual({ x: 2, y: 2, w: 4, h: 2 });
  });

  it('turns a pointer drag into whole cells, kept inside the grid', () => {
    const layout = { x: 2, y: 2, w: 4, h: 3 };
    expect(shifted(layout, { x: 20, y: -9 }, false)).toEqual({ x: 8, y: 0, w: 4, h: 3 });
    expect(shifted(layout, { x: 20, y: -9 }, true)).toEqual({ x: 2, y: 2, w: 10, h: 2 });
  });

  it('counts the rows the widgets use, for the canvas height', () => {
    expect(rowsUsed([])).toBe(0);
    expect(rowsUsed([at('w1', { x: 0, y: 3, w: 2, h: 5 })])).toBe(8);
  });
});

describe('051 FR-013, FR-028 (property): widgets never overlap and never leave the grid', () => {
  it('adding widgets of any size one after another', () => {
    fc.assert(
      fc.property(fc.array(size, { maxLength: 30 }), (sizes) => {
        expect(layoutIsValid(addAll(sizes))).toBe(true);
      }),
    );
  });

  it('dropping widgets anywhere, inside the edges or not', () => {
    fc.assert(
      fc.property(fc.array(fc.tuple(size, cell), { maxLength: 30 }), (drops) => {
        const placed = drops.reduce<Placed[]>((all, [next, where], index) => {
          const layout = placeDrop(next, where, all);
          return layout ? [...all, at(`w${index}`, layout)] : all;
        }, []);
        expect(layoutIsValid(placed)).toBe(true);
      }),
    );
  });

  it('moving and resizing with the keyboard, applying only the steps that fit', () => {
    fc.assert(
      fc.property(fc.array(size, { minLength: 1, maxLength: 12 }), steps, (sizes, keys) => {
        let placed = addAll(sizes);
        const id = placed[0]?.id ?? '';
        for (const { step, resize } of keys) {
          const current = placed.find((widget) => widget.id === id);
          const next = current && stepped(current.layout, step, resize);
          if (next && fits(next, placed, id)) {
            placed = placed.map((widget) => (widget.id === id ? at(id, next) : widget));
          }
        }
        expect(layoutIsValid(placed)).toBe(true);
      }),
    );
  });

  it('a pointer drag of any length stays inside the grid', () => {
    fc.assert(
      fc.property(size, cell, fc.boolean(), (start, cells, resize) => {
        const moved = shifted({ x: 0, y: 0, ...start }, cells, resize);
        expect(moved.x + moved.w).toBeLessThanOrEqual(GRID_COLUMNS);
        expect(moved.y + moved.h).toBeLessThanOrEqual(GRID_ROWS);
        expect(isInside(moved)).toBe(true);
      }),
    );
  });
});
