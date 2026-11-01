import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { renderWithProviders } from '@/testing/render';
import { DEFAULT_ADDRESS } from '../shared/report-address';
import type { PlanDetailsWidget } from './api';
import { value, widget } from './plan-details-fixture';
import { ValueWidget } from './ValueWidget';

function renderWidget(w: PlanDetailsWidget, route = '/reports/plan-details') {
  return renderWithProviders(
    <ValueWidget
      widget={w}
      level="activity"
      address={{ ...DEFAULT_ADDRESS, year: 2026 }}
      attributeId="SA-12"
      currency="EUR"
    />,
    { route, user: testUser(), path: '/reports/plan-details' },
  );
}

describe('ValueWidget (043 FR-009 – FR-017)', () => {
  it('draws every value with its figure; a value opens All work filtered by it', async () => {
    renderWidget(widget());

    expect(screen.getByRole('link', { name: 'Not filled in' })).toHaveAttribute(
      'href',
      '/all-work?year=2026&attr=SA-12%3Anone',
    );
    expect(screen.getByText('12')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('gathers the values after the ten largest into Other on the chart, never in the table (FR-017)', async () => {
    const values = Array.from({ length: 12 }, (_, i) =>
      value(`O${String(i).padStart(2, '0')}`, 20 - i),
    );
    const { user } = renderWidget(widget({ values, other: value('other', 17, undefined, false) }));

    expect(screen.getByText('Other')).toBeInTheDocument();
    expect(screen.queryByText('O11')).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Table' }));
    expect(await screen.findByRole('cell', { name: 'O11' })).toBeInTheDocument();
  });

  it('says how many records it does not apply to, and the filled-in per cent (043-C2)', () => {
    renderWidget(
      widget({
        key: 'attr12FilledIn',
        title: 'Product launch filled in',
        measure: 'filledIn',
        values: [value('filled', 15, undefined, false), value('none', 5), value('doesNotApply', 8)],
        filledInPercent: 75,
      }),
    );

    expect(
      screen.getByText('75 % filled in · 8 activities where it does not apply are not counted'),
    ).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Does not apply' })).not.toBeInTheDocument();
  });

  it('draws Does not apply as a bar on a value widget, with no note about it', () => {
    renderWidget(widget({ values: [value('EU', 12), value('doesNotApply', 8, undefined, false)] }));

    expect(screen.getByText('Does not apply')).toBeInTheDocument();
    expect(screen.queryByText(/where it does not apply are not counted/)).not.toBeInTheDocument();
  });

  it('shows a value that opens All work as a link', () => {
    renderWidget(widget());

    expect(screen.getByRole('link', { name: 'Not filled in' })).toHaveClass(/barLabel/);
  });

  it('shows planned cost in the report currency, and says when records count in each value', () => {
    renderWidget(
      widget({
        key: 'attr12PlannedCost',
        title: 'Planned cost by Channel',
        measure: 'plannedCost',
        countsInEach: true,
        values: [value('EXT', 2, '1500.00'), value('INT', 1, '500.00')],
        total: value('total', 2, '1500.00', false),
      }),
    );

    expect(
      screen.getByText(/Activities with several values count in each group/),
    ).toBeInTheDocument();
    expect(screen.getByText(/1,500\.00/)).toBeInTheDocument();
  });
});
