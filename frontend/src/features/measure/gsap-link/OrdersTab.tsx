import { WalletCards } from 'lucide-react';
import { Link } from 'react-router';
import { useFormatters } from '@/format/useFormatters';
import { EmptyState } from '@/ui/EmptyState/EmptyState';
import { purchaseRequestHref, recordHref } from '@/ui/record-links';
import type { GsapLinkReport, GsapOrderRow, GsapOrdersTab } from './api';
import type { GsapTable } from './gsap-address';
import { coverageText, pathText } from './gsap-text';
import { PagedTable } from './PagedTable';
import type { PagedColumn } from './PagedTable';
import styles from './GsapLink.module.css';

/** Amount columns: right-aligned, largest first. */
const amount = { numeric: true, largestFirst: true };

interface OrdersTabProps {
  tab: GsapOrdersTab;
  finance: GsapLinkReport['finance'];
  total: number;
  table: GsapTable;
  onChange: (patch: Partial<GsapTable>) => void;
}

/** Purchase orders (044 US3): one row per line item and PO with its amount, invoiced and paid (FR-011 – FR-013). */
export function OrdersTab({ tab, finance, total, table, onChange }: OrdersTabProps) {
  const { date, money } = useFormatters();
  if (finance?.plans === 0) {
    return (
      <EmptyState
        icon={<WalletCards />}
        title="You can't see financial data"
        text="Purchase orders show for plans where your role includes viewing financial data."
      />
    );
  }
  const columns: PagedColumn<GsapOrderRow>[] = [
    {
      key: 'lineItem',
      label: 'Line item',
      cell: (r) => (
        <>
          <Link to={recordHref('lineItem', r.lineItemId)}>{r.lineItem}</Link>
          <span className={styles.sub}>{r.lineItemId}</span>
        </>
      ),
    },
    { key: 'plan', label: 'Plan › Activity', cell: (r) => pathText(r.path) },
    {
      key: 'purchaseRequest',
      label: 'Purchase request',
      cell: (r) =>
        r.purchaseRequestId ? (
          <Link to={purchaseRequestHref(r.purchaseRequestId)}>{r.purchaseRequestId}</Link>
        ) : (
          '—'
        ),
    },
    { key: 'gsapPr', label: 'GSAP PR number', sortable: false, cell: (r) => r.gsapPrNumber || '—' },
    {
      key: 'poNumber',
      label: 'PO number',
      cell: (r) => (
        <>
          {r.poNumber}
          <span className={styles.sub}>{`Lines ${r.poLines.join(', ')}`}</span>
        </>
      ),
    },
    { key: 'poDate', label: 'PO date', largestFirst: true, cell: (r) => date(r.poDate) },
    { key: 'vendor', label: 'Vendor', cell: (r) => r.vendor || '—' },
    { key: 'poAmount', label: 'PO amount', ...amount, cell: (r) => money(r.poAmount) },
    { key: 'invoiced', label: 'Invoiced', ...amount, cell: (r) => money(r.invoiced) },
    { key: 'paid', label: 'Paid', ...amount, cell: (r) => money(r.paid) },
  ];
  const { poAmount, invoiced, paid } = tab.totals;
  const note = coverageText(finance);
  return (
    <div className={styles.body}>
      {note && <p className={styles.note}>{note}</p>}
      <PagedTable
        title="Purchase orders"
        widget="orders"
        columns={columns}
        rows={tab.rows}
        rowKey={(r) => `${r.lineItemId} ${r.poNumber}`}
        table={table}
        total={total}
        empty="No purchase order is linked to a line item in view."
        totals={['Total', '', '', '', '', '', '', money(poAmount), money(invoiced), money(paid)]}
        onChange={onChange}
      />
    </div>
  );
}
