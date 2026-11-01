import { formatWhole } from '@/format/decimal';
import { useFormatters } from '@/format/useFormatters';
import { cx } from '@/ui/cx';
import { ReportWidget } from '../shared';
import type { UserAccessBand } from './api';
import styles from './UserAccess.module.css';

interface SignInBandsProps {
  bands: UserAccessBand[];
  people: number;
  recordedSince: string;
}

/** The Last sign-in widget: the bands as bars with their counts in words, and its table view (FR-007, SC-006). */
export function SignInBands({ bands, people, recordedSince }: SignInBandsProps) {
  const { date } = useFormatters();
  const largest = Math.max(1, ...bands.map((b) => b.count));
  return (
    <ReportWidget
      report="userAccess"
      widget="signInBands"
      title="Last sign-in"
      tip="The people of the table by how long ago they last signed in. Never: not since sign-ins were first recorded."
      note={`Sign-ins recorded since ${date(recordedSince)}`}
      chart={
        <ul className={styles.bars} aria-label="People by last sign-in">
          {bands.map((b) => (
            <li key={b.key} className={styles.bar}>
              <span>{b.label}</span>
              <span className={styles.track} aria-hidden="true">
                <span
                  className={cx(styles.fill, b.key === 'never' && styles.fillNever)}
                  style={{ width: `${String((b.count / largest) * 100)}%` }}
                />
              </span>
              <span>{formatWhole(b.count)}</span>
            </li>
          ))}
        </ul>
      }
      table={{
        columns: [{ label: 'Last sign-in' }, { label: 'People', numeric: true }],
        rows: bands.map((b) => ({ key: b.key, cells: [b.label, formatWhole(b.count)] })),
        total: ['Total', formatWhole(people)],
      }}
    />
  );
}
