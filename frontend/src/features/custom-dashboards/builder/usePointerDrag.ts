import { useRef } from 'react';
import type { PointerEvent } from 'react';
import type { WidgetLayout } from '../model';
import { shifted } from './grid';

export interface Pitch {
  col: number;
  row: number;
}

interface DragOptions {
  layout: WidgetLayout;
  resize: boolean;
  pitch: () => Pitch;
  onPreview: (layout: WidgetLayout | null) => void;
  onDrop: (layout: WidgetLayout) => void;
}

interface Start {
  pointerId: number;
  clientX: number;
  clientY: number;
  pitch: Pitch;
}

export function usePointerDrag({ layout, resize, pitch, onPreview, onDrop }: DragOptions) {
  const start = useRef<Start | null>(null);

  function target(event: PointerEvent<HTMLElement>): WidgetLayout | null {
    const from = start.current;
    if (!from || from.pointerId !== event.pointerId) {
      return null;
    }
    const cells = {
      x: Math.round((event.clientX - from.clientX) / from.pitch.col),
      y: Math.round((event.clientY - from.clientY) / from.pitch.row),
    };
    return shifted(layout, cells, resize);
  }

  function onPointerDown(event: PointerEvent<HTMLElement>) {
    if (event.button !== 0) {
      return;
    }
    event.currentTarget.setPointerCapture(event.pointerId);
    start.current = {
      pointerId: event.pointerId,
      clientX: event.clientX,
      clientY: event.clientY,
      pitch: pitch(),
    };
  }

  function onPointerMove(event: PointerEvent<HTMLElement>) {
    const next = target(event);
    if (next) {
      onPreview(next);
    }
  }

  function onPointerUp(event: PointerEvent<HTMLElement>) {
    const next = target(event);
    start.current = null;
    onPreview(null);
    if (
      next &&
      (next.x !== layout.x || next.y !== layout.y || next.w !== layout.w || next.h !== layout.h)
    ) {
      onDrop(next);
    }
  }

  function onPointerCancel() {
    start.current = null;
    onPreview(null);
  }

  return { onPointerDown, onPointerMove, onPointerUp, onPointerCancel };
}
