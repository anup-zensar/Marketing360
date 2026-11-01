import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { components } from '@/api/generated/schema';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { reply } from '@/testing/replies';
import { PerformancePanel } from './PerformancePanel';

const PANEL = 'GET /api/v1/activities/{activityId}/performance';

type Panel = components['schemas']['PerformancePanel'];

function panel(overrides: Partial<Panel> = {}): Panel {
  return {
    state: 'ok',
    stages: [
      { stage: 'inq', goal: 100, actual: 80, conversion: null, reason: '' },
      { stage: 'mql', goal: 10, actual: 8, conversion: 10.0, reason: '' },
    ],
    revenue: {
      goal: { amount: '5000.00', currency: 'USD' },
      actual: { amount: '1000.00', currency: 'USD' },
      excludedCount: 1,
      reason: '',
    },
    webMetrics: [{ metric: 'sessions', months: [{ month: '2026-02', value: 1200 }] }],
    sources: [
      { sourceId: 'PS-1001', sourceName: 'Eloqua', freshness: 'stale', dataThrough: '2026-02-01' },
    ],
    ...overrides,
  };
}

function emptyStages(
  reason: 'noSource' | 'noGoal' | 'noData' | 'stale' | 'failed',
): Partial<Panel> {
  return {
    stages: [
      { stage: 'inq', goal: null, actual: null, conversion: null, reason },
      { stage: 'mql', goal: null, actual: null, conversion: null, reason },
    ],
    revenue: { goal: null, actual: null, excludedCount: 0, reason },
    webMetrics: [],
    sources: [],
  };
}

function renderPanel(answer: Panel) {
  mockApi([reply(PANEL, { body: answer }, { activityId: 'A-1042' })]);
  return renderWithProviders(<PerformancePanel activityId="A-1042" />, { user: testUser() });
}

/** SC-008: whatever the state, nothing that looks like a figure appears. */
function expectNoSampleFigure() {
  expect(screen.queryByText(/\d/)).not.toBeInTheDocument();
}

describe('PerformancePanel', () => {
  it('046 FR-029 shows goal, actual, conversion, revenue and freshness when approved', async () => {
    renderPanel(panel());

    expect(await screen.findByRole('row', { name: /INQ 100 80/ })).toBeInTheDocument();
    expect(screen.getByRole('row', { name: /MQL 10 8 10\.0 %/ })).toBeInTheDocument();
    expect(screen.getByText(/1 excluded without a rate/)).toBeInTheDocument();
    expect(screen.getByText('stale')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('046 FR-027 gates a non-approved activity with the stated message and no figures', async () => {
    renderPanel(panel({ state: 'notApproved', stages: [], webMetrics: [], sources: [] }));

    expect(
      await screen.findByText('Approve the activity to see performance data.'),
    ).toBeInTheDocument();
    expectNoSampleFigure();
  });

  it('046 FR-032 states no source and shows no sample figure', async () => {
    renderPanel(panel(emptyStages('noSource')));

    const reasons = await screen.findAllByText('No active data source supplies this figure.');
    expect(reasons.length).toBeGreaterThan(0);
    expectNoSampleFigure();
  });

  it('046 FR-032 states no goal and shows no sample figure', async () => {
    renderPanel(panel(emptyStages('noGoal')));

    expect((await screen.findAllByText('No goal is set for this figure.')).length).toBeGreaterThan(
      0,
    );
    expectNoSampleFigure();
  });

  it('046 FR-032 states no data yet and shows no sample figure', async () => {
    renderPanel(panel(emptyStages('noData')));

    expect((await screen.findAllByText('No data has arrived yet.')).length).toBeGreaterThan(0);
    expectNoSampleFigure();
  });

  it('046 FR-032 states a stale source and shows no sample figure', async () => {
    renderPanel(panel(emptyStages('stale')));

    expect((await screen.findAllByText('The data source is stale.')).length).toBeGreaterThan(0);
    expectNoSampleFigure();
  });

  it('046 FR-032 states a failed load and shows no sample figure', async () => {
    renderPanel(panel(emptyStages('failed')));

    expect((await screen.findAllByText('The last load failed.')).length).toBeGreaterThan(0);
    expectNoSampleFigure();
  });
});
