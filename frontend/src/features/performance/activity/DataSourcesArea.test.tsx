import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { components } from '@/api/generated/schema';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { VIEWER } from '@/testing/fixtures/viewer';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { reply } from '@/testing/replies';
import { DataSourcesArea } from './DataSourcesArea';

const AREA = 'GET /api/v1/activities/{activityId}/performance/sources';

function area(
  overrides: Partial<components['schemas']['DataSourcesArea']> = {},
): components['schemas']['DataSourcesArea'] {
  return {
    externalName: 'Spring fair',
    aliases: [
      {
        id: 'ENA-1',
        activityId: 'A-1042',
        name: 'Spring fair 2025',
        version: 1,
        createdAt: '2026-10-01T09:00:00Z',
        createdBy: { id: 'U-1001', name: 'Alex Morgan' },
        updatedAt: '2026-10-01T09:00:00Z',
        updatedBy: { id: 'U-1001', name: 'Alex Morgan' },
      },
    ],
    mediaCodeCount: 2,
    integrated: true,
    goalEntered: true,
    sources: [
      {
        sourceId: 'PS-1001',
        sourceName: 'Eloqua',
        integrationId: 'elq-77',
        factCount: 14,
        lastFactDate: '2026-02-01',
      },
    ],
    ...overrides,
  };
}

function renderArea(answer: Parameters<typeof reply>[1], user = testUser()) {
  mockApi([reply(AREA, answer, { activityId: 'A-1042' })]);
  return renderWithProviders(<DataSourcesArea activityId="A-1042" />, { user });
}

describe('DataSourcesArea', () => {
  it('046 FR-031 shows the keys, flags and each source’s totals', async () => {
    renderArea({ body: area() });

    expect(await screen.findByText('Spring fair')).toBeInTheDocument();
    expect(screen.getByText('Spring fair 2025')).toBeInTheDocument();
    expect(screen.getByRole('cell', { name: 'elq-77' })).toBeInTheDocument();
    expect(screen.getByRole('cell', { name: '14' })).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('046 FR-031 says why the table is empty before any source delivers', async () => {
    renderArea({ body: area({ sources: [], integrated: false, aliases: [] }) });

    expect(
      await screen.findByText(/No data source delivers for this activity yet/),
    ).toBeInTheDocument();
  });

  it('046 FR-032 shows the error state with a retry when the read fails', async () => {
    renderArea({
      status: 500,
      body: { error: { code: 'internal_error', message: 'x', fields: [] } },
    });

    expect(await screen.findByRole('button', { name: /try again/i })).toBeInTheDocument();
  });

  it('046 FR-043 hides the key actions from a reader without performance.link', async () => {
    renderArea({ body: area() }, VIEWER);

    await screen.findByText('Spring fair');
    expect(
      screen.queryByRole('button', { name: 'Register integration ID' }),
    ).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Remove ID' })).not.toBeInTheDocument();
  });
});
