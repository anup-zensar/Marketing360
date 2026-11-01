import { useFormatters } from '@/format/useFormatters';
import { cx } from '@/ui/cx';
import type { Widget, WidgetData, WidgetDataColumn } from '../model';
import { formatKey, formatValue } from './format';
import { rowKey } from './describe';
import styles from './WidgetTable.module.css';

interface WidgetTableProps {
  widget: Widget;
  data: WidgetData;
}

/** The same rows as the chart, as a real table (constitution F6). */
export function WidgetTable({ widget, data }: WidgetTableProps) {
  const { dateFormat } = useFormatters();
  const columns = chosenColumns(widget, data);
  const showsTotals = widget.type === 'table' && (widget.display.table?.totals ?? false);
  const caption = widget.title || 'Values';

  return (
    <div className={styles.scroll} role="region" aria-label={`${caption}, table`} tabIndex={0}>
      <table className={styles.table}>
        <caption className="visually-hidden">{caption}</caption>
        <thead>
          <tr>
            {columns.map((column) => (
              <th key={column.key} scope="col" className={cx(isMeasure(column) && styles.number)}>
                {column.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.rows.map((row) => (
            <tr key={rowKey(row)}>
              {columns.map((column, index) => {
                const text = cellText(row, column, { widget, dateFormat });
                return index === 0 && !isMeasure(column) ? (
                  <th key={column.key} scope="row">
                    {text}
                  </th>
                ) : (
                  <td key={column.key} className={cx(isMeasure(column) && styles.number)}>
                    {text}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
        {showsTotals && (
          <tfoot>
            <tr>
              {columns.map((column, index) => (
                <TotalCell
                  key={column.key}
                  column={column}
                  isLabel={index === 0 && !isMeasure(column)}
                  text={totalText(data, column, { widget, dateFormat })}
                />
              ))}
            </tr>
          </tfoot>
        )}
      </table>
    </div>
  );
}

interface TotalCellProps {
  column: WidgetDataColumn;
  isLabel: boolean;
  text: string;
}

function TotalCell({ column, isLabel, text }: TotalCellProps) {
  if (isLabel) {
    return <th scope="row">Total</th>;
  }
  return (
    <td className={cx(isMeasure(column) && styles.number)}>{isMeasure(column) ? text : ''}</td>
  );
}

interface Formatting {
  widget: Widget;
  dateFormat: ReturnType<typeof useFormatters>['dateFormat'];
}

function chosenColumns(widget: Widget, data: WidgetData): WidgetDataColumn[] {
  const chosen = widget.type === 'table' ? (widget.display.table?.columns ?? []) : [];
  if (chosen.length === 0) {
    return data.columns;
  }
  return chosen.flatMap((key) => data.columns.filter((column) => column.key === key));
}

function isMeasure(column: WidgetDataColumn): boolean {
  return /^m\d+$/.test(column.key);
}

function cellText(
  row: WidgetData['rows'][number],
  column: WidgetDataColumn,
  { widget, dateFormat }: Formatting,
): string {
  if (column.key === 'dimension' || column.key === 'series') {
    const key = column.key === 'dimension' ? row.dimension : row.series;
    return key ? formatKey(key, column, dateFormat) : '';
  }
  const value = row.values[Number(column.key.slice(1))];
  return value ? formatValue(value, column, widget.display, dateFormat) : '—';
}

function totalText(data: WidgetData, column: WidgetDataColumn, { widget, dateFormat }: Formatting) {
  const value = isMeasure(column) ? data.totals[Number(column.key.slice(1))] : undefined;
  return value ? formatValue(value, column, widget.display, dateFormat) : '';
}
