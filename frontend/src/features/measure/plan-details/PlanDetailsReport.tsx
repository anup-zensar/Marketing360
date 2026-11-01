import { useFormatters } from '@/format/useFormatters';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { SummaryCard } from '@/ui/SummaryCard/SummaryCard';
import { ReportFrame } from '../shared';
import type { ReportAddress } from '../shared';
import { usePlanDetails } from './api';
import type { PlanDetailsReport as Report } from './api';
import { noAttributesText, RECORDS } from './plan-details-text';
import { ValueWidget } from './ValueWidget';
import styles from './PlanDetails.module.css';

/** Plan details (feature 043): how records and planned cost spread over each Shell attribute of a level. */
export function PlanDetailsReport() {
  return (
    <ReportFrame report="planDetails" period show="activity">
      {(_frame, address) => <PlanDetails address={address} />}
    </ReportFrame>
  );
}

function PlanDetails({ address }: { address: ReportAddress }) {
  const report = usePlanDetails(address);
  if (report.isError) {
    return <ErrorState error={report.error} onRetry={() => void report.refetch()} />;
  }
  if (!report.data) {
    return (
      <div aria-label="Loading Plan details">
        <Skeleton lines={6} height="24px" />
      </div>
    );
  }
  return <Widgets report={report.data} address={address} />;
}

function Widgets({ report, address }: { report: Report; address: ReportAddress }) {
  const { money } = useFormatters();
  const level = report.show;
  const records = RECORDS[level].many;
  const { cards } = report;
  const percent =
    cards.allFilledInPercent === null ? '' : ` (${String(cards.allFilledInPercent)} %)`;
  return (
    <div className={styles.report}>
      <div className={styles.cards}>
        <SummaryCard label={records} value={cards.records.toLocaleString('en-GB')} />
        <SummaryCard
          label="Planned cost"
          value={money(cards.plannedCost)}
          tip="Their planned cost, as every screen shows it, in the report's currency."
        />
        <SummaryCard
          label="All filled in"
          value={`${cards.allFilledIn.toLocaleString('en-GB')} of ${cards.records.toLocaleString('en-GB')}${percent}`}
          tip="Records whose every needed field is filled in."
        />
      </div>
      {report.builtIn.length > 0 && (
        <section aria-label="Built-in fields" className={styles.widgets}>
          {report.builtIn.map((widget) => (
            <ValueWidget
              key={widget.key}
              widget={widget}
              level={level}
              address={address}
              currency={report.currency}
            />
          ))}
        </section>
      )}
      {report.attributesSetUp ? (
        report.attributes.map((row) => (
          <section key={row.attributeId} aria-label={row.label} className={styles.widgets}>
            {row.widgets.map((widget) => (
              <ValueWidget
                key={widget.key}
                widget={widget}
                level={level}
                address={address}
                attributeId={row.attributeId}
                currency={report.currency}
              />
            ))}
          </section>
        ))
      ) : (
        <EmptyState
          title={noAttributesText(level)}
          text="Shell attributes are set up in Settings › Shell attributes."
        />
      )}
    </div>
  );
}
