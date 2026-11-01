import { screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { reportFrame } from '@/testing/fixtures/reports';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, reply } from '@/testing/replies';
import type { GsapLinkReport as Report } from './api';
import { swbsReport } from './gsap-link-fixture';
import { GsapLinkReport } from './GsapLinkReport';

const EMPTY = { plans: 0, campaigns: 0, initiatives: 0, activities: 0, lineItems: 0 };
const PATH = '/api/v1/reports/gsap-link';

type Answer = { body: Report } | { status: number; body: unknown };

function renderReport(
  answer: Answer = { body: swbsReport() },
  frame = reportFrame({ report: 'gsapLink' }),
  route = '/reports/gsap-link',
) {
  const api = mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/reports/frame', { body: frame }),
    reply('GET /api/v1/reports/gsap-link', answer),
    reply('GET /api/v1/reports/{report}/views', { body: { items: [], total: 0 } }),
  ]);
  const rendered = renderWithProviders(<GsapLinkReport />, {
    route,
    user: testUser(),
    path: '/reports/gsap-link',
  });
  return { api, ...rendered };
}

const lastSearch = (api: ReturnType<typeof mockApi>) =>
  api.requests.filter((r) => r.path === PATH).at(-1)?.search ?? '';

describe('GsapLinkReport (044 US1, FR-001, FR-015 – FR-018)', () => {
  it('opens on SWBS not received, longest waiting first, a row opening its line item', async () => {
    const { api } = renderReport();

    expect(await screen.findByRole('link', { name: 'Stand build' })).toHaveAttribute(
      'href',
      '/line-items/L-1006',
    );
    expect(screen.getByRole('tab', { name: 'SWBS not received' })).toHaveAttribute(
      'aria-selected',
      'true',
    );
    expect(lastSearch(api)).toContain('tab=swbs');
    expect(lastSearch(api)).toContain('sort=-waiting');
    await expectNoAxeViolations();
  });

  it('keeps the tab in the address and starts it on its own sort (FR-001)', async () => {
    const { router, user } = renderReport();
    await screen.findByRole('link', { name: 'Stand build' });

    await user.click(screen.getByRole('tab', { name: 'Purchase orders' }));

    expect(router.state.location.search).toBe('?tab=orders');
  });

  it('keeps the sort and the page size in the address, back to page 1 (FR-015)', async () => {
    const { router, user, api } = renderReport(
      { body: swbsReport() },
      reportFrame({ report: 'gsapLink' }),
      '/reports/gsap-link?page=2',
    );
    const table = await screen.findByRole('table', { name: 'SWBS not received' });

    await user.click(within(table).getByRole('button', { name: 'Line item' }));
    expect(router.state.location.search).toBe('?sort=lineItem');

    await user.selectOptions(screen.getByLabelText('Per page'), '50');
    expect(router.state.location.search).toBe('?sort=lineItem&pageSize=50');
    expect(lastSearch(api)).toContain('pageSize=50');
  });

  it('says when no line item waits, and keeps the export', async () => {
    renderReport({ body: swbsReport({ total: 0 }) });

    expect(
      await screen.findByText('No line item is waiting for an SWBS code.'),
    ).toBeInTheDocument();
    expect(screen.queryByRole('table', { name: 'SWBS not received' })).not.toBeInTheDocument();
  });

  it("shows the frame's empty state when the filters exclude everything (FR-018)", async () => {
    renderReport(
      { body: swbsReport() },
      reportFrame({ report: 'gsapLink', inView: EMPTY }),
      '/reports/gsap-link?planType=crossMarket',
    );

    expect(await screen.findByText('Your filters exclude everything')).toBeInTheDocument();
    expect(screen.queryByRole('tab', { name: 'SWBS not received' })).not.toBeInTheDocument();
  });

  it('offers Try again when the tab cannot be read', async () => {
    renderReport({ status: 500, body: apiError('internal_error', 'Something went wrong.') });

    expect(await screen.findByRole('button', { name: 'Try again' })).toBeInTheDocument();
  });
});
