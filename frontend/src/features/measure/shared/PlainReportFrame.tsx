import { useSearchParams } from 'react-router';
import { ApiError } from '@/api/errors';
import { useFormatters } from '@/format/useFormatters';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { NoAccess } from '@/ui/NoAccess/NoAccess';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { usePageTitle } from '@/ui/usePageTitle';
import { REPORT_ROUTES } from '../reports';
import { downloadReport, useReportFrame } from './api';
import { ExportMenu } from './ExportMenu';
import type { ReportFrameProps } from './ReportFrame';
import { SavedViewsMenu } from './SavedViewsMenu';
import { useReportAddress } from './useReportAddress';
import styles from './ReportFrame.module.css';

/**
 * The report page without the plan filters (045-C1): the header with saved views and export, then the widgets; no
 * filter panel, In view, chips, Period, Show or Currency.
 */
export function PlainReportFrame({ report, children }: ReportFrameProps) {
  const name = REPORT_ROUTES.find((r) => r.key === report)?.name ?? 'Report';
  usePageTitle(name);
  const [address] = useReportAddress();
  const [params] = useSearchParams();
  const frame = useReportFrame(report, address);
  const { dateTime } = useFormatters();

  if (frame.error instanceof ApiError && frame.error.kind === 'forbidden') {
    return <NoAccess />;
  }
  const data = frame.data;
  return (
    <>
      <PageHeader
        breadcrumb={[{ label: 'Reports', to: '/reports' }, { label: name }]}
        title={name}
        meta={data && `Figures as of ${dateTime(data.generatedAt)}`}
        actions={
          data && (
            <>
              <SavedViewsMenu report={report} />
              <ExportMenu
                label="Export"
                formats={['xlsx']}
                download={(format) => downloadReport(report, address, format, { own: params })}
              />
            </>
          )
        }
      />
      {data ? (
        <section className={styles.content} aria-label={name}>
          {children && <div className={styles.widgets}>{children(data, address)}</div>}
        </section>
      ) : frame.isError ? (
        <ErrorState error={frame.error} onRetry={() => void frame.refetch()} />
      ) : (
        <div aria-label="Loading the report">
          <Skeleton lines={8} height="24px" />
        </div>
      )}
    </>
  );
}
