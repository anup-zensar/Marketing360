import { ArrowDown, ArrowUp, ArrowUpDown, ChevronLeft, ChevronRight } from 'lucide-react';
import { useId } from 'react';
import type { ReactNode } from 'react';
import { useSearchParams } from 'react-router';
import { formatWhole } from '@/format/decimal';
import { Button } from '@/ui/Button/Button';
import { cx } from '@/ui/cx';
import { useReportAddress } from '../shared';
import { downloadReport } from '../shared/api';
import { ExportMenu } from '../shared/ExportMenu';
import { PAGE_SIZES } from './gsap-address';
import type { GsapTable } from './gsap-address';
import styles from './GsapLink.module.css';

export interface PagedColumn<R> {
  /** The column's sort key in the contract. */
  key: string;
  label: string;
  numeric?: boolean;
  /** Waiting, dates and amounts sort largest first on the first click. */
  largestFirst?: boolean;
  cell: (row: R) => ReactNode;
}

export interface PagedTableProps<R> {
  title: string;
  /** The widget's key in the export (contract: swbs, waiting, orders). */
  widget: string;
  columns: PagedColumn<R>[];
  rows: R[];
  rowKey: (row: R) => string;
  table: GsapTable;
  total: number;
  /** Said when the tab has no row at all. */
  empty: string;
  /** The totals row over every page (FR-013), one cell per column. */
  totals?: ReactNode[];
  onChange: (patch: Partial<GsapTable>) => void;
}

/** A tab's table card: its title and export of every row (FR-019), then the page of rows. */
export function PagedTable<R>(props: PagedTableProps<R>) {
  const { title, widget } = props;
  const headingId = useId();
  const [address] = useReportAddress();
  const [own] = useSearchParams();
  return (
    <section className={styles.card} aria-labelledby={headingId}>
      <div className={styles.cardHeader}>
        <h2 id={headingId} className={styles.heading}>
          {title}
        </h2>
        <ExportMenu
          label="Export"
          formats={['xlsx', 'csv']}
          download={(format) => downloadReport('gsapLink', address, format, { widget, own })}
        />
      </div>
      <Rows {...props} />
    </section>
  );
}

/** One page of a tab's rows, sorted by any column, with the page size and the pages (044 FR-015). */
function Rows<R>(props: PagedTableProps<R>) {
  const { title, columns, rows, rowKey, table, total, empty, totals, onChange } = props;
  if (total === 0) {
    return <p className={styles.empty}>{empty}</p>;
  }
  return (
    <>
      <div className={styles.scroll}>
        <table className={styles.table}>
          <caption className="visually-hidden">{title}</caption>
          <thead>
            <tr>
              {columns.map((column) => (
                <SortHeading
                  key={column.key}
                  column={column}
                  sort={table.sort}
                  onSort={(sort) => onChange({ sort })}
                />
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row) => (
              <tr key={rowKey(row)}>
                {columns.map((column) => (
                  <td key={column.key} className={cx(column.numeric && styles.numeric)}>
                    {column.cell(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
          {totals && (
            <tfoot>
              <tr>
                {totals.map((cell, i) => (
                  <td
                    key={columns[i]?.key ?? i}
                    className={cx(columns[i]?.numeric && styles.numeric)}
                  >
                    {cell}
                  </td>
                ))}
              </tr>
            </tfoot>
          )}
        </table>
      </div>
      <Pager table={table} total={total} onChange={onChange} />
    </>
  );
}

interface SortHeadingProps<R> {
  column: PagedColumn<R>;
  sort: string;
  onSort: (sort: string) => void;
}

/** A heading that sorts by its column; a second click turns the order round. */
function SortHeading<R>({ column, sort, onSort }: SortHeadingProps<R>) {
  const current = sort.replace(/^-/, '') === column.key;
  const descending = sort.startsWith('-');
  const Icon = !current ? ArrowUpDown : descending ? ArrowDown : ArrowUp;
  const next = current ? !descending : Boolean(column.largestFirst);
  return (
    <th
      scope="col"
      className={cx(column.numeric && styles.numeric)}
      aria-sort={current ? (descending ? 'descending' : 'ascending') : 'none'}
    >
      <button
        type="button"
        className={styles.sort}
        onClick={() => onSort(`${next ? '-' : ''}${column.key}`)}
      >
        {column.label}
        <Icon aria-hidden="true" className={current ? undefined : styles.idle} />
      </button>
    </th>
  );
}

interface PagerProps {
  table: GsapTable;
  total: number;
  onChange: (patch: Partial<GsapTable>) => void;
}

/** "21–40 of 45" with Previous and Next, and the rows per page. */
function Pager({ table, total, onChange }: PagerProps) {
  const id = useId();
  const { page, pageSize } = table;
  const pages = Math.max(1, Math.ceil(total / pageSize));
  const first = Math.min(total, (page - 1) * pageSize + 1);
  const last = Math.min(total, page * pageSize);
  return (
    <nav className={styles.pager} aria-label="Pages">
      <div className={styles.size}>
        <label htmlFor={id}>Per page</label>
        <select
          id={id}
          className={styles.select}
          value={pageSize}
          onChange={(event) => {
            const size = PAGE_SIZES.find((s) => String(s) === event.target.value);
            if (size !== undefined) {
              onChange({ pageSize: size });
            }
          }}
        >
          {PAGE_SIZES.map((size) => (
            <option key={size} value={size}>
              {size}
            </option>
          ))}
        </select>
      </div>
      <p className={styles.range} aria-live="polite">
        {`${formatWhole(first)}–${formatWhole(last)} of ${formatWhole(total)}`}
      </p>
      <div className={styles.buttons}>
        <Button
          size="sm"
          icon={<ChevronLeft />}
          disabled={page <= 1}
          onClick={() => onChange({ page: page - 1 })}
        >
          Previous
        </Button>
        <Button size="sm" disabled={page >= pages} onClick={() => onChange({ page: page + 1 })}>
          Next
          <ChevronRight aria-hidden="true" />
        </Button>
      </div>
    </nav>
  );
}
