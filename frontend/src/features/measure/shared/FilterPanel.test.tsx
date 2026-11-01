import { screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { reportFrame } from '@/testing/fixtures/reports';
import { renderWithProviders } from '@/testing/render';
import { FilterPanel } from './FilterPanel';
import { DEFAULT_ADDRESS } from './report-address';
import type { ReportAddress } from './report-address';

function renderPanel(choices = reportFrame().choices, applied: ReportAddress = DEFAULT_ADDRESS) {
  const onApply = vi.fn();
  const rendered = renderWithProviders(
    <FilterPanel choices={choices} applied={applied} year={2026} onApply={onApply} />,
  );
  return { ...rendered, onApply };
}

describe('FilterPanel (039 FR-008 – FR-013)', () => {
  it("lists the groups in FR-008's order with their counters, my teams first", async () => {
    renderPanel();

    const legends = screen
      .getAllByRole('group')
      .map(
        (g) =>
          within(g).queryByText(/^(Plans|Plan type|Team|Owner|Activity type|Activity status)/)
            ?.textContent,
      );
    expect(legends.filter(Boolean)).toEqual([
      'Plans (2 / 2)',
      'Plan type (3 / 3)',
      'Team (2 / 2)',
      'Owner (2 / 2)',
      'Activity type (1 / 1)',
      'Activity status (6 / 6)',
    ]);
    const team = screen.getByRole('group', { name: /Team/ });
    expect(within(team).getAllByRole('checkbox')[0]).toHaveAccessibleName('Lubricants Europe');
    await expectNoAxeViolations();
  });

  it('changes nothing until Apply, then applies the values ticked', async () => {
    const { user, onApply } = renderPanel();

    await user.click(screen.getByRole('checkbox', { name: 'FY2026 Mobility' }));

    expect(screen.getByRole('status')).toHaveTextContent('Changes not applied yet');
    expect(onApply).not.toHaveBeenCalled();
    await user.click(screen.getByRole('button', { name: 'Apply' }));
    expect(onApply).toHaveBeenCalledWith(
      expect.objectContaining({ plans: ['P-1001'], year: 2026 }),
    );
  });

  it('treats every value and no value ticked alike: no restriction (FR-012)', async () => {
    const { user, onApply } = renderPanel(reportFrame().choices, {
      ...DEFAULT_ADDRESS,
      planTypes: ['strategic'],
    });
    const planType = screen.getByRole('group', { name: /Plan type/ });

    await user.click(within(planType).getByRole('button', { name: 'Clear all' }));

    expect(within(planType).getByText('(0 / 3)')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Apply' }));
    expect(onApply).toHaveBeenCalledWith(expect.objectContaining({ planTypes: [] }));
  });

  it('offers a search above ten values (FR-010)', async () => {
    const owners = Array.from({ length: 11 }, (_, i) => ({
      id: `U-${String(2000 + i)}`,
      name: `Owner ${String(i)}`,
    }));
    const { user } = renderPanel({ ...reportFrame().choices, owners });
    const owner = screen.getByRole('group', { name: /Owner/ });

    await user.type(within(owner).getByRole('searchbox', { name: 'Search Owner' }), 'Owner 1');

    expect(
      within(owner)
        .getAllByRole('checkbox')
        .map((c) => c.getAttribute('aria-label') ?? c.closest('label')?.textContent),
    ).toHaveLength(2);
  });
});
