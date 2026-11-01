import { screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { nth } from '@/testing/dom';
import { planBudgetRanked, planBudgetReport } from '@/testing/fixtures/plan-budget';
import { reportFrame } from '@/testing/fixtures/reports';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import type { MockedApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import type { Answer } from '@/testing/replies';
import { PlanBudgetReport } from './PlanBudgetReport';

function renderPage(
  report: Answer<'GET /api/v1/reports/plan-budget'> = { body: planBudgetReport() },
  route = '/reports/plan-budget?year=2026',
) {
  const api = mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/reports/frame', { body: reportFrame() }),
    reply('GET /api/v1/reports/{report}/views', { body: { items: [], total: 0 } }),
    reply('GET /api/v1/reports/plan-budget', report),
    reply('GET /api/v1/reports/plan-budget/ranked', { body: planBudgetRanked() }),
    reply('GET /api/v1/reports/{report}/export', () => new Response('file')),
  ]);
  const rendered = renderWithProviders(<PlanBudgetReport />, {
    route,
    user: testUser(),
    path: '/reports/plan-budget',
  });
  return { ...rendered, api };
}

const asked = (api: MockedApi, path: string) =>
  api.requests.filter((r) => r.path === path).map((r) => new URLSearchParams(r.search));

describe('PlanBudgetReport (040 US1)', () => {
  it('shows the ten cards, "—" where a figure has none, and is accessible (FR-002)', async () => {
    renderPage({
      body: planBudgetReport({
        cards: {
          ...planBudgetReport().cards,
          leftToPlan: { amount: '-25.00', currency: 'USD' },
          budgetSpent: null,
        },
      }),
    });

    const cards = await screen.findByLabelText('Summary');
    for (const label of [
      'Budget',
      'Planned cost',
      'Committed',
      'Actual',
      'Budget used',
      'Left to plan',
      'Unspent budget',
      'Budget spent',
      'Not yet committed',
      '% committed',
    ]) {
      expect(within(cards).getByText(label)).toBeInTheDocument();
    }
    expect(within(cards).getByText('50%')).toBeInTheDocument();
    expect(within(cards).getByText('—')).toBeInTheDocument();
    expect(within(cards).getByText('Over-allocated')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('shortens a large amount on its card and keeps the full amount at hand (FR-002)', async () => {
    const usd = (amount: string) => ({ amount, currency: 'USD' });
    renderPage({
      body: planBudgetReport({
        cards: { ...planBudgetReport().cards, budget: usd('2000000.00') },
      }),
    });

    const cards = await screen.findByLabelText('Summary');
    expect(
      within(cards).getByRole('button', { name: /^\$2(\.0)?M \(\$2,000,000\.00\)$/ }),
    ).toBeInTheDocument();
  });

  it('charts spend by month with today marked, and lists every period in the table (FR-004, FR-007)', async () => {
    const { user } = renderPage();

    expect(await screen.findByRole('img', { name: /today is in/ })).toBeInTheDocument();
    expect(screen.getByText('Actual posted in other years: $50.00.')).toBeInTheDocument();
    await user.click(nth(screen.getAllByRole('button', { name: 'Table' }), 0));

    const table = screen.getByRole('table', { name: /Spend by month/ });
    expect(within(table).getAllByRole('row')).toHaveLength(14);
    expect(within(table).getByText('Total')).toBeInTheDocument();
  });

  it('names planned cost without dates and in other years under the chart (FR-005)', async () => {
    const usd = (amount: string) => ({ amount, currency: 'USD' });
    const notes = {
      notPhased: null,
      noDates: usd('70.00'),
      otherYears: usd('40.00'),
      actualOtherYears: null,
    };
    renderPage({ body: planBudgetReport({ notes }) });

    expect(
      await screen.findByText(
        'Planned cost without dates: $70.00. Planned cost in other years: $40.00.',
      ),
    ).toBeInTheDocument();
  });

  it('exports the summary cards on their own (FR-016)', async () => {
    URL.createObjectURL = () => 'blob:export';
    URL.revokeObjectURL = () => {};
    const save = vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
    const { user, api } = renderPage();

    await user.click(await screen.findByRole('button', { name: 'Export summary' }));
    await user.click(await screen.findByRole('menuitem', { name: 'CSV (.csv)' }));

    await waitFor(() => expect(asked(api, '/api/v1/reports/plan-budget/export')).toHaveLength(1));
    const sent = nth(asked(api, '/api/v1/reports/plan-budget/export'), 0);
    expect([sent.get('widget'), sent.get('format')]).toEqual(['cards', 'csv']);
    save.mockRestore();
  });

  it('asks for quarters when the period changes (FR-004)', async () => {
    const { user, api } = renderPage();

    await user.click(await screen.findByRole('button', { name: 'Quarters' }));

    await waitFor(() =>
      expect(asked(api, '/api/v1/reports/plan-budget').at(-1)?.get('period')).toBe('quarter'),
    );
  });

  it('fails on its own: the ranked list stays when the figures fail', async () => {
    renderPage({ status: 500, body: apiError('internal_error', 'Something went wrong.') });

    expect(await screen.findByRole('button', { name: 'Try again' })).toBeInTheDocument();
    expect(await screen.findByText('Helix launch')).toBeInTheDocument();
  });
});
