import { ArrowDown, ArrowUp, ArrowUpDown } from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router';
import { useFormatters } from '@/format/useFormatters';
import { Button } from '@/ui/Button/Button';
import { ReportWidget } from '../shared';
import type { ReportTableData } from '../shared';
import type { ApprovalsGridData, ApprovalsRowData } from './api';
import { CommentDialog } from './CommentDialog';
import { GridFilters } from './GridFilters';
import { GridPager } from './GridPager';
import { PAGE_SIZES } from './schemas';
import type { ApprovalsParams, SortKey } from './schemas';
import { NOTHING, OUTCOME_WORDS, recordWord } from './words';
import styles from './ApprovalsGrid.module.css';

const COLUMNS: { key: SortKey; label: string; numeric?: boolean }[] = [
  { key: 'itemId', label: 'Record ID' },
  { key: 'itemName', label: 'Record name' },
  { key: 'recordType', label: 'Record type' },
  { key: 'approver', label: 'Approver' },
  { key: 'decidedOn', label: 'Decided on' },
  { key: 'outcome', label: 'Status' },
  { key: 'submittedOn', label: 'Submitted on' },
  { key: 'level', label: 'Level', numeric: true },
  { key: 'purchaseOrder', label: 'Purchase order' },
  { key: 'amount', label: 'Amount', numeric: true },
  { key: 'owner', label: 'Owner' },
  { key: 'comments', label: 'Comments' },
];

const TIP =
  'One row per named approver at each level of each submission. Others at a level read Not needed once ' +
  'someone decided. Amounts are the activity’s planned cost now, or the purchase request’s amount.';
const LONG_COMMENT = 60;
const NEEDS_FINANCE = 'You can’t see financial data for these plans, so amounts show —.';

export interface ApprovalsGridProps {
  grid?: ApprovalsGridData;
  covers?: { plans: number; of: number };
  params: ApprovalsParams;
  onParams: (changes: Partial<ApprovalsParams>) => void;
  failed?: boolean;
  onRetry?: () => void;
}

function financeNote(covers?: { plans: number; of: number }): string | undefined {
  if (!covers || covers.of === 0 || covers.plans === covers.of) {
    return undefined;
  }
  return covers.plans === 0
    ? NEEDS_FINANCE
    : `Amounts cover ${covers.plans} of ${covers.of} plans; the rest need financial data access.`;
}

function recordHref(row: ApprovalsRowData): string {
  return row.recordType === 'activity'
    ? `/activities/${row.itemId}`
    : `/purchase-requests/${row.itemId}`;
}

function SortHeading(props: {
  column: (typeof COLUMNS)[number];
  params: ApprovalsParams;
  onParams: ApprovalsGridProps['onParams'];
}) {
  const { column, params, onParams } = props;
  const current = params.sort === column.key;
  const Icon = !current ? ArrowUpDown : params.order === 'asc' ? ArrowUp : ArrowDown;
  const sortBy = () =>
    onParams({ sort: column.key, order: current && params.order === 'asc' ? 'desc' : 'asc' });
  return (
    <th
      scope="col"
      className={column.numeric ? styles.numeric : undefined}
      aria-sort={current ? (params.order === 'asc' ? 'ascending' : 'descending') : 'none'}
    >
      <button type="button" className={styles.sort} onClick={sortBy}>
        {column.label}
        <Icon aria-hidden="true" className={current ? undefined : styles.idle} />
      </button>
    </th>
  );
}

function Row({ row, onRead }: { row: ApprovalsRowData; onRead: (row: ApprovalsRowData) => void }) {
  const { date, money } = useFormatters();
  const long = row.comments.length > LONG_COMMENT;
  return (
    <tr>
      <td>{row.itemId}</td>
      <th scope="row" className={styles.name}>
        {row.canOpen ? <Link to={recordHref(row)}>{row.itemName}</Link> : row.itemName}
      </th>
      <td>{recordWord(row.recordType)}</td>
      <td>{row.approver?.name ?? NOTHING}</td>
      <td>{row.decidedOn ? date(row.decidedOn) : NOTHING}</td>
      <td>{OUTCOME_WORDS[row.outcome] ?? row.outcome}</td>
      <td>{date(row.submittedOn)}</td>
      <td className={styles.numeric}>{row.level}</td>
      <td>{row.purchaseOrderNumber ?? NOTHING}</td>
      <td className={styles.numeric}>{row.amount ? money(row.amount) : NOTHING}</td>
      <td>{row.owner.name}</td>
      <td className={styles.commentCell}>
        {row.comments === ''
          ? NOTHING
          : long
            ? `${row.comments.slice(0, LONG_COMMENT)}… `
            : row.comments}
        {long && (
          <Button type="button" size="sm" variant="tertiary" onClick={() => onRead(row)}>
            Read in full
          </Button>
        )}
      </td>
    </tr>
  );
}

function tableOf(grid: ApprovalsGridData): ReportTableData {
  return {
    columns: COLUMNS.map((c) => ({ label: c.label, numeric: c.numeric })),
    rows: grid.items.map((row, i) => ({
      key: `${row.itemId}-${row.level}-${i}`,
      cells: [
        row.itemId,
        row.itemName,
        recordWord(row.recordType),
        row.approver?.name ?? NOTHING,
        row.decidedOn ?? NOTHING,
        OUTCOME_WORDS[row.outcome] ?? row.outcome,
        row.submittedOn,
        String(row.level),
        row.purchaseOrderNumber ?? NOTHING,
        row.amount ? `${row.amount.amount} ${row.amount.currency}` : NOTHING,
        row.owner.name,
        row.comments || NOTHING,
      ],
    })),
  };
}

function Empty({ filtered, onReset }: { filtered: boolean; onReset: () => void }) {
  return filtered ? (
    <div className={styles.empty}>
      <p>Your filters exclude everything.</p>
      <Button type="button" variant="secondary" size="sm" onClick={onReset}>
        Reset filters
      </Button>
    </div>
  ) : (
    <p className={styles.empty}>
      No approvals yet. Rows appear once a record is submitted for approval.
    </p>
  );
}

export function ApprovalsGrid({
  grid,
  covers,
  params,
  onParams,
  failed,
  onRetry,
}: ApprovalsGridProps) {
  const [reading, setReading] = useState<ApprovalsRowData | null>(null);
  const filtered = Boolean(
    params.q || params.recordTypes.length || params.approvers.length || params.outcomes.length,
  );
  const reset = () => onParams({ q: '', recordTypes: [], approvers: [], outcomes: [] });
  const note = financeNote(covers);
  const empty: ReportTableData = { columns: [{ label: 'Record ID' }], rows: [] };
  return (
    <>
      <ReportWidget
        report="approvals"
        widget="approvals"
        title="Approvals"
        tip={TIP}
        note={note}
        loading={!grid && !failed}
        failed={failed}
        onRetry={onRetry}
        table={grid ? tableOf(grid) : empty}
        chart={
          grid && (
            <>
              <GridFilters choices={grid.choices} params={params} onChange={onParams} />
              {grid.total === 0 ? (
                <Empty filtered={filtered} onReset={reset} />
              ) : (
                <div className={styles.scroll}>
                  <table className={styles.table}>
                    <caption className="visually-hidden">
                      Approvals, one row per approver step
                    </caption>
                    <thead>
                      <tr>
                        {COLUMNS.map((c) => (
                          <SortHeading key={c.key} column={c} params={params} onParams={onParams} />
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {grid.items.map((row, i) => (
                        <Row
                          key={`${row.itemId}-${row.level}-${i}`}
                          row={row}
                          onRead={setReading}
                        />
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
              <GridPager grid={grid} sizes={PAGE_SIZES} onParams={onParams} />
            </>
          )
        }
      />
      <CommentDialog
        title={reading && `Comment on ${reading.itemName}`}
        comment={reading?.comments ?? ''}
        onClose={() => setReading(null)}
      />
    </>
  );
}
