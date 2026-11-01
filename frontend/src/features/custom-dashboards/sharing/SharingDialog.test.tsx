import { screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { CHINA_REF, EUROPE_REF } from '@/testing/fixtures/shared-dashboards';
import { renderWithProviders } from '@/testing/render';
import { SharingDialog } from './SharingDialog';
import type { SharingDialogProps } from './SharingDialog';

const OLD_REF = { id: 'T-1003', name: 'Old Markets', colour: 'gold' as const, active: false };

function renderDialog(overrides: Partial<SharingDialogProps> = {}) {
  const onSave = vi.fn<(teamIds: string[]) => void>();
  const onClose = vi.fn<() => void>();
  const view = renderWithProviders(
    <SharingDialog
      open
      recordLabel="dashboard"
      recordName="Lubricants Europe overview"
      teams={[EUROPE_REF, CHINA_REF, OLD_REF]}
      selected={['T-1001']}
      saving={false}
      error={null}
      onSave={onSave}
      onClose={onClose}
      {...overrides}
    />,
  );
  return { ...view, onSave, onClose };
}

describe('051 FR-046: SharingDialog', () => {
  it('lists the owner’s teams, ticked where shared, with the note (US4 scenario 1)', async () => {
    renderDialog();

    const dialog = screen.getByRole('dialog', { name: "Share 'Lubricants Europe overview'" });
    expect(within(dialog).getByRole('checkbox', { name: 'Lubricants Europe' })).toBeChecked();
    expect(within(dialog).getByRole('checkbox', { name: 'Mobility China' })).not.toBeChecked();
    expect(within(dialog).getByRole('checkbox', { name: 'Old Markets (inactive)' })).toBeVisible();
    expect(dialog).toHaveTextContent(
      'People in these teams can view this dashboard. Each of them sees only the data they may see.',
    );
    await expectNoAxeViolations(dialog);
  });

  it('saves every ticked team in the listed order', async () => {
    const { user, onSave } = renderDialog();

    await user.click(screen.getByRole('checkbox', { name: 'Mobility China' }));
    await user.click(screen.getByRole('checkbox', { name: 'Lubricants Europe' }));
    await user.click(screen.getByRole('checkbox', { name: 'Old Markets (inactive)' }));
    await user.click(screen.getByRole('button', { name: 'Save' }));

    expect(onSave).toHaveBeenCalledWith(['T-1002', 'T-1003']);
  });

  it('saves with Enter and makes the record private when nothing is ticked', async () => {
    const { user, onSave } = renderDialog();

    await user.click(screen.getByRole('checkbox', { name: 'Lubricants Europe' }));
    await user.keyboard('{Enter}');

    expect(onSave).toHaveBeenCalledWith([]);
  });

  it('closes on Escape without saving', async () => {
    const { user, onSave, onClose } = renderDialog();

    await user.click(screen.getByRole('checkbox', { name: 'Mobility China' }));
    await user.keyboard('{Escape}');

    expect(onClose).toHaveBeenCalled();
    expect(onSave).not.toHaveBeenCalled();
  });

  it('shows the refusal and keeps the choices', async () => {
    renderDialog({ error: 'You can share only with your teams.', recordLabel: 'dataset' });

    expect(screen.getByRole('alert')).toHaveTextContent('You can share only with your teams.');
    expect(screen.getByRole('dialog')).toHaveTextContent('can view this dataset');
  });

  it('says so when the owner belongs to no team', () => {
    renderDialog({ teams: [], selected: [] });

    expect(screen.getByRole('dialog')).toHaveTextContent(
      "You aren't in any team yet, so there is no one to share this dashboard with.",
    );
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
  });

  it('shows nothing while closed', () => {
    renderDialog({ open: false });

    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });
});
