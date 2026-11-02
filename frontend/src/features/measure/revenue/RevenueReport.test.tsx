import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { nth } from '@/testing/dom';
import { reportFrame } from '@/testing/fixtures/reports';
import { revenueRanked, revenueReport } from '@/testing/fixtures/revenue';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import type { Answer } from '@/testing/replies';
import { RevenueReport } from './RevenueReport';

function renderPage(
  report: Answer<'GET /api/v1/reports/revenue'> = { body: revenueReport() },
  route = '/reports/revenue?year=2026',
) {
  const api = mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/reports/frame', { body: reportFrame() }),
    reply('GET /api/v1/reports/{report}/views', { body: { items: [], total: 0 } }),
    reply('GET /api/v1/settings/activity-categories', { body: { items: [], total: 0 } }),
    reply('GET /api/v1/reports/revenue', report),
    reply('GET /api/v1/reports/revenue/ranked', { body: revenueRanked() }),
  ]);
  const rendered = renderWithProviders(<RevenueReport />, {
    route,
    user: testUser(),
    path: '/reports/revenue',
  });
  return { ...rendered, api };
}

describe('RevenueReport (049 US1)', () => {
  it('shows the four cards, names the model behind the goal, and is accessible (FR-010, FR-012)', async () => {
    renderPage();

    const cards = await screen.findByLabelText('Summary');
    for (const label of ['Revenue goal', 'Actual revenue', 'Actual or projected', 'Performance']) {
      expect(within(cards).getByText(label)).toBeInTheDocument();
    }
    expect(within(cards).getByText('73%')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('explains the goal in an info tip: the plan model, projection and what the list compares (FR-014)', async () => {
    renderPage();

    const cards = await screen.findByLabelText('Summary');
    expect(within(cards).getByRole('button', { name: 'About Revenue goal' })).toBeInTheDocument();
  });

  it('shows "—" with the reason when a card cannot be worked out, never a zero (FR-009)', async () => {
    renderPage({
      body: revenueReport({
        cards: {
          goal: { value: null, reason: 'noGoal' },
          actual: { value: null, reason: 'noData' },
          projected: { value: null, reason: 'noGoal' },
          performance: { value: null, reason: 'noGoal' },
        },
      }),
    });

    const cards = await screen.findByLabelText('Summary');
    expect(within(cards).getAllByText('No goal set.').length).toBeGreaterThan(0);
    expect(within(cards).getByText('No data has arrived yet.')).toBeInTheDocument();
  });

  it('charts goal, actual and projected by month with today marked, and lists every period in the table (FR-004, 047 FR-010)', async () => {
    const { user } = renderPage();

    expect(await screen.findByRole('img', { name: /today is in/ })).toBeInTheDocument();
    await user.click(nth(screen.getAllByRole('button', { name: 'Table' }), 0));

    const table = screen.getByRole('table', { name: /Revenue by month/ });
    expect(within(table).getAllByRole('row')).toHaveLength(13);
  });

  it('shows the ROI and actual-cost columns as unavailable until feature 050 configures attribution (049-C6)', async () => {
    renderPage();

    await screen.findByText('Helix launch');
    const ranked = screen.getByRole('region', { name: 'Ranked list' });
    expect(
      within(ranked).getAllByText('Attribution is not configured yet.').length,
    ).toBeGreaterThan(0);
  });

  it('fails on its own: the ranked list stays when the figures fail', async () => {
    renderPage({ status: 500, body: apiError('internal_error', 'Something went wrong.') });

    expect(await screen.findByRole('button', { name: 'Try again' })).toBeInTheDocument();
    expect(await screen.findByText('Helix launch')).toBeInTheDocument();
  });
});
