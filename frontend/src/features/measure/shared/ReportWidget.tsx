import type { ReactNode } from 'react';
import { Button } from '@/ui/Button/Button';
import { Card } from '@/ui/Card/Card';
import { InfoTip } from '@/ui/InfoTip/InfoTip';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import type { ReportKey } from '../reports';
import { downloadReport } from './api';
import { ExportMenu } from './ExportMenu';
import { ReportTable } from './ReportTable';
import type { ReportTableData } from './ReportTable';
import { useReportAddress } from './useReportAddress';
import styles from './ReportWidget.module.css';

export interface ReportWidgetProps {
  report: ReportKey;
  /** The widget's key: in the address's table choice and the export's widget (039 contracts). */
  widget: string;
  title: string;
  /** The ⓘ tip naming each figure and its definition (FR-020). */
  tip: string;
  chart: ReactNode;
  table: ReportTableData;
  /** A line under the title: "Covers 1 of 2 plans", "8 activities where it does not apply are not counted". */
  note?: string;
  loading?: boolean;
  failed?: boolean;
  onRetry?: () => void;
}

/** A widget with its tip, Chart · Table switch in the address, export, and its own loading and failure (FR-020 – FR-023). */
export function ReportWidget(props: ReportWidgetProps) {
  const { report, widget, title, tip, chart, table, note, loading, failed, onRetry } = props;
  const [address, setAddress] = useReportAddress();
  const asTable = address.tables.includes(widget);
  const show = (tableView: boolean) =>
    setAddress({
      ...address,
      tables: tableView
        ? [...address.tables, widget]
        : address.tables.filter((key) => key !== widget),
    });
  const actions = (
    <div className={styles.actions}>
      <InfoTip label={title} text={tip} />
      <div className={styles.switch} role="group" aria-label={`Show ${title} as`}>
        {[false, true].map((tableView) => (
          <Button
            key={String(tableView)}
            type="button"
            size="sm"
            variant={asTable === tableView ? 'secondary' : 'tertiary'}
            aria-pressed={asTable === tableView}
            onClick={() => show(tableView)}
          >
            {tableView ? 'Table' : 'Chart'}
          </Button>
        ))}
      </div>
      <ExportMenu
        label="Export"
        formats={['csv', 'xlsx']}
        download={(format) => downloadReport(report, address, format, widget)}
      />
    </div>
  );
  return (
    <Card title={title} actions={actions}>
      {note && <p className={styles.note}>{note}</p>}
      {failed ? (
        <div className={styles.failed} role="alert">
          <p>This part of the report couldn&apos;t be loaded</p>
          <Button type="button" variant="secondary" size="sm" onClick={onRetry}>
            Try again
          </Button>
        </div>
      ) : loading ? (
        <Skeleton lines={5} height="20px" />
      ) : asTable ? (
        <ReportTable title={title} data={table} />
      ) : (
        chart
      )}
    </Card>
  );
}
