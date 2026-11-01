import { afterEach, describe, expect, it } from 'vitest';
import { pngFileName, svgToPng } from './png';

interface Drawn {
  width: number;
  height: number;
  fills: string[];
  images: number;
}

const ORIGINAL = {
  getContext: Object.getOwnPropertyDescriptor(HTMLCanvasElement.prototype, 'getContext'),
  toBlob: Object.getOwnPropertyDescriptor(HTMLCanvasElement.prototype, 'toBlob'),
  decode: Object.getOwnPropertyDescriptor(HTMLImageElement.prototype, 'decode'),
};

function replace(target: object, name: string, value: unknown) {
  Object.defineProperty(target, name, { configurable: true, writable: true, value });
}

function restore(target: object, name: string, descriptor: PropertyDescriptor | undefined) {
  if (descriptor) {
    Object.defineProperty(target, name, descriptor);
  } else {
    Reflect.deleteProperty(target, name);
  }
}

/** jsdom draws nothing: a canvas that records what is drawn on it. */
function stubCanvas(hasContext = true): Drawn {
  const drawn: Drawn = { width: 0, height: 0, fills: [], images: 0 };
  replace(HTMLImageElement.prototype, 'decode', () => Promise.resolve());
  replace(HTMLCanvasElement.prototype, 'getContext', function (this: HTMLCanvasElement) {
    if (!hasContext) {
      return null;
    }
    drawn.width = this.width;
    drawn.height = this.height;
    const context = {
      fillStyle: '',
      fillRect: () => drawn.fills.push(context.fillStyle),
      drawImage: () => (drawn.images += 1),
    };
    return context;
  });
  replace(HTMLCanvasElement.prototype, 'toBlob', (done: (blob: Blob | null) => void) =>
    done(new Blob(['png'], { type: 'image/png' })),
  );
  return drawn;
}

describe('051 FR-038: Download as PNG', () => {
  afterEach(() => {
    restore(HTMLCanvasElement.prototype, 'getContext', ORIGINAL.getContext);
    restore(HTMLCanvasElement.prototype, 'toBlob', ORIGINAL.toBlob);
    restore(HTMLImageElement.prototype, 'decode', ORIGINAL.decode);
  });

  it('draws the chart at twice its size on the surface colour', async () => {
    const drawn = stubCanvas();
    const blob = await svgToPng({
      svgUrl: 'data:image/svg+xml;charset=UTF-8,%3Csvg%3E%3C%2Fsvg%3E',
      width: 400,
      height: 300,
      background: '#ffffff',
    });
    expect(blob.type).toBe('image/png');
    expect(drawn).toEqual({ width: 800, height: 600, fills: ['#ffffff'], images: 1 });
  });

  it('says so when the browser cannot draw', async () => {
    stubCanvas(false);
    await expect(
      svgToPng({ svgUrl: 'data:image/svg+xml,', width: 1, height: 1, background: '' }),
    ).rejects.toThrow('This browser cannot draw the picture.');
  });

  it.each([
    ['Market view', 'Planned cost by type', 'SP&A One – Market view – Planned cost by type.png'],
    ['Q1: plans/teams', 'Cost?', 'SP&A One – Q1- plans-teams – Cost-.png'],
    ['Market view', '', 'SP&A One – Market view.png'],
    ['', '', 'SP&A One – Chart.png'],
  ])(
    '051 US2-6: names the file "SP&A One – dashboard – widget": %s, %s',
    (dashboard, title, expected) => {
      expect(pngFileName(dashboard, title)).toBe(expected);
    },
  );
});
