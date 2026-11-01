import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { testUser } from '@/testing/fixtures/users';
import { PEOPLE, PRIYA } from '@/testing/fixtures/people';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { reply } from '@/testing/replies';
import { SourceForm } from './SourceForm';

function renderForm() {
  const api = mockApi([
    reply('GET /api/v1/people', { body: PEOPLE }),
    reply('POST /api/v1/performance/sources', {
      status: 201,
      body: {
        id: 'PS-1001',
        name: 'Adobe Analytics',
        description: '',
        ownerId: PRIYA.id,
        category: 'webAnalytics',
        deliveryRoute: 'upload',
        cadence: 'monthly',
        integrationId: '',
        status: 'active',
        freshness: 'stale',
        lastLoadedAt: null,
        dataThrough: null,
        factCount: 0,
        version: 1,
        createdAt: '2026-10-01T09:00:00Z',
        createdBy: { id: 'U-1001', name: 'Alex Morgan' },
        updatedAt: '2026-10-01T09:00:00Z',
        updatedBy: { id: 'U-1001', name: 'Alex Morgan' },
      },
    }),
  ]);
  const onClose = { closed: false };
  const rendered = renderWithProviders(
    <SourceForm
      onClose={() => {
        onClose.closed = true;
      }}
    />,
    { user: testUser() },
  );
  return { api, onClose, ...rendered };
}

describe('SourceForm', () => {
  it('046 FR-001 reports every missing field together, not one at a time', async () => {
    const { user } = renderForm();

    await user.click(await screen.findByRole('button', { name: 'Save' }));

    const nameProblems = await screen.findAllByText(
      'Give the data source a name of 1 to 120 characters.',
    );
    expect(nameProblems.length).toBeGreaterThan(0);
    expect(screen.getAllByText('Choose an owner for this data source.').length).toBeGreaterThan(0);
  });

  it('046 FR-001 creates a source from its fields and closes', async () => {
    const { api, onClose, user } = renderForm();

    await user.type(await screen.findByLabelText(/Name/), 'Adobe Analytics');
    await user.click(screen.getByRole('combobox', { name: /Owner/ }));
    await user.click(await screen.findByRole('option', { name: new RegExp(PRIYA.name) }));
    await user.click(screen.getByRole('button', { name: 'Save' }));

    expect(await screen.findByText(/created/)).toBeInTheDocument();
    const posted = api.requests.find((r) => r.method === 'POST');
    expect(posted).toBeDefined();
    expect(onClose.closed).toBe(true);
  });
});
