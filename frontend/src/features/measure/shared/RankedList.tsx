import { Link, useSearchParams } from 'react-router';
import { recordHref } from '@/ui/record-links';
import type { PlanningRecordType } from '@/ui/record-links';
import { Button } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { StatusChip } from '@/ui/StatusChip/StatusChip';
import type { ReportKey } from '../reports';
import { downloadReport } from './api';
import { ExportMenu } from './ExportMenu';
import { Paging, RankedControls } from './RankedControls';
import type { Choice } from './RankedControls';
import { useRankedState } from './ranked-params';
import { useReportAddress } from './useReportAddress';
import styles from './RankedList.module.css';

export interface RankedEntryData<F> {
  id: string;
  level: string;
  name: string;
  planId: string;
  figures: F;
  children?: RankedEntryData<F>[];
}

export interface RankedData<F> {
  total: number;
  trail: { id: string; level: string; name: string }[];
  groups: { key: string; label: string; totals: F; entries: RankedEntryData<F>[] }[];
  drillLost: boolean;
}

export interface RankedColumn<F> {
  key: string;
  label: string;
  value: (figures: F) => string;
}

export interface RankedListProps<F> {
  report: ReportKey;
  data: RankedData<F> | undefined;
  /** The list's load failure, shown with Try again; null while it is fine. */
  error: unknown;
  onRetry: () => void;
  columns: RankedColumn<F>[];
  groups: Choice[];
  defaultSort: string;
  /** A word on an entry the reader must notice, such as "Over-allocated". */
  flag?: (figures: F) => string | undefined;
}

const DRILLS_TO = new Set(['plan', 'campaign', 'initiative', 'activity']);

function isRecordType(level: string): level is PlanningRecordType {
  return DRILLS_TO.has(level) || level === 'lineItem';
}

/** The ranked list (040 FR-008 – FR-015): top entries with their children, drill with a trail, shared with 049. */
export function RankedList<F>(props: RankedListProps<F>) {
  const { report, data, error, onRetry, columns, groups, defaultSort, flag } = props;
  const { state, change } = useRankedState(defaultSort);
  const [address] = useReportAddress();
  const [params] = useSearchParams();
  const sorts = columns.map((c) => ({ value: c.key, label: c.label }));
  const back = data && data.trail.length > 1 ? (data.trail.at(-2)?.id ?? '') : '';
  return (
    <Card
      title="Ranked list"
      actions={
        <ExportMenu
          label="Export"
          formats={['xlsx', 'csv']}
          download={(format) =>
            downloadReport(report, address, format, { widget: 'ranked', own: params })
          }
        />
      }
    >
      <RankedControls state={state} sorts={sorts} groups={groups} onChange={change} />
      {data && data.trail.length > 0 && (
        <nav className={styles.trail} aria-label="Drill trail">
          <Button size="sm" variant="tertiary" onClick={() => change({ drill: back })}>
            Back
          </Button>
          {data.trail.map((step, i) =>
            i < data.trail.length - 1 ? (
              <Button
                key={step.id}
                size="sm"
                variant="tertiary"
                onClick={() => change({ drill: step.id })}
              >
                {step.name}
              </Button>
            ) : (
              <span key={step.id} aria-current="location">
                {step.name}
              </span>
            ),
          )}
        </nav>
      )}
      {data?.drillLost && (
        <p className={styles.note}>
          The record you opened is no longer in view, so the list starts at the top.
        </p>
      )}
      {error ? (
        <ErrorState error={error} onRetry={onRetry} />
      ) : data ? (
        <RankedTable
          data={data}
          columns={columns}
          sort={state}
          flag={flag}
          onDrill={(id) => change({ drill: id })}
        />
      ) : (
        <Skeleton lines={5} height="24px" />
      )}
      {data && <Paging state={state} total={data.total} onChange={change} />}
    </Card>
  );
}

interface TableProps<F> {
  data: RankedData<F>;
  columns: RankedColumn<F>[];
  sort: { sort: string; order: 'desc' | 'asc' };
  flag?: (figures: F) => string | undefined;
  onDrill: (id: string) => void;
}

function RankedTable<F>({ data, columns, sort, flag, onDrill }: TableProps<F>) {
  if (data.groups.length === 0) {
    return <p className={styles.note}>Nothing to rank on this page.</p>;
  }
  return (
    <div className={styles.tableWrap}>
      <table className={styles.table}>
        <thead>
          <tr>
            <th scope="col">Name</th>
            {columns.map((c) => (
              <th
                key={c.key}
                scope="col"
                className={c.key === sort.sort ? styles.sorted : styles.numeric}
                aria-sort={
                  c.key === sort.sort
                    ? sort.order === 'asc'
                      ? 'ascending'
                      : 'descending'
                    : undefined
                }
              >
                {c.label}
              </th>
            ))}
          </tr>
        </thead>
        {data.groups.map((g) => (
          <tbody key={g.key}>
            {g.label && (
              <tr className={styles.groupRow}>
                <th scope="rowgroup">{g.label}</th>
                {columns.map((c) => (
                  <td key={c.key} className={styles.numeric}>
                    {c.value(g.totals)}
                  </td>
                ))}
              </tr>
            )}
            {g.entries.map((e) => (
              <Rows
                key={`${g.key}-${e.id}`}
                entry={e}
                columns={columns}
                flag={flag}
                onDrill={onDrill}
              />
            ))}
          </tbody>
        ))}
      </table>
    </div>
  );
}

interface RowsProps<F> {
  entry: RankedEntryData<F>;
  columns: RankedColumn<F>[];
  flag?: (figures: F) => string | undefined;
  onDrill: (id: string) => void;
}

function Rows<F>({ entry, columns, flag, onDrill }: RowsProps<F>) {
  const row = (e: RankedEntryData<F>, child: boolean) => {
    const word = flag?.(e.figures);
    const href = isRecordType(e.level) ? recordHref(e.level, e.id, e.planId) : undefined;
    return (
      <tr key={e.id} className={child ? styles.child : styles.top}>
        <th scope="row">
          {child && DRILLS_TO.has(e.level) ? (
            <Button size="sm" variant="tertiary" onClick={() => onDrill(e.id)}>
              {e.name}
            </Button>
          ) : (
            <span className={styles.name}>{e.name}</span>
          )}
          {href && (
            <Link className={styles.open} to={href} aria-label={`Open ${e.name} in Plan`}>
              Open
            </Link>
          )}
          {word && <StatusChip tone="problem">{word}</StatusChip>}
        </th>
        {columns.map((c) => (
          <td key={c.key} className={styles.numeric}>
            {c.value(e.figures)}
          </td>
        ))}
      </tr>
    );
  };
  return (
    <>
      {row(entry, false)}
      {entry.children?.map((child) => row(child, true))}
    </>
  );
}
