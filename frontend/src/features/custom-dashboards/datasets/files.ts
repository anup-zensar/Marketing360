export const MAX_FILE_BYTES = 5 * 1024 * 1024;
export const LIMITS_TEXT = 'CSV or Excel .xlsx, up to 5 MB, 10,000 rows, 50 columns.';

export function fileProblem(file: File | null): string | null {
  if (!file) {
    return 'Choose a CSV or Excel .xlsx file.';
  }
  if (!/\.(xlsx|csv)$/i.test(file.name)) {
    return 'SP&A One can read .xlsx and CSV (UTF-8) files.';
  }
  if (file.size > MAX_FILE_BYTES) {
    return 'The file is larger than 5 MB.';
  }
  return null;
}

export function base64Of(file: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const url = String(reader.result);
      resolve(url.slice(url.indexOf(',') + 1));
    };
    reader.onerror = () => reject(reader.error ?? new Error('The file could not be read.'));
    reader.readAsDataURL(file);
  });
}

export function nameFromFile(fileName: string): string {
  return fileName.replace(/\.[^.]+$/, '').slice(0, 80);
}
