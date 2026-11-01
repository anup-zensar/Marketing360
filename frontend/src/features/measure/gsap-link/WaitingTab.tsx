import { Link } from 'react-router';
import { formatWhole } from '@/format/decimal';
import { useFormatters } from '@/format/useFormatters';
import { purchaseRequestHref } from '@/ui/record-links';
import { SummaryCard } from '@/ui/SummaryCard/SummaryCard';
import type { GsapLinkReport, GsapWaitingRow, GsapWaitingTab } from './api';
import type { GsapTable } from './gsap-address';
import { coverageText, pathText, ROUTING_LABELS, workingDaysText } from './gsap-text';
import { PagedTable } from './PagedTable';
import type { PagedColumn } from './PagedTable';
import styles from './GsapLink.module.css';

interface WaitingTabProps {
  tab: GsapWaitingTab;
  finance: GsapLinkReport['finance'];
  total: number;
  table: GsapTable;
  onChange: (patch: Partial<GsapTable>) => void;
}

/** Waiting for GSAP (044 US2): approved purchase requests GSAP has not answered, by working days since routing. */
export function WaitingTab({ tab, finance, total, table, onChange }: WaitingTabProps) {
  const { dateTime, money } = useFormatters();
  const columns: PagedColumn<GsapWaitingRow>[] = [
    {
      key: 'purchaseRequest',
      label: 'Purchase request',
      cell: (r) => (
        <>
          <Link to={purchaseRequestHref(r.purchaseRequestId)}>{r.name}</Link>
          <span className={styles.sub}>{r.purchaseRequestId}</span>
        </>
      ),
    },
    { key: 'vendor', label: 'Vendor', cell: (r) => r.vendor || '—' },
    { key: 'owner', label: 'Owner', cell: (r) => r.owner },
    { key: 'plan', label: 'Plan › Activity', cell: (r) => pathText(r.path) },
    {
      key: 'routedAt',
      label: 'Routed on',
      largestFirst: true,
      cell: (r) => (
        <>
          {dateTime(r.routedAt)}
          <span className={styles.sub}>{ROUTING_LABELS[r.routing]}</span>
        </>
      ),
    },
    { key: 'gsapPrNumber', label: 'GSAP PR number', cell: (r) => r.gsapPrNumber || '—' },
    {
      key: 'amount',
      label: 'Amount',
      numeric: true,
      largestFirst: true,
      cell: (r) => (r.amount ? money(r.amount) : '—'),
    },
    {
      key: 'waiting',
      label: 'Waiting',
      numeric: true,
      largestFirst: true,
      cell: (r) => (
        <>
          {workingDaysText(r.waitingWorkingDays)}
          {r.overTurnaround && <span className={styles.flag}>Over 2 working days</span>}
        </>
      ),
    },
  ];
  return (
    <div className={styles.body}>
      <div className={styles.cards} aria-label="Summary">
        <SummaryCard
          label="Waiting"
          value={formatWhole(tab.waiting)}
          tip="Approved purchase requests routed to the requisitioner that GSAP has not answered with a matching purchase order."
        />
        <SummaryCard
          label="Amount waiting"
          value={tab.amount ? money(tab.amount) : '—'}
          tip="The purchase requests' amounts, in plans where you may see financial data."
          detail={coverageText(finance)}
        />
        <SummaryCard
          label="Over 2 working days"
          value={formatWhole(tab.overTurnaround)}
          tone={tab.overTurnaround > 0 ? 'problem' : undefined}
          tip="Waiting more than two working days, Monday to Friday, since the routing GSAP has not answered."
        />
      </div>
      <PagedTable
        title="Waiting for GSAP"
        widget="waiting"
        columns={columns}
        rows={tab.rows}
        rowKey={(r) => r.purchaseRequestId}
        table={table}
        total={total}
        empty="No purchase request is waiting for GSAP."
        onChange={onChange}
      />
    </div>
  );
}
