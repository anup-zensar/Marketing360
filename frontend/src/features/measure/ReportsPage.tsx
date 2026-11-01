import { ChartColumnIncreasing } from 'lucide-react';
import { Link } from 'react-router';
import { ApiError } from '@/api/errors';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { NoAccess } from '@/ui/NoAccess/NoAccess';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { StatusChip } from '@/ui/StatusChip/StatusChip';
import { usePageTitle } from '@/ui/usePageTitle';
import { useReportCatalogue } from './shared/api';
import styles from './ReportsPage.module.css';

/** The reports the person may open, in the catalogue's order, each with its one-line purpose (039 FR-001). */
export function ReportsPage() {
  usePageTitle('Reports');
  const catalogue = useReportCatalogue();
  if (catalogue.error instanceof ApiError && catalogue.error.kind === 'forbidden') {
    return <NoAccess />;
  }
  return (
    <>
      <PageHeader
        title="Reports"
        description="Read-only reports of your plans, approvals, purchase requests and their figures."
      />
      {catalogue.data ? (
        <ul className={styles.list}>
          {catalogue.data.items.map((report) => (
            <li key={report.key} className={styles.report}>
              <ChartColumnIncreasing aria-hidden="true" className={styles.icon} />
              <div className={styles.text}>
                {report.available ? (
                  <Link to={`/reports/${report.path}`} className={styles.name}>
                    {report.name}
                  </Link>
                ) : (
                  <span className={styles.name}>{report.name}</span>
                )}
                <p className={styles.purpose}>{report.purpose}</p>
              </div>
              {!report.available && <StatusChip tone="neutral">Coming soon</StatusChip>}
            </li>
          ))}
        </ul>
      ) : catalogue.isError ? (
        <ErrorState error={catalogue.error} onRetry={() => void catalogue.refetch()} />
      ) : (
        <div aria-label="Loading the reports">
          <Skeleton lines={6} height="48px" />
        </div>
      )}
    </>
  );
}
