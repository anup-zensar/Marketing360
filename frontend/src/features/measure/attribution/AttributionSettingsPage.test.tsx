import { render, screen, waitFor } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { components } from '@/api/generated/schema';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { reply } from '@/testing/replies';
import { RoiValue } from '../shared/RoiValue';
import { AttributionSettingsPage } from './AttributionSettingsPage';

type Settings = components['schemas']['AttributionSettings'];

function settings(overrides: Partial<Settings> = {}): Settings {
  return {
    model: 'equallyDistributed',
    models: [
      { key: 'firstTouch', name: 'First touch', description: 'All to the first.' },
      { key: 'lastTouch', name: 'Last touch', description: 'All to the last.' },
      {
        key: 'equallyDistributed',
        name: 'Equally distributed',
        description: 'The same share each.',
      },
      { key: 'linearTimeDecay', name: 'Linear time decay', description: 'Later gets more.' },
      { key: 'uShaped', name: 'U-shaped', description: '40/20/40.' },
    ],
    version: 1,
    updatedAt: null,
    updatedBy: null,
    canChange: { available: true, reason: '' },
    ...overrides,
  };
}

function renderPage(current: Settings = settings()) {
  const api = mockApi([
    reply('GET /api/v1/me', { body: testUser() }),
    reply('GET /api/v1/settings/attribution', { body: current }),
    reply('PUT /api/v1/settings/attribution', { body: settings({ model: 'uShaped', version: 2 }) }),
  ]);
  return { ...renderWithProviders(<AttributionSettingsPage />, { user: testUser() }), api };
}

describe('AttributionSettingsPage (050 US1)', () => {
  it('lists the five models with the designated one chosen, and is accessible (FR-001)', async () => {
    renderPage();

    expect(await screen.findByRole('radio', { name: /Equally distributed/ })).toBeChecked();
    expect(screen.getAllByRole('radio')).toHaveLength(5);
    expect(screen.getByText('Choose another model to save.')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('asks before restating every figure, then saves the new model with its version (FR-002)', async () => {
    const { user, api } = renderPage();

    await user.click(await screen.findByRole('radio', { name: /U-shaped/ }));
    await user.click(screen.getByRole('button', { name: 'Save' }));
    expect(await screen.findByText(/will be restated with U-shaped/)).toBeInTheDocument();
    expect(api.bodiesOf('PUT /api/v1/settings/attribution')).toEqual([]);

    await user.click(screen.getByRole('button', { name: 'Change the model' }));
    await waitFor(() =>
      expect(api.bodiesOf('PUT /api/v1/settings/attribution')).toEqual([
        { version: 1, model: 'uShaped' },
      ]),
    );
  });

  it('shows the model without Save, with the reason, to a person who may not change it (FR-002)', async () => {
    renderPage(
      settings({ canChange: { available: false, reason: 'Changing it needs Manage settings.' } }),
    );

    expect(await screen.findByText('Changing it needs Manage settings.')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Save' })).not.toBeInTheDocument();
    expect(screen.getByRole('radio', { name: /U-shaped/ })).toBeDisabled();
  });
});

describe('RoiValue (050 US3, FR-007, FR-008)', () => {
  it('names the model and the cost basis next to every ROI', () => {
    render(
      <>
        <RoiValue roi={50} model="U-shaped" basis="actual" />
        <RoiValue roi={null} model="U-shaped" basis="actual" />
        <RoiValue roi={-12} model="Last touch" basis="planned" />
      </>,
    );

    expect(screen.getByText('50%')).toBeInTheDocument();
    expect(screen.getByText('U-shaped · on actual cost')).toBeInTheDocument();
    expect(screen.getByText('No actual cost yet · U-shaped')).toBeInTheDocument();
    expect(screen.getByText('Last touch · on planned cost')).toBeInTheDocument();
  });
});
