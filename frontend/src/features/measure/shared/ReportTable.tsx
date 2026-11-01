import styles from './ReportWidget.module.css';

export interface ReportTableColumn {
  label: string;
  numeric?: boolean;
}

export interface ReportTableRow {
  key: string;
  cells: string[];
}

export interface ReportTableData {
  columns: ReportTableColumn[];
  rows: ReportTableRow[];
  /** The totals row, where the widget totals (FR-021). */
  total?: string[];
}

/** A widget's table view: every value the chart draws, with the same labels, then the totals (FR-021). */
export function ReportTable({ title, data }: { title: string; data: ReportTableData }) {
  return (
    <div className={styles.tableWrap}>
      <table className={styles.table}>
        <caption className={styles.caption}>{title}</caption>
        <thead>
          <tr>
            {data.columns.map((column) => (
              <th
                key={column.label}
                scope="col"
                className={column.numeric ? styles.numeric : undefined}
              >
                {column.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.rows.map((row) => (
            <tr key={row.key}>
              {row.cells.map((cell, i) => (
                <td
                  key={data.columns[i]?.label ?? i}
                  className={data.columns[i]?.numeric ? styles.numeric : undefined}
                >
                  {cell}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
        {data.total && (
          <tfoot>
            <tr className={styles.total}>
              {data.total.map((cell, i) => (
                <td
                  key={data.columns[i]?.label ?? i}
                  className={data.columns[i]?.numeric ? styles.numeric : undefined}
                >
                  {cell}
                </td>
              ))}
            </tr>
          </tfoot>
        )}
      </table>
    </div>
  );
}
