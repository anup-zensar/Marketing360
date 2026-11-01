import { ChartColumnIncreasing } from 'lucide-react';
import type { ReactNode } from 'react';
import { ApiError } from '@/api/errors';
import { useFormatters } from '@/format/useFormatters';
import { Button } from '@/ui/Button/Button';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { FilterChips } from '@/ui/FilterChips/FilterChips';
import { NoAccess } from '@/ui/NoAccess/NoAccess';
import { PageHeader } from '@/ui/PageHeader/PageHeader';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { usePageTitle } from '@/ui/usePageTitle';
import { REPORT_ROUTES } from '../reports';
import type { ReportKey } from '../reports';
import { useReportFrame } from './api';
import type { ReportFrame as Frame } from './api';
import { downloadReport } from './api';
import { ExportMenu } from './ExportMenu';
import { FilterPanel } from './FilterPanel';
import { DEFAULT_ADDRESS, isRestricted, resetFilters, writeAddress } from './report-address';
import type { ReportAddress, Show } from './report-address';
import { chipsOf, coverageText, inViewText, leftOutText, withoutChip } from './report-text';
import { CurrencyControl, InViewLine, PeriodControl, ShowControl } from './ReportControls';
import { useReportAddress } from './useReportAddress';
import styles from './ReportFrame.module.css';

export interface ReportFrameProps {
  report: ReportKey;
  /** The controls the report uses (FR-016, FR-017); Currency is always offered (FR-018). */
  period?: boolean;
  show?: Show;
  /** The report's widgets for the frame; absent for a report not built yet. */
  children?: (frame: Frame, address: ReportAddress) => ReactNode;
  comingSoon?: string;
}

/** The address as the frame applied it: the resolved year and currency, left-out values dropped. */
function appliedAddress(address: ReportAddress, frame: Frame): ReportAddress {
  const f = frame.filter;
  return {
    ...address,
    year: f.year,
    plans: f.plans,
    teams: f.teams,
    owners: f.owners,
    activityTypes: f.activityTypes,
  };
}

/** The panel starts a new draft when the applied filters change, not when Period, Show or Currency do. */
function panelKey(applied: ReportAddress): string {
  const { period, show, currency, tables } = DEFAULT_ADDRESS;
  return writeAddress({ ...applied, period, show, currency, tables }).toString();
}

/** What every report page shares (039 FR-006 – FR-019, FR-031; contracts/ui.md). */
export function ReportFrame({ report, period, show, children, comingSoon }: ReportFrameProps) {
  const name = REPORT_ROUTES.find((r) => r.key === report)?.name ?? 'Report';
  usePageTitle(name);
  const [address, setAddress] = useReportAddress();
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
        meta={data && `${data.filter.year} · Figures as of ${dateTime(data.generatedAt)}`}
        actions={
          data && (
            <ExportMenu
              label="Export"
              formats={['xlsx']}
              download={(format) => downloadReport(report, address, format)}
            />
          )
        }
      />
      {data ? (
        <div className={styles.layout} aria-busy={frame.isFetching || undefined}>
          <FilterPanel
            key={panelKey(appliedAddress(address, data))}
            choices={data.choices}
            applied={appliedAddress(address, data)}
            year={data.filter.year}
            onApply={setAddress}
          />
          <section className={styles.content} aria-label={name}>
            <InViewLine text={inViewText(data.inView)} />
            <div className={styles.chipsRow}>
              <FilterChips
                chips={chipsOf(appliedAddress(address, data), data.choices)}
                onRemove={(key) => setAddress(withoutChip(address, key))}
              />
              {isRestricted(address) && (
                <Button
                  type="button"
                  size="sm"
                  variant="tertiary"
                  onClick={() => setAddress(resetFilters(address))}
                >
                  Reset filters
                </Button>
              )}
            </div>
            {data.leftOut > 0 && (
              <p className={styles.note} role="status">
                {leftOutText(data.leftOut)}
              </p>
            )}
            {coverageText(data.finance.plans, data.finance.of) && (
              <p className={styles.note}>{coverageText(data.finance.plans, data.finance.of)}</p>
            )}
            <div className={styles.controls}>
              {period && (
                <PeriodControl
                  value={address.period}
                  onChange={(p) => setAddress({ ...address, period: p })}
                />
              )}
              {show && (
                <ShowControl
                  value={address.show ?? show}
                  onChange={(s) => setAddress({ ...address, show: s })}
                />
              )}
              <CurrencyControl
                currencies={data.choices.currencies}
                value={data.filter.currency}
                onChange={(c) => setAddress({ ...address, currency: c })}
              />
            </div>
            <FrameBody
              data={data}
              address={address}
              onReset={() => setAddress(resetFilters(address))}
              comingSoon={comingSoon}
            >
              {children}
            </FrameBody>
          </section>
        </div>
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

interface FrameBodyProps {
  data: Frame;
  address: ReportAddress;
  onReset: () => void;
  comingSoon?: string;
  children: ReportFrameProps['children'];
}

/** The widgets, or why there are none (FR-031). */
function FrameBody({ data, address, onReset, comingSoon, children }: FrameBodyProps) {
  if (data.choices.plans.length > 0 && data.inView.plans === 0 && isRestricted(address)) {
    return (
      <EmptyState
        icon={<ChartColumnIncreasing />}
        title="Your filters exclude everything"
        text="No record matches every filter."
        action={
          <Button type="button" variant="secondary" onClick={onReset}>
            Reset filters
          </Button>
        }
      />
    );
  }
  if (data.inView.plans === 0) {
    const others = data.choices.years.filter((y) => y !== data.filter.year);
    return (
      <EmptyState
        icon={<ChartColumnIncreasing />}
        title={`No plans you may see for ${data.filter.year}`}
        text={others.length > 0 ? `There are plans for ${others.join(', ')}.` : undefined}
      />
    );
  }
  if (!children) {
    return <EmptyState icon={<ChartColumnIncreasing />} title="Coming soon" text={comingSoon} />;
  }
  return <div className={styles.widgets}>{children(data, address)}</div>;
}
