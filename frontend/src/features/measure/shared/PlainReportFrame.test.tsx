import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { reportFrame } from '@/testing/fixtures/reports';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import { ReportFrame } from './ReportFrame';

function renderPlain(forbidden = false) {
  mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/reports/{report}/views', { body: { items: [], total: 0 } }),
    reply(
      'GET /api/v1/reports/frame',
      forbidden
        ? { status: 403, body: apiError('forbidden', "Changing access isn't part of your role.") }
        : { body: reportFrame({ report: 'userAccess' }) },
    ),
  ]);
  return renderWithProviders(
    <ReportFrame report="userAccess" filters={false}>
      {() => <p>The widgets</p>}
    </ReportFrame>,
    { route: '/reports/user-access', user: testUser(), path: '/reports/user-access' },
  );
}

describe('ReportFrame without the plan filters (045-C1)', () => {
  it('keeps the header, saved views and export, and leaves out the filters, In view and Currency', async () => {
    renderPlain();

    expect(await screen.findByText('The widgets')).toBeInTheDocument();
    expect(screen.getByRole('heading', { level: 1, name: 'User access' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Reports' })).toHaveAttribute('href', '/reports');
    expect(screen.getByRole('button', { name: 'Export' })).toBeInTheDocument();
    expect(screen.getByText(/^Figures as of/)).toBeInTheDocument();
    expect(screen.queryByText(/^In view: /)).not.toBeInTheDocument();
    expect(screen.queryByRole('combobox', { name: 'Currency' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Apply' })).not.toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('shows the standard no-access page when the report is refused', async () => {
    renderPlain(true);

    expect(await screen.findByRole('heading', { name: /access/i })).toBeInTheDocument();
    expect(screen.queryByText('The widgets')).not.toBeInTheDocument();
  });
});
