import { screen, waitFor } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { approvalsReport } from '@/testing/fixtures/approvals-report';
import { reportFrame } from '@/testing/fixtures/reports';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import { ApprovalsReport } from './ApprovalsReport';

type Frame = ReturnType<typeof reportFrame>;

function renderReport(frame: Frame | 'forbidden', route = '/reports/approvals') {
  const frameReply =
    frame === 'forbidden'
      ? { status: 403, body: apiError('forbidden', "Seeing reports isn't part of your role.") }
      : { body: { ...frame, report: 'approvals' as const } };
  mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/reports/{report}/views', { body: { items: [], total: 0 } }),
    reply('GET /api/v1/reports/frame', frameReply),
    reply('GET /api/v1/reports/approvals', { body: approvalsReport() }),
  ]);
  return renderWithProviders(<ApprovalsReport />, {
    route,
    user: testUser(),
    path: '/reports/approvals',
  });
}

describe('ApprovalsReport (041 US1)', () => {
  it('shows both widgets inside the frame, without Period and Show', async () => {
    renderReport(reportFrame());

    expect(
      await screen.findByRole('heading', { name: 'Current approval status' }),
    ).toBeInTheDocument();
    expect(await screen.findByText('1.3 days')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1, name: 'Approvals' })).toBeInTheDocument();
    expect(screen.queryByRole('group', { name: 'Period' })).toBeNull();
    expect(screen.queryByRole('group', { name: 'Show' })).toBeNull();
  });

  it('keeps the chosen record type in the address', async () => {
    const { user, router } = renderReport(reportFrame());

    await user.click(await screen.findByRole('button', { name: 'Purchase requests' }));

    await waitFor(() =>
      expect(new URLSearchParams(router.state.location.search).get('statusType')).toBe(
        'purchaseRequest',
      ),
    );
  });

  it('says there are no plans to see for the year', async () => {
    renderReport(
      reportFrame({
        inView: { plans: 0, campaigns: 0, initiatives: 0, activities: 0, lineItems: 0 },
      }),
    );

    expect(await screen.findByText('No plans you may see for 2026')).toBeInTheDocument();
  });

  it('says the filters exclude everything, with Reset filters', async () => {
    renderReport(
      reportFrame({
        inView: { plans: 0, campaigns: 0, initiatives: 0, activities: 0, lineItems: 0 },
      }),
      '/reports/approvals?planType=strategic',
    );

    expect(await screen.findByText('Your filters exclude everything')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Reset filters' }).length).toBeGreaterThan(0);
  });

  it('shows the standard no-access page without "See reports"', async () => {
    renderReport('forbidden');

    expect(await screen.findByRole('heading', { name: /access/i })).toBeInTheDocument();
  });
});
