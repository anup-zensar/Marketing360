import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { approvalsReport, noApprovalsReport } from '@/testing/fixtures/approvals-report';
import { testUser } from '@/testing/fixtures/users';
import { expectNoAxeViolations } from '@/testing/axe';
import { renderWithProviders } from '@/testing/render';
import type { StatusWidgetProps } from './StatusWidget';
import { StatusWidget } from './StatusWidget';

function renderWidget(props: Partial<StatusWidgetProps> = {}, route = '/reports/approvals') {
  return renderWithProviders(
    <StatusWidget
      data={approvalsReport().statusMix}
      statusType="activity"
      onStatusType={() => undefined}
      {...props}
    />,
    { route, user: testUser(), path: '/reports/approvals' },
  );
}

describe('StatusWidget (041 FR-006 – FR-008)', () => {
  it('shows each outcome with its count and whole-number share, and who has waited longest', async () => {
    renderWidget();

    expect(screen.getByRole('heading', { name: 'Current approval status' })).toBeInTheDocument();
    expect(screen.getByText(/1 waiting · oldest since/)).toBeInTheDocument();
    expect(screen.getByText('Replaced by a new submission')).toBeInTheDocument();
    expect(screen.getByText('Approved automatically')).toBeInTheDocument();
    expect(screen.getAllByText('2 · 25 %')).toHaveLength(2);
    await expectNoAxeViolations();
  });

  it('offers only the record types that have rounds, and tells the page which was chosen', async () => {
    const onStatusType = vi.fn();
    const { user } = renderWidget({ onStatusType });

    expect(screen.queryByRole('button', { name: 'Modified purchase requests' })).toBeNull();
    await user.click(screen.getByRole('button', { name: 'Purchase requests' }));

    expect(onStatusType).toHaveBeenCalledWith('purchaseRequest');
  });

  it('lists the same figures in the table view, with the total', () => {
    renderWidget({}, '/reports/approvals?table=status');

    expect(screen.getAllByRole('row').map((row) => row.textContent)).toEqual([
      'StatusCountShare',
      'Waiting113 %',
      'Approved225 %',
      'Declined113 %',
      'Recalled225 %',
      'Replaced by a new submission112 %',
      'Approved automatically112 %',
      'Total8100 %',
    ]);
  });

  it('says there are no approvals yet, and what creates them', () => {
    renderWidget({ data: noApprovalsReport().statusMix });

    expect(screen.getByText(/No approvals yet\. They start when/)).toBeInTheDocument();
    expect(screen.queryByRole('group', { name: 'Record type' })).toBeNull();
  });

  it('shows a loading skeleton', () => {
    renderWidget({ data: undefined });

    expect(screen.getByText('Loading…')).toBeInTheDocument();
  });

  it('shows an error with Try again', async () => {
    const onRetry = vi.fn();
    const { user } = renderWidget({ data: undefined, failed: true, onRetry });
    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(onRetry).toHaveBeenCalledOnce();
  });
});
