import { useFormatters } from '@/format/useFormatters';
import { SegmentedFilter } from '@/ui/SegmentedFilter/SegmentedFilter';
import { ReportWidget } from '../shared';
import type { ReportTableData } from '../shared';
import type { StatusMix } from './api';
import { NOTHING, OUTCOME_WORDS, RECORD_WORDS } from './words';
import { RECORD_TYPES } from './schemas';
import type { RecordType } from './schemas';
import styles from './ApprovalsReport.module.css';

const BAR_COLOURS: Record<string, string> = {
  waiting: 'var(--chart-planned)',
  approved: 'var(--chart-actual)',
  declined: 'var(--chart-budget)',
  recalled: 'var(--chart-committed)',
  replaced: 'var(--chart-committed)',
  autoApproved: 'var(--chart-actual)',
};

const TIP =
  'Each round of the selected record type counted once by how it ended. Waiting is open now. Replaced means ' +
  'a new submission took its place. Shares are whole numbers that add to 100.';

export interface StatusWidgetProps {
  data?: StatusMix;
  statusType: RecordType;
  onStatusType: (type: RecordType) => void;
  failed?: boolean;
  onRetry?: () => void;
}

function tableOf(mix: StatusMix): ReportTableData {
  return {
    columns: [
      { label: 'Status' },
      { label: 'Count', numeric: true },
      { label: 'Share', numeric: true },
    ],
    rows: mix.items.map((item) => ({
      key: item.outcome,
      cells: [
        OUTCOME_WORDS[item.outcome] ?? item.outcome,
        String(item.count),
        `${item.sharePercent} %`,
      ],
    })),
    total: ['Total', String(mix.total), mix.total > 0 ? '100 %' : NOTHING],
  };
}

function Bars({ mix }: { mix: StatusMix }) {
  const { dateTime } = useFormatters();
  return (
    <>
      <p className={styles.waiting}>
        {mix.waiting.count} waiting
        {mix.waiting.oldestSince && ` · oldest since ${dateTime(mix.waiting.oldestSince)}`}
      </p>
      <ul className={styles.bars}>
        {mix.items.map((item) => (
          <li key={item.outcome} className={styles.bar}>
            <span>{OUTCOME_WORDS[item.outcome] ?? item.outcome}</span>
            <span className={styles.track} aria-hidden="true">
              <span
                className={styles.fill}
                style={{ width: `${item.sharePercent}%`, background: BAR_COLOURS[item.outcome] }}
              />
            </span>
            <span className={styles.value}>
              {item.count} · {item.sharePercent} %
            </span>
          </li>
        ))}
      </ul>
    </>
  );
}

export function StatusWidget({
  data,
  statusType,
  onStatusType,
  failed,
  onRetry,
}: StatusWidgetProps) {
  const empty: ReportTableData = { columns: [{ label: 'Status' }], rows: [] };
  return (
    <ReportWidget
      report="approvals"
      widget="status"
      title="Current approval status"
      tip={TIP}
      loading={!data && !failed}
      failed={failed}
      onRetry={onRetry}
      table={data ? tableOf(data) : empty}
      chart={
        data && (
          <>
            {data.recordTypes.length > 1 && (
              <div className={styles.selector}>
                <SegmentedFilter
                  label="Record type"
                  options={RECORD_TYPES.filter((type) => data.recordTypes.includes(type)).map(
                    (type) => ({
                      value: type,
                      label: RECORD_WORDS[type],
                    }),
                  )}
                  value={statusType}
                  onChange={onStatusType}
                />
              </div>
            )}
            {data.recordTypes.length === 0 ? (
              <p className={styles.empty}>
                No approvals yet. They start when an activity or a purchase request is submitted for
                approval.
              </p>
            ) : (
              <Bars mix={data} />
            )}
          </>
        )
      }
    />
  );
}
