import { toCents } from '@/format/decimal';
import { useFormatters } from '@/format/useFormatters';
import { ReportWidget } from '../shared';
import type { ReportTableData } from '../shared';
import type { RevenuePeriod, RevenueReport } from './api';
import { reasonText } from './reasons';
import styles from './Revenue.module.css';

const WIDTH = 720;
const HEIGHT = 240;
const BASELINE = 205;
const PLOT = 180;
const SERIES = ['goal', 'actual', 'projected'] as const;

type Series = (typeof SERIES)[number];

function cents(period: RevenuePeriod, series: Series): bigint {
  const value = period[series];
  return value ? toCents(value.amount) : 0n;
}

function height(value: bigint, top: bigint): number {
  return top > 0n && value > 0n ? Number((value * BigInt(PLOT)) / top) : 0;
}

/** Revenue goal, actual and projected by period, split at today (047 FR-010, 049-C4). */
export function RevenueChart({ report }: { report: RevenueReport }) {
  const { money, month } = useFormatters();
  const label = (key: string) => (key.includes('Q') ? key.slice(5) : month(key));
  const amount = (value: RevenuePeriod['actual']) => (value ? money(value) : '—');
  const table: ReportTableData = {
    columns: [
      { label: 'Period' },
      ...['Goal', 'Actual', 'Projected'].map((column) => ({ label: column, numeric: true })),
    ],
    rows: report.periods.map((p) => ({
      key: p.key,
      cells: [label(p.key), amount(p.goal), amount(p.actual), amount(p.projected)],
    })),
  };
  return (
    <ReportWidget
      report="revenue"
      widget="periods"
      title={report.period === 'quarter' ? 'Revenue by quarter' : 'Revenue by month'}
      tip="Actual to the end of the last period, projected from today, against the activities' goal."
      chart={<Bars periods={report.periods} label={label} />}
      table={table}
      note={report.reason ? reasonText(report.reason) : undefined}
    />
  );
}

function Bars({ periods, label }: { periods: RevenuePeriod[]; label: (key: string) => string }) {
  const top = periods
    .flatMap((p) => SERIES.map((s) => cents(p, s)))
    .reduce((a, b) => (a > b ? a : b), 0n);
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
        aria-label={`Revenue goal, actual and projected by period${today ? `; today is in ${label(today.key)}` : ''}. The Table view lists every period.`}
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
                const h = height(cents(p, s), top);
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
