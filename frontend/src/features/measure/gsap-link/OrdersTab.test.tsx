import { screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { renderWithProviders } from '@/testing/render';
import type { GsapLinkReport } from './api';
import { ordersReport } from './gsap-link-fixture';
import { OrdersTab } from './OrdersTab';

function renderTab(report: GsapLinkReport = ordersReport()) {
  if (!report.orders) {
    throw new Error('The fixture has no Purchase orders tab.');
  }
  const onChange = vi.fn();
  const rendered = renderWithProviders(
    <OrdersTab
      tab={report.orders}
      finance={report.finance}
      total={report.total}
      table={{ tab: 'orders', sort: 'lineItem', page: 1, pageSize: 20 }}
      onChange={onChange}
    />,
    { route: '/reports/gsap-link?tab=orders', user: testUser(), path: '/reports/gsap-link' },
  );
  return { onChange, ...rendered };
}

describe('OrdersTab (044 US3, FR-004, FR-011 – FR-013, FR-016)', () => {
  it('lists each line item and PO with its lines, the request it came from and the figures', async () => {
    renderTab();

    const table = screen.getByRole('table', { name: 'Purchase orders' });
    expect(within(table).getByRole('link', { name: 'Venue hire Berlin' })).toHaveAttribute(
      'href',
      '/line-items/L-1003',
    );
    expect(within(table).getByRole('link', { name: 'PR-1001' })).toHaveAttribute(
      'href',
      '/purchase-requests/PR-1001',
    );
    expect(within(table).getByText('Lines 10, 20')).toBeInTheDocument();
    expect(within(table).getByText('Keynote Speakers Ltd')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('ends with the totals of every row, not only the page (FR-013)', () => {
    renderTab();

    const totals = within(screen.getByRole('table', { name: 'Purchase orders' })).getAllByRole(
      'row',
    );
    expect(totals.at(-1)).toHaveTextContent(/^Total.*30,800\.00.*24,500\.00.*12,000\.00/);
  });

  it('says which plans the amounts cover (FR-004)', () => {
    renderTab();

    expect(screen.getByText('Covers 1 of 2 plans')).toBeInTheDocument();
  });

  it("says when the person can't see financial data in any plan in view (FR-018)", () => {
    renderTab(ordersReport({ finance: { plans: 0, of: 2 }, total: 0 }));

    expect(screen.getByText("You can't see financial data")).toBeInTheDocument();
    expect(screen.queryByRole('table')).not.toBeInTheDocument();
  });

  it('sorts amounts largest first', async () => {
    const { user, onChange } = renderTab();

    await user.click(screen.getByRole('button', { name: 'PO amount' }));
    expect(onChange).toHaveBeenLastCalledWith({ sort: '-poAmount' });
  });
});
