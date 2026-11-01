import { ArrowDown, ArrowRight, ArrowUp } from 'lucide-react';
import { useFormatters } from '@/format/useFormatters';
import { cx } from '@/ui/cx';
import type { CustomDashboardFilters, Widget, WidgetData } from '../model';
import { measureColumn } from './chart-data';
import { formatShare, formatValue } from './format';
import styles from './WidgetView.module.css';

interface KeyFigureProps {
  widget: Widget;
  data: WidgetData;
  filters: CustomDashboardFilters | null;
}

export function KeyFigure({ widget, data, filters }: KeyFigureProps) {
  const { dateFormat } = useFormatters();
  const value = data.totals[0] ?? data.rows[0]?.values[0];
  const column = measureColumn(data, 0);
  const text = value ? formatValue(value, column, widget.display, dateFormat) : '—';
  const wantsPeriod = widget.display.kpi?.comparison === 'previousPeriod';
  const hasPeriod = Boolean(filters?.dateFrom && filters.dateTo);

  return (
    <div className={styles.keyFigure}>
      <p className={styles.figure}>{text}</p>
      <p className={styles.figureLabel}>{column.label}</p>
      <Comparison widget={widget} data={data} />
      {wantsPeriod && !hasPeriod && (
        <p className={styles.figureNote}>
          Set a date range on the dashboard to compare with the previous period.
        </p>
      )}
    </div>
  );
}

function Comparison({ widget, data }: Omit<KeyFigureProps, 'filters'>) {
  const { dateFormat, dateRange } = useFormatters();
  const comparison = data.comparison;
  if (!comparison) {
    return null;
  }
  const compared = formatValue(
    comparison.value,
    measureColumn(data, 0),
    widget.display,
    dateFormat,
  );
  if (comparison.kind === 'target') {
    const reached = comparison.change === null ? '—' : formatShare(comparison.change);
    return <p className={styles.figureNote}>{`${reached} of the target of ${compared}`}</p>;
  }
  const period = comparison.period
    ? ` (${dateRange(comparison.period.from, comparison.period.to)})`
    : '';
  if (comparison.change === null) {
    return (
      <p
        className={styles.figureNote}
      >{`Nothing to compare with in the previous period${period}`}</p>
    );
  }
  const isDown = comparison.change.startsWith('-');
  const isFlat = Number(comparison.change) === 0;
  const higherIsBetter = widget.display.kpi?.higherIsBetter ?? true;
  const isBetter = !isFlat && isDown !== higherIsBetter;
  const Icon = isFlat ? ArrowRight : isDown ? ArrowDown : ArrowUp;
  const size = formatShare(comparison.change.replace(/^-/, ''));
  const words = isFlat
    ? `No change on the previous period${period}`
    : `${isDown ? 'Down' : 'Up'} ${size}, ${isBetter ? 'better' : 'worse'} than the previous period${period}`;
  return (
    <p className={cx(styles.figureNote, styles.change, isBetter && styles.better)}>
      <Icon aria-hidden="true" />
      {words}
    </p>
  );
}
