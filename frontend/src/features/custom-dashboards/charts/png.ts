const SCALE = 2;
const UNSAFE_IN_FILE_NAMES = /[\\/:*?"<>|]+/g;

export interface PngSource {
  svgUrl: string;
  width: number;
  height: number;
  background: string;
}

/** The chart's SVG drawn at twice its size on the surface colour; nothing leaves the machine. */
export async function svgToPng({ svgUrl, width, height, background }: PngSource): Promise<Blob> {
  const image = new Image();
  image.src = svgUrl;
  await image.decode();
  const canvas = document.createElement('canvas');
  canvas.width = width * SCALE;
  canvas.height = height * SCALE;
  const context = canvas.getContext('2d');
  if (!context) {
    throw new Error('This browser cannot draw the picture.');
  }
  context.fillStyle = background;
  context.fillRect(0, 0, canvas.width, canvas.height);
  context.drawImage(image, 0, 0, canvas.width, canvas.height);
  return new Promise((resolve, reject) => {
    canvas.toBlob(
      (blob) => (blob ? resolve(blob) : reject(new Error('The picture could not be made.'))),
      'image/png',
    );
  });
}

export function pngFileName(dashboardName: string, widgetTitle: string): string {
  const name = [dashboardName, widgetTitle]
    .map((part) => part.replace(UNSAFE_IN_FILE_NAMES, '-').trim())
    .filter(Boolean)
    .join(' - ');
  return `${name || 'Chart'}.png`;
}
