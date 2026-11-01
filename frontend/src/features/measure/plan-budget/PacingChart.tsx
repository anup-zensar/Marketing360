import { toCents } from '@/format/decimal';
import { useFormatters } from '@/format/useFormatters';
import { ReportWidget } from '../shared';
import type { ReportTableData } from '../shared';
import type { PlanBudgetPeriod, PlanBudgetReport } from './api';
import styles from './PlanBudget.module.css';

const WIDTH = 720;
const HEIGHT = 240;
const BASELINE = 205;
const PLOT = 180;
const SERIES = ['budget', 'planned', 'actual'] as const;

type Series = (typeof SERIES)[number];

function cents(period: PlanBudgetPeriod, series: Series): bigint {
  const value = period[series];
  return value ? toCents(value.amount) : 0n;
}

function height(value: bigint, top: bigint): number {
  return top > 0n && value > 0n ? Number((value * BigInt(PLOT)) / top) : 0;
}

/** Budget, planned cost and actual per period; actual stops at today's period, which is marked (040 FR-004). */
export function PacingChart({ report }: { report: PlanBudgetReport }) {
  const { money, month } = useFormatters();
  const label = (key: string) => (key.includes('Q') ? key.slice(5) : month(key));
  const amount = (value: PlanBudgetPeriod['actual']) => (value ? money(value) : '—');
  const last = report.periods.at(-1);
  const table: ReportTableData = {
    columns: [
      { label: 'Period' },
      ...[
        'Budget',
        'Planned cost',
        'Actual',
        'Budget to date',
        'Planned to date',
        'Actual to date',
      ].map((column) => ({ label: column, numeric: true })),
    ],
    rows: report.periods.map((p) => ({
      key: p.key,
      cells: [
        p.state === 'inProgress' ? `${label(p.key)} (in progress)` : label(p.key),
        money(p.budget),
        money(p.planned),
        amount(p.actual),
        money(p.budgetToDate),
        money(p.plannedToDate),
        amount(p.actualToDate),
      ],
    })),
    total: last
      ? ['Total', money(last.budgetToDate), money(last.plannedToDate), amount(lastActual(report))]
      : undefined,
  };
  return (
    <ReportWidget
      report="planBudget"
      widget="periods"
      title={report.period === 'quarter' ? 'Spend by quarter' : 'Spend by month'}
      tip="Budget as phased, planned cost as spread and actual by posting date; actual stops at today."
      chart={<Bars periods={report.periods} label={label} />}
      table={table}
      note={notesOf(report, money).join(' ') || undefined}
    />
  );
}

function lastActual(report: PlanBudgetReport) {
  return report.periods.findLast((p) => p.actualToDate !== null)?.actualToDate ?? null;
}

function notesOf(
  report: PlanBudgetReport,
  money: (m: PlanBudgetPeriod['budget']) => string,
): string[] {
  const n = report.notes;
  return [
    n.notPhased && `Budget not phased by month: ${money(n.notPhased)}.`,
    n.noDates && `Planned cost without dates: ${money(n.noDates)}.`,
    n.otherYears && `Planned cost in other years: ${money(n.otherYears)}.`,
    n.actualOtherYears && `Actual posted in other years: ${money(n.actualOtherYears)}.`,
  ].filter((note): note is string => Boolean(note));
}

function Bars({ periods, label }: { periods: PlanBudgetPeriod[]; label: (key: string) => string }) {
  const top = periods
    .flatMap((p) => SERIES.map((s) => cents(p, s)))
    .reduce((a, b) => (a > b ? a : b), 0n);
  const group = WIDTH / Math.max(periods.length, 1);
  const bar = (group * 0.7) / SERIES.length;
  const today = periods.find((p) => p.state === 'inProgress');
  return (
    <>
      <ul className={styles.legend}>
        {SERIES.map((s) => (
          <li key={s}>
            <span className={`${styles.swatch} ${styles[s]}`} aria-hidden="true" />
            {{ budget: 'Budget', planned: 'Planned cost', actual: 'Actual' }[s]}
          </li>
        ))}
      </ul>
      <svg
        className={styles.chart}
        viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
        role="img"
        aria-label={`Budget, planned cost and actual by period${today ? `; today is in ${label(today.key)}` : ''}. The Table view lists every period.`}
      >
        {periods.map((p, i) => {
          const x = i * group + group * 0.15;
          return (
            <g key={p.key}>
              {p === today && (
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
                {p === today ? 'Today' : label(p.key).slice(0, 3)}
              </text>
            </g>
          );
        })}
        <line className={styles.axis} x1={0} x2={WIDTH} y1={BASELINE} y2={BASELINE} />
      </svg>
    </>
  );
}
