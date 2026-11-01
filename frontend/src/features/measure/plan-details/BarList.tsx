import { Link } from 'react-router';
import { cx } from '@/ui/cx';
import styles from './PlanDetails.module.css';

export interface Bar {
  key: string;
  label: string;
  /** A count or an amount in cents; href is absent when the value opens nothing. */
  size: number;
  text: string;
  href?: string;
  muted?: boolean;
}

/** Horizontal bars with their labels and figures in words, so colour is never the only signal (D-345, FR-021). */
export function BarList({ title, bars }: { title: string; bars: Bar[] }) {
  const largest = Math.max(1, ...bars.map((bar) => Math.abs(bar.size)));
  return (
    <ul className={styles.bars} aria-label={title}>
      {bars.map((bar) => (
        <li key={bar.key} className={styles.bar}>
          {bar.href ? (
            <Link to={bar.href} className={styles.barLabel}>
              {bar.label}
            </Link>
          ) : (
            <span className={styles.barLabel}>{bar.label}</span>
          )}
          <span className={styles.track} aria-hidden="true">
            <span
              className={cx(styles.fill, bar.muted && styles.fillMuted)}
              style={{ width: `${String((Math.abs(bar.size) / largest) * 100)}%` }}
            />
          </span>
          <span className={styles.barValue}>{bar.text}</span>
        </li>
      ))}
    </ul>
  );
}
