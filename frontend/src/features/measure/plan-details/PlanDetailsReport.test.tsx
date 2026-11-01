import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { reportFrame } from '@/testing/fixtures/reports';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { reply } from '@/testing/replies';
import type { PlanDetailsReport as Report } from './api';
import { planDetails } from './plan-details-fixture';
import { PlanDetailsReport } from './PlanDetailsReport';

const EMPTY = { plans: 0, campaigns: 0, initiatives: 0, activities: 0, lineItems: 0 };

function renderReport(
  report: Report = planDetails(),
  frame = reportFrame({ report: 'planDetails' }),
  route = '/reports/plan-details',
) {
  const api = mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/reports/frame', { body: frame }),
    reply('GET /api/v1/reports/plan-details', { body: report }),
    reply('GET /api/v1/reports/{report}/views', { body: { items: [], total: 0 } }),
  ]);
  const { user } = renderWithProviders(<PlanDetailsReport />, {
    route,
    user: testUser(),
    path: '/reports/plan-details',
  });
  return Object.assign(api, { user });
}

describe('PlanDetailsReport (043 US1, FR-005 – FR-016)', () => {
  it('shows the cards, the built-in section and one row per attribute', async () => {
    const api = renderReport();

    expect(await screen.findByText('Activities by Company code')).toBeInTheDocument();
    expect(screen.getByText('20 of 28 (71 %)')).toBeInTheDocument();
    expect(screen.getByRole('region', { name: 'Built-in fields' })).toBeInTheDocument();
    expect(screen.getByRole('region', { name: 'Company code' })).toBeInTheDocument();
    expect(screen.getByText('79 % filled in')).toBeInTheDocument();
    expect(api.requests.find((r) => r.path === '/api/v1/reports/plan-details')?.search).toContain(
      'period=month',
    );
    await expectNoAxeViolations();
  });

  it('says "No data yet" when no record of the level is in view, and offers no export (FR-016)', async () => {
    const empty = planDetails({
      cards: {
        records: 0,
        plannedCost: { amount: '0.00', currency: 'EUR' },
        allFilledIn: 0,
        allFilledInPercent: null,
      },
      builtIn: [],
    });
    renderReport(empty);

    expect(await screen.findByText('No data yet')).toBeInTheDocument();
    expect(
      screen.getByText("Activities create this report's data; there are none in view."),
    ).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Open the plans' })).toHaveAttribute('href', '/plans');
    expect(screen.queryByText('Activities by Company code')).not.toBeInTheDocument();
  });

  it('says the filters exclude every record of the level, and resets them (FR-016)', async () => {
    const empty = planDetails({
      cards: {
        records: 0,
        plannedCost: { amount: '0.00', currency: 'EUR' },
        allFilledIn: 0,
        allFilledInPercent: null,
      },
      builtIn: [],
    });
    const { user } = renderReport(
      empty,
      undefined,
      '/reports/plan-details?activityStatus=declined',
    );

    expect(await screen.findByText('Your filters exclude everything')).toBeInTheDocument();
    expect(screen.queryByText('No data yet')).not.toBeInTheDocument();
    const reset = screen.getAllByRole('button', { name: 'Reset filters' }).at(-1);
    if (!reset) {
      throw new Error('No Reset filters button.');
    }
    await user.click(reset);
    expect(await screen.findByText('No data yet')).toBeInTheDocument();
  });

  it('says when no attribute of the level is set up (FR-016)', async () => {
    renderReport(planDetails({ attributes: [], attributesSetUp: false }));

    expect(
      await screen.findByText('No activity attributes are set up in Settings.'),
    ).toBeInTheDocument();
  });

  it("shows the frame's empty states: no plans for the year, or filters that exclude everything", async () => {
    renderReport(
      planDetails(),
      reportFrame({ report: 'planDetails', inView: EMPTY }),
      '/reports/plan-details?planType=crossMarket',
    );

    expect(await screen.findByText('Your filters exclude everything')).toBeInTheDocument();
    expect(screen.queryByText('Activities by Company code')).not.toBeInTheDocument();
  });

  it('opens on the level in the address (FR-001)', async () => {
    renderReport(
      planDetails({ show: 'campaign', builtIn: [], attributes: [], attributesSetUp: false }),
      undefined,
      '/reports/plan-details?show=campaign',
    );

    expect(
      await screen.findByText('No campaign attributes are set up in Settings.'),
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Campaign' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
  });
});
