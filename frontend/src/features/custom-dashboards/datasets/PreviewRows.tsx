import type { DatasetPreview } from '../model';
import styles from './datasets.module.css';

interface PreviewRowsProps {
  preview: DatasetPreview;
  labels: string[];
}

export function PreviewRows({ preview, labels }: PreviewRowsProps) {
  if (preview.rows.length === 0) {
    return <p className={styles.facts}>The file has no rows under its header.</p>;
  }
  return (
    <div role="region" aria-label="First rows" tabIndex={0} className={styles.scroller}>
      <table className={styles.table}>
        <caption className="visually-hidden">
          The first {preview.rows.length} rows of the file
        </caption>
        <thead>
          <tr>
            {preview.columns.map((column, index) => (
              <th key={column.position} scope="col">
                {labels[index] || column.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {preview.rows.map((row, rowIndex) => (
            <tr key={rowIndex}>
              {preview.columns.map((column, index) => (
                <td key={column.position}>{row[index] ?? ''}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
