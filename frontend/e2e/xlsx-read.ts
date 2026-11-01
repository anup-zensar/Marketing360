import { inflateRawSync } from 'node:zlib';

// Reads an .xlsx's sheet names from its zip central directory: the sizes there are right even when a part was
// written with a data descriptor, as Go's archive/zip does.

const CENTRAL_ENTRY = 0x02014b50;
const LOCAL_HEADER = 30;

function part(file: Buffer, name: string): Buffer {
  for (let at = file.lastIndexOf(Buffer.from([0x50, 0x4b, 0x05, 0x06])); at > 0; at--) {
    if (file.readUInt32LE(at) !== CENTRAL_ENTRY) continue;
    const nameLength = file.readUInt16LE(at + 28);
    if (file.toString('utf8', at + 46, at + 46 + nameLength) !== name) continue;
    const method = file.readUInt16LE(at + 10);
    const size = file.readUInt32LE(at + 20);
    const local = file.readUInt32LE(at + 42);
    const start =
      local + LOCAL_HEADER + file.readUInt16LE(local + 26) + file.readUInt16LE(local + 28);
    const data = file.subarray(start, start + size);
    return method === 0 ? data : inflateRawSync(data);
  }
  throw new Error(`${name} is not in the file`);
}

export function sheetNames(file: Buffer): string[] {
  const workbook = part(file, 'xl/workbook.xml').toString('utf8');
  return [...workbook.matchAll(/<sheet [^>]*name="([^"]*)"/g)].map((match) => match[1] ?? '');
}
