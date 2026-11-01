import { screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { nth } from '@/testing/dom';
import { renderWithProviders } from '@/testing/render';
import type { ReportAttributeGroup } from './api';
import { LevelTabs } from './LevelTabs';

const GROUPS: ReportAttributeGroup[] = [
  {
    level: 'activity',
    attributeId: 'SA-B',
    label: 'Brand',
    values: [
      { id: 'SO-HX', label: 'Helix', active: true },
      { id: 'SO-PZ', label: 'Pennzoil', active: false },
      { id: 'none', label: 'Not filled in', active: true },
    ],
  },
  {
    level: 'activity',
    attributeId: 'SA-CH',
    label: 'Channel',
    values: [{ id: 'SO-EX', label: 'External', active: true }],
  },
  {
    level: 'campaign',
    attributeId: 'SA-CL',
    label: 'Cluster',
    values: [{ id: 'SO-EU', label: 'Europe', active: true }],
  },
];

function renderTabs() {
  const onChange = vi.fn();
  return {
    ...renderWithProviders(<LevelTabs groups={GROUPS} ticked={{}} onChange={onChange} />),
    onChange,
  };
}

/** The tab's own Select all or Clear all, above its groups' buttons of the same name. */
const tabButton = (name: string) =>
  nth(within(screen.getByRole('tabpanel')).getAllByRole('button', { name }), 0);

describe('LevelTabs (039 FR-008, FR-010)', () => {
  it('offers one tab per level in order, opens on Activity and marks an inactive value', async () => {
    renderTabs();

    expect(screen.getAllByRole('tab').map((tab) => tab.textContent)).toEqual([
      'Plan',
      'Campaign',
      'Initiative',
      'Activity',
      'Line item',
    ]);
    expect(screen.getByRole('tab', { name: 'Activity' })).toHaveAttribute('aria-selected', 'true');
    expect(screen.getByRole('checkbox', { name: 'Pennzoil (inactive)' })).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('selects all or clears all of the open tab only, leaving other levels alone (FR-010)', async () => {
    const { user, onChange } = renderTabs();

    await user.click(tabButton('Clear all'));
    expect(onChange).toHaveBeenLastCalledWith({ 'attr:SA-B': [], 'attr:SA-CH': [] });
    await user.click(tabButton('Select all'));
    expect(onChange).toHaveBeenLastCalledWith({ 'attr:SA-B': null, 'attr:SA-CH': null });
  });

  it('says when a level has no attributes to filter by', async () => {
    const { user } = renderTabs();

    await user.click(screen.getByRole('tab', { name: 'Plan' }));

    expect(screen.getByText('No plan attributes to filter by.')).toBeInTheDocument();
  });
});
