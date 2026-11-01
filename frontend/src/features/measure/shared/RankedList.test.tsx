import { screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { components } from '@/api/generated/schema';
import { expectNoAxeViolations } from '@/testing/axe';
import { planBudgetRanked } from '@/testing/fixtures/plan-budget';
import { testUser } from '@/testing/fixtures/users';
import { renderWithProviders } from '@/testing/render';
import { RankedList } from './RankedList';
import type { RankedColumn, RankedData } from './RankedList';

type Figures = components['schemas']['PlanBudgetFigures'];

const COLUMNS: RankedColumn<Figures>[] = [
  { key: 'budget', label: 'Budget', value: (f) => f.budget?.amount ?? '—' },
  { key: 'planned', label: 'Planned cost', value: (f) => f.planned?.amount ?? '—' },
];

function renderList(
  data: RankedData<Figures> = planBudgetRanked(),
  route = '/reports/plan-budget',
) {
  return renderWithProviders(
    <RankedList
      report="planBudget"
      data={data}
      error={null}
      onRetry={() => {}}
      columns={COLUMNS}
      groups={[
        { value: 'none', label: 'None' },
        { value: 'owner', label: 'Owner' },
      ]}
      defaultSort="planned"
      flag={(f) => (f.overAllocated ? 'Over-allocated' : undefined)}
    />,
    { route, user: testUser(), path: '/reports/plan-budget' },
  );
}

const search = (router: { state: { location: { search: string } } }) =>
  new URLSearchParams(router.state.location.search);

describe('RankedList (040 US2, US3)', () => {
  it('lists each top entry with its children, marks the sorted column and over-allocation (FR-008, FR-009, FR-012)', async () => {
    renderList();

    const rows = screen.getAllByRole('row').map((row) => row.textContent);
    expect(rows[1]).toMatch(/^Helix launchOpen/);
    expect(rows[2]).toMatch(/^Launch eventsOpen/);
    expect(screen.getByRole('columnheader', { name: 'Planned cost' })).toHaveAttribute(
      'aria-sort',
      'descending',
    );
    const fleet = screen.getByRole('rowheader', { name: /^Rimula fleet/ });
    expect(within(fleet).getByText('Over-allocated')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Open Helix launch in Plan' })).toHaveAttribute(
      'href',
      '/plans/P-1001/campaigns/C-1001',
    );
    await expectNoAxeViolations();
  });

  it('keeps sort, group and page in the address, a change starting again at page 1 (FR-013, FR-014)', async () => {
    const { user, router } = renderList(
      planBudgetRanked({ total: 30 }),
      '/reports/plan-budget?year=2026',
    );

    await user.click(screen.getByRole('button', { name: 'Next' }));
    expect(search(router).get('page')).toBe('2');
    await user.selectOptions(screen.getByRole('combobox', { name: 'Sort by' }), 'budget');
    await user.selectOptions(screen.getByRole('combobox', { name: 'Group by' }), 'owner');

    const params = search(router);
    expect([
      params.get('sort'),
      params.get('group'),
      params.get('page'),
      params.get('year'),
    ]).toEqual(['budget', 'owner', null, '2026']);
  });

  it('opens a child as the new top with a trail, and Back goes up one level (FR-010)', async () => {
    const drilled = planBudgetRanked({
      trail: [
        { id: 'P-1001', level: 'plan', name: 'FY2026 Lubricants' },
        { id: 'C-1001', level: 'campaign', name: 'Helix launch' },
        { id: 'I-1001', level: 'initiative', name: 'Launch events' },
      ],
    });
    const { user, router } = renderList(drilled);

    await user.click(screen.getByRole('button', { name: 'Launch events' }));
    expect(search(router).get('drill')).toBe('I-1001');
    const trail = screen.getByRole('navigation', { name: 'Drill trail' });
    expect(within(trail).getByText('Launch events')).toHaveAttribute('aria-current', 'location');

    await user.click(within(trail).getByRole('button', { name: 'Back' }));
    await waitFor(() => expect(search(router).get('drill')).toBe('C-1001'));
  });

  it('says when the opened record is no longer in view (FR-010)', () => {
    renderList(planBudgetRanked({ drillLost: true }));

    expect(screen.getByText(/no longer in view/)).toBeInTheDocument();
  });
});
