import { screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { Permission } from '@/session/access';
import { withPermissions } from '@/testing/fixtures/shared-dashboards';
import { renderWithProviders } from '@/testing/render';
import { Sidebar } from './Sidebar';

function renderSidebar(permissions: Permission[], route = '/') {
  renderWithProviders(<Sidebar />, { user: withPermissions(permissions), route });
}

describe('051 FR-002: Dashboards and Dashboard builder in the sidebar', () => {
  it('shows Dashboard builder right after Dashboards to holders of Build dashboards', () => {
    renderSidebar(['plans.view', 'dashboards.view', 'dashboards.build']);

    const builder = screen.getByRole('link', { name: 'Dashboard builder' });
    expect(builder).toHaveAttribute('href', '/builder');
    const links = screen.getAllByRole('link').map((link) => link.textContent.trim());
    expect(links.indexOf('Dashboard builder')).toBe(links.indexOf('Dashboards') + 1);
  });

  it('hides Dashboard builder without Build dashboards', () => {
    renderSidebar(['plans.view', 'dashboards.view']);

    expect(screen.getByRole('link', { name: 'Dashboards' })).toHaveAttribute('href', '/dashboards');
    expect(screen.queryByRole('link', { name: 'Dashboard builder' })).not.toBeInTheDocument();
  });

  it.each<[string, Permission[]]>([
    ['View plans', ['plans.view']],
    ['See dashboards', ['dashboards.view']],
  ])('shows Dashboards to holders of %s alone', (_name, permissions) => {
    renderSidebar(permissions);

    expect(screen.getByRole('link', { name: 'Dashboards' })).toBeInTheDocument();
  });

  it('hides Dashboards from someone with neither', () => {
    renderSidebar(['finance.view']);

    expect(screen.queryByRole('link', { name: 'Dashboards' })).not.toBeInTheDocument();
  });

  it.each(['/builder', '/builder/DB-1001', '/datasets', '/datasets/DS-1001'])(
    'highlights Dashboard builder on %s',
    (route) => {
      renderSidebar(['plans.view', 'dashboards.view', 'dashboards.build'], route);

      expect(screen.getByRole('link', { name: 'Dashboard builder' })).toHaveAttribute(
        'aria-current',
        'page',
      );
      expect(screen.getByRole('link', { name: 'Dashboards' })).not.toHaveAttribute('aria-current');
    },
  );

  it('highlights Dashboards on a custom dashboard', () => {
    renderSidebar(['plans.view', 'dashboards.view'], '/dashboards/DB-1001');

    expect(screen.getByRole('link', { name: 'Dashboards' })).toHaveAttribute(
      'aria-current',
      'page',
    );
  });
});
