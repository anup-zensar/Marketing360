import { ReportWidget } from '../shared';
import type { ReportTableData } from '../shared';
import type { Average } from './api';
import { NOTHING } from './words';
import styles from './ApprovalsReport.module.css';

const TIP =
  'Days from submission to the final decision, over rounds a person approved or declined, to one decimal. ' +
  'Automatic approvals, recalled and replaced rounds, and rounds still waiting are not counted. Each approver ' +
  'is measured from submission to their own decision.';

export interface AverageWidgetProps {
  data?: Average;
  failed?: boolean;
  onRetry?: () => void;
}

function tableOf(avg: Average): ReportTableData {
  return {
    columns: [
      { label: 'Approver' },
      { label: 'Days', numeric: true },
      { label: 'Decisions', numeric: true },
    ],
    rows: [
      { key: 'all', cells: ['All approvers', avg.days ?? NOTHING, String(avg.decisions)] },
      ...avg.approvers.map((a) => ({
        key: a.person.id,
        cells: [a.person.name, a.days, String(a.decisions)],
      })),
    ],
  };
}

function Figures({ avg }: { avg: Average }) {
  if (avg.days === null) {
    return (
      <>
        <p className={styles.figure}>{NOTHING}</p>
        <p className={styles.empty}>
          No decisions yet. Decisions appear once someone approves or declines.
        </p>
      </>
    );
  }
  return (
    <>
      <p className={styles.figure}>{avg.days} days</p>
      <p className={styles.waiting}>Over {avg.decisions} decided rounds</p>
      <ul className={styles.approvers} aria-label="Average days per approver, slowest first">
        {avg.approvers.map((a) => (
          <li key={a.person.id}>
            <span>{a.person.name}</span>
            <span className={styles.value}>
              {a.days} days · {a.decisions}
            </span>
          </li>
        ))}
      </ul>
    </>
  );
}

export function AverageWidget({ data, failed, onRetry }: AverageWidgetProps) {
  const empty: ReportTableData = { columns: [{ label: 'Approver' }], rows: [] };
  return (
    <ReportWidget
      report="approvals"
      widget="average"
      title="Average days to a decision"
      tip={TIP}
      loading={!data && !failed}
      failed={failed}
      onRetry={onRetry}
      table={data ? tableOf(data) : empty}
      chart={data && <Figures avg={data} />}
    />
  );
}
