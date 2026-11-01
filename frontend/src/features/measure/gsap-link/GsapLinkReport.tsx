import { ErrorState } from '@/ui/ErrorState/ErrorState';
import { Skeleton } from '@/ui/Skeleton/Skeleton';
import { Tabs } from '@/ui/Tabs/Tabs';
import { ReportFrame } from '../shared';
import type { ReportAddress } from '../shared';
import { useGsapLink } from './api';
import type { GsapLinkReport as Report } from './api';
import { GSAP_TABS, useGsapTable } from './gsap-address';
import type { GsapTable } from './gsap-address';
import { TAB_LABELS } from './gsap-text';
import { SwbsTab } from './SwbsTab';
import { WaitingTab } from './WaitingTab';

/** GSAP link (feature 044): SWBS not received, Waiting for GSAP and Purchase orders, in 039's frame. */
export function GsapLinkReport() {
  return (
    <ReportFrame report="gsapLink">
      {(_, address) => <GsapLinkBody address={address} />}
    </ReportFrame>
  );
}

function GsapLinkBody({ address }: { address: ReportAddress }) {
  const { table, change } = useGsapTable();
  const report = useGsapLink(address, table);
  return (
    <Tabs
      label="GSAP link"
      tabs={GSAP_TABS.map((value) => ({ value, label: TAB_LABELS[value] }))}
      value={table.tab}
      onChange={(tab) => change({ tab })}
    >
      {report.data ? (
        <TabBody report={report.data} table={table} onChange={change} />
      ) : report.isError ? (
        <ErrorState error={report.error} onRetry={() => void report.refetch()} />
      ) : (
        <div aria-label="Loading the tab">
          <Skeleton lines={6} height="24px" />
        </div>
      )}
    </Tabs>
  );
}

interface TabBodyProps {
  report: Report;
  table: GsapTable;
  onChange: (patch: Partial<GsapTable>) => void;
}

function TabBody({ report, table, onChange }: TabBodyProps) {
  if (report.swbs) {
    return <SwbsTab tab={report.swbs} total={report.total} table={table} onChange={onChange} />;
  }
  if (report.waiting) {
    return (
      <WaitingTab
        tab={report.waiting}
        finance={report.finance}
        total={report.total}
        table={table}
        onChange={onChange}
      />
    );
  }
  return null;
}
