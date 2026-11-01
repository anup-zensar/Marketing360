import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { approvalsReport, noApprovalsReport } from '@/testing/fixtures/approvals-report';
import { testUser } from '@/testing/fixtures/users';
import { expectNoAxeViolations } from '@/testing/axe';
import { renderWithProviders } from '@/testing/render';
import type { AverageWidgetProps } from './AverageWidget';
import { AverageWidget } from './AverageWidget';

function renderWidget(props: Partial<AverageWidgetProps> = {}, route = '/reports/approvals') {
  return renderWithProviders(<AverageWidget data={approvalsReport().average} {...props} />, {
    route,
    user: testUser(),
    path: '/reports/approvals',
  });
}

describe('AverageWidget (041 FR-009 – FR-011)', () => {
  it('shows the average and each approver, slowest first', async () => {
    renderWidget();

    expect(screen.getByText('1.3 days')).toBeInTheDocument();
    expect(screen.getByText('Over 3 decided rounds')).toBeInTheDocument();
    const approvers = screen.getAllByRole('listitem').map((item) => item.textContent);
    expect(approvers).toEqual(['Bob Ruiz1.5 days · 2', 'Erin Poe1.0 days · 1']);
    await expectNoAxeViolations();
  });

  it('lists an All approvers row first in the table view', () => {
    renderWidget({}, '/reports/approvals?table=average');

    expect(screen.getAllByRole('row').map((row) => row.textContent)).toEqual([
      'ApproverDaysDecisions',
      'All approvers1.33',
      'Bob Ruiz1.52',
      'Erin Poe1.01',
    ]);
  });

  it('shows a dash and says there are no decisions yet, never zero', () => {
    renderWidget({ data: noApprovalsReport().average });

    expect(screen.getByText('—')).toBeInTheDocument();
    expect(screen.getByText(/No decisions yet/)).toBeInTheDocument();
    expect(screen.queryByText('0 days')).toBeNull();
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
