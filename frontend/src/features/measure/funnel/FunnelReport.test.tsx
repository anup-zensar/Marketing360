import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { nth } from '@/testing/dom';
import { reportFrame } from '@/testing/fixtures/reports';
import { funnelRanked, funnelReport } from '@/testing/fixtures/revenue';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import type { Answer } from '@/testing/replies';
import { FunnelReport } from './FunnelReport';

function renderPage(
  report: Answer<'GET /api/v1/reports/funnel'> = { body: funnelReport() },
  route = '/reports/funnel?year=2026',
) {
  const api = mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/reports/frame', { body: reportFrame() }),
    reply('GET /api/v1/reports/{report}/views', { body: { items: [], total: 0 } }),
    reply('GET /api/v1/settings/activity-categories', { body: { items: [], total: 0 } }),
    reply('GET /api/v1/reports/funnel', report),
    reply('GET /api/v1/reports/funnel/ranked', { body: funnelRanked() }),
  ]);
  const rendered = renderWithProviders(<FunnelReport />, {
    route,
    user: testUser(),
    path: '/reports/funnel',
  });
  return { ...rendered, api };
}

describe('FunnelReport (049 US2)', () => {
  it('opens on the first available stage and shows its cards, accessibly (FR-015)', async () => {
    renderPage();

    expect(await screen.findByDisplayValue('INQ')).toBeInTheDocument();
    const cards = await screen.findByLabelText('Summary');
    for (const label of [
      'INQ goal',
      'INQ actual',
      'INQ actual or projected',
      'Performance',
      'Conversion',
    ]) {
      expect(within(cards).getByText(label)).toBeInTheDocument();
    }
    await expectNoAxeViolations();
  });

  it('states how many plans are uncovered by the selected stage (FR-018, 049-C11)', async () => {
    renderPage({ body: funnelReport({ uncoveredCount: 2 }) });

    expect(
      await screen.findByText(/2 plans use a model that does not allow this stage/),
    ).toBeInTheDocument();
  });

  it('shows "—" for conversion when there is no previous stage, never a zero (FR-017)', async () => {
    renderPage({
      body: funnelReport({
        cards: {
          ...funnelReport().cards,
          conversion: { value: null, reason: 'noPreviousStage' },
        },
      }),
    });

    const cards = await screen.findByLabelText('Summary');
    expect(within(cards).getByText('—')).toBeInTheDocument();
  });

  it('charts goal, actual and projected volume by month with today marked (FR-004, 047 FR-010)', async () => {
    const { user } = renderPage();

    expect(await screen.findByRole('img', { name: /today is in/ })).toBeInTheDocument();
    await user.click(nth(screen.getAllByRole('button', { name: 'Table' }), 0));

    const table = screen.getByRole('table', { name: /Volume by month/ });
    expect(within(table).getAllByRole('row')).toHaveLength(13);
  });

  it('fails on its own: the ranked list stays when the figures fail', async () => {
    renderPage({ status: 500, body: apiError('internal_error', 'Something went wrong.') });

    expect(await screen.findByRole('button', { name: 'Try again' })).toBeInTheDocument();
    expect(await screen.findByText('Helix launch')).toBeInTheDocument();
  });
});
