import { useFormatters } from '@/format/useFormatters';
import { ReportWidget } from '../shared';
import type { ReportTableData } from '../shared';
import type { FunnelPeriod, FunnelReport } from '../revenue/api';
import { reasonText } from './reasons';
import styles from './Funnel.module.css';

const WIDTH = 720;
const HEIGHT = 240;
const BASELINE = 205;
const PLOT = 180;
const SERIES = ['goal', 'actual', 'projected'] as const;

type Series = (typeof SERIES)[number];

function valueOf(period: FunnelPeriod, series: Series): number {
  return period[series] ?? 0;
}

function height(value: number, top: number): number {
  return top > 0 && value > 0 ? (value / top) * PLOT : 0;
}

/** One funnel stage's goal, actual and projected by period, split at today (047 FR-010, 049-C4). */
export function FunnelChart({ report }: { report: FunnelReport }) {
  const { month } = useFormatters();
  const label = (key: string) => (key.includes('Q') ? key.slice(5) : month(key));
  const whole = (value: number | null) => (value === null ? '—' : value.toLocaleString());
  const table: ReportTableData = {
    columns: [
      { label: 'Period' },
      ...['Goal', 'Actual', 'Projected'].map((column) => ({ label: column, numeric: true })),
    ],
    rows: report.periods.map((p) => ({
      key: p.key,
      cells: [label(p.key), whole(p.goal), whole(p.actual), whole(p.projected)],
    })),
  };
  return (
    <ReportWidget
      report="funnel"
      widget="periods"
      title={report.period === 'quarter' ? 'Volume by quarter' : 'Volume by month'}
      tip="Actual to the end of the last period, projected from today, against the activities' goal."
      chart={<Bars periods={report.periods} label={label} />}
      table={table}
      note={report.reason ? reasonText(report.reason) : undefined}
    />
  );
}

function Bars({ periods, label }: { periods: FunnelPeriod[]; label: (key: string) => string }) {
  const top = periods
    .flatMap((p) => SERIES.map((s) => valueOf(p, s)))
    .reduce((a, b) => Math.max(a, b), 0);
  const group = WIDTH / Math.max(periods.length, 1);
  const bar = (group * 0.7) / SERIES.length;
  const today = periods.find((p) => p.inProgress);
  return (
    <>
      <ul className={styles.legend}>
        {SERIES.map((s) => (
          <li key={s}>
            <span className={`${styles.swatch} ${styles[s]}`} aria-hidden="true" />
            {{ goal: 'Goal', actual: 'Actual', projected: 'Projected' }[s]}
          </li>
        ))}
      </ul>
      <svg
        className={styles.chart}
        viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
        role="img"
        aria-label={`Goal, actual and projected volume by period${today ? `; today is in ${label(today.key)}` : ''}. The Table view lists every period.`}
      >
        {periods.map((p, i) => {
          const x = i * group + group * 0.15;
          return (
            <g key={p.key}>
              {p.inProgress && (
                <rect
                  className={styles.today}
                  x={i * group}
                  y={0}
                  width={group}
                  height={BASELINE}
                />
              )}
              {SERIES.map((s, j) => {
                const h = height(valueOf(p, s), top);
                return (
                  <rect
                    key={s}
                    className={styles[s]}
                    x={x + j * bar}
                    y={BASELINE - h}
                    width={bar}
                    height={h}
                  />
                );
              })}
              <text
                className={styles.axisLabel}
                x={x + group * 0.35}
                y={HEIGHT - 14}
                textAnchor="middle"
              >
                {p.inProgress ? 'Today' : label(p.key).slice(0, 3)}
              </text>
            </g>
          );
        })}
        <line className={styles.axis} x1={0} x2={WIDTH} y1={BASELINE} y2={BASELINE} />
      </svg>
    </>
  );
}
