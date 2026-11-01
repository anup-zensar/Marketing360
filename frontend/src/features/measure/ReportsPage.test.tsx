import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { catalogue } from '@/testing/fixtures/reports';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import { ReportsPage } from './ReportsPage';

describe('ReportsPage (039 FR-001, US1 scenarios 1 and 2)', () => {
  it('lists the reports in order with their purposes; Coming soon has no link', async () => {
    mockApi([reply('GET /api/v1/reports', { body: catalogue() })]);
    renderWithProviders(<ReportsPage />, { user: testUser() });

    const items = await screen.findAllByRole('listitem');
    expect(items.map((item) => item.querySelector('a, span')?.textContent)).toEqual([
      'Plan budget',
      'Approvals',
      'Purchase requests',
      'Plan details',
      'GSAP link',
      'Summary',
      'Revenue',
      'Funnel stages',
    ]);
    expect(screen.getByRole('link', { name: 'Plan details' })).toHaveAttribute(
      'href',
      '/reports/plan-details',
    );
    expect(screen.queryByRole('link', { name: 'Plan budget' })).not.toBeInTheDocument();
    expect(screen.getAllByText('Coming soon')).toHaveLength(7);
    expect(screen.getByText('What Plan budget answers.')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('shows the standard no-access page without "See reports"', async () => {
    mockApi([
      reply('GET /api/v1/me', { body: testUser() }),
      reply('GET /api/v1/reports', {
        status: 403,
        body: apiError('forbidden', "Seeing reports isn't part of your role."),
      }),
    ]);
    renderWithProviders(<ReportsPage />, { user: testUser() });

    expect(await screen.findByRole('heading', { name: /access/i })).toBeInTheDocument();
  });
});
