import { screen, within } from '@testing-library/react';
import type { HttpHandler } from 'msw';
import { describe, expect, it } from 'vitest';
import type { Permission } from '@/session/access';
import { dashboard } from '@/testing/fixtures/dashboards';
import { datasetSummary } from '@/testing/fixtures/datasets';
import { seedList } from '@/testing/fixtures/plans';
import { customDashboard, withPermissions } from '@/testing/fixtures/shared-dashboards';
import { TEAMS } from '@/testing/fixtures/teams';
import { mockApi } from '@/testing/mock-api';
import { renderRoutes } from '@/testing/render';
import { pending, reply } from '@/testing/replies';
import '@/features/custom-dashboards';
import '@/features/dashboards';
import { routes } from './routes';

const ALL: Permission[] = ['plans.view', 'dashboards.view', 'dashboards.build'];

function open(route: string, permissions: Permission[] = ALL, answers: HttpHandler[] = []) {
  mockApi([
    reply('GET /api/v1/me', { body: withPermissions(permissions) }),
    reply('GET /api/v1/custom-dashboards', { body: { mine: [], sharedWithMe: [] } }),
    reply('GET /api/v1/custom-dashboards/{dashboardId}', { body: customDashboard() }),
    reply('GET /api/v1/datasets', { body: { mine: [datasetSummary()], sharedWithMe: [] } }),
    reply('GET /api/v1/plans', { body: seedList() }),
    reply('GET /api/v1/teams', { body: TEAMS }),
    reply('POST /api/v1/data-sources/query', pending),
    ...answers,
  ]);
  return renderRoutes(routes, { route });
}

async function heading(name: string) {
  return screen.findByRole('heading', { level: 1, name }, { timeout: 10_000 });
}

describe('051 FR-002 – FR-004: the dashboard routes', () => {
  it('shows the Dashboards list at /dashboards instead of opening the Executive dashboard', async () => {
    const { router } = open('/dashboards');

    expect(await heading('Dashboards')).toBeInTheDocument();
    expect(router.state.location.pathname).toBe('/dashboards');
    const navigation = screen.getByRole('navigation', { name: 'Main' });
    expect(within(navigation).getByRole('link', { name: 'Dashboards' })).toHaveAttribute(
      'aria-current',
      'page',
    );
    expect(document.title).toBe('Dashboards · SP&A One');
  });

  it.each([
    ['/dashboards/executive', 'Executive dashboard', 'GET /api/v1/dashboards/executive'],
    ['/dashboards/regional', 'Regional dashboard', 'GET /api/v1/dashboards/regional'],
  ] as const)('keeps 036’s %s working', async (route, title, key) => {
    open(route, ALL, [reply(key, { body: dashboard() })]);

    expect(await heading(title)).toBeInTheDocument();
  });

  it('opens a custom dashboard at /dashboards/<id>', async () => {
    open('/dashboards/DB-1001');

    expect(await heading('Lubricants Europe overview')).toBeInTheDocument();
  });

  it('opens Datasets at /datasets for builders', async () => {
    open('/datasets');

    expect(await heading('Datasets')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Dashboard builder' })).toHaveAttribute(
      'aria-current',
      'page',
    );
  });

  it.each([
    ['/builder', ['plans.view', 'dashboards.view']],
    ['/builder/DB-1001', ['plans.view', 'dashboards.view']],
    ['/datasets', ['plans.view', 'dashboards.view']],
    ['/datasets/DS-1001', ['plans.view', 'dashboards.view']],
    ['/dashboards/DB-1001', ['plans.view']],
    ['/dashboards/executive', ['dashboards.view']],
  ] as const)('shows No access at %s without its permission', async (route, permissions) => {
    open(route, [...permissions]);

    expect(await heading('No access')).toBeInTheDocument();
  });

  it.each([
    ['/dashboards', 'Dashboards'],
    ['/dashboards/DB-1001', 'Dashboard'],
    ['/builder', 'New dashboard'],
    ['/builder/DB-1001', 'Edit dashboard'],
    ['/datasets', 'Datasets'],
    ['/datasets/DS-1001', 'Dataset'],
  ])('names %s in the browser tab while the session is checked', async (route, title) => {
    mockApi([reply('GET /api/v1/me', pending)]);
    renderRoutes(routes, { route });

    await screen.findByRole('status');
    expect(document.title).toBe(`${title} · SP&A One`);
  });
});
