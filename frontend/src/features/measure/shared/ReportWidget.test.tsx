import { screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { testUser } from '@/testing/fixtures/users';
import { VIEWER } from '@/testing/fixtures/viewer';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { reply } from '@/testing/replies';
import { ReportWidget } from './ReportWidget';
import type { ReportWidgetProps } from './ReportWidget';

const TABLE = {
  columns: [{ label: 'Value' }, { label: 'Activities', numeric: true }],
  rows: [
    { key: 'GB01', cells: ['GB01', '12'] },
    { key: 'none', cells: ['Not filled in', '6'] },
  ],
  total: ['Total', '18'],
};

function renderWidget(
  props: Partial<ReportWidgetProps> = {},
  route = '/reports/plan-details',
  user = testUser(),
) {
  return renderWithProviders(
    <ReportWidget
      report="planDetails"
      widget="attr12Count"
      title="Activities by Company code"
      tip="Activities per value of Company code."
      chart={<p>The chart</p>}
      table={TABLE}
      {...props}
    />,
    { route, user, path: '/reports/plan-details' },
  );
}

describe('ReportWidget and ReportTable (039 FR-020 – FR-023)', () => {
  afterEach(() => vi.restoreAllMocks());

  it('shows its title, ⓘ tip and chart, and switches to the table kept in the address', async () => {
    const { user, router } = renderWidget();

    expect(screen.getByRole('heading', { name: 'Activities by Company code' })).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'About Activities by Company code' }),
    ).toBeInTheDocument();
    expect(screen.getByText('The chart')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Table' }));

    await waitFor(() =>
      expect(new URLSearchParams(router.state.location.search).getAll('table')).toEqual([
        'attr12Count',
      ]),
    );
    expect(screen.getByRole('table', { name: 'Activities by Company code' })).toBeInTheDocument();
    expect(screen.getAllByRole('row').map((row) => row.textContent)).toEqual([
      'ValueActivities',
      'GB0112',
      'Not filled in6',
      'Total18',
    ]);
    await expectNoAxeViolations();
  });

  it('opens on the table when the address says so', () => {
    renderWidget({}, '/reports/plan-details?table=attr12Count');

    expect(screen.getByRole('button', { name: 'Table' })).toHaveAttribute('aria-pressed', 'true');
    expect(screen.queryByText('The chart')).not.toBeInTheDocument();
  });

  it('offers its export as CSV and Excel', async () => {
    const { user } = renderWidget();

    await user.click(screen.getByRole('button', { name: 'Export' }));

    expect(await screen.findByRole('menuitem', { name: 'CSV (.csv)' })).toBeInTheDocument();
    expect(screen.getByRole('menuitem', { name: 'Excel (.xlsx)' })).toBeInTheDocument();
  });

  it('says why exporting is unavailable without "Export data"', () => {
    const access = VIEWER.access && {
      ...VIEWER.access,
      permissions: VIEWER.access.permissions.filter((p) => p.key !== 'data.export'),
    };
    renderWidget({}, '/reports/plan-details', { ...VIEWER, access });

    const exportButton = screen.getByRole('button', { name: 'Export' });
    expect(exportButton).toBeDisabled();
    expect(exportButton).toHaveAccessibleDescription(/Exporting/);
  });

  it('says why it is empty instead of drawing zeros (FR-031)', () => {
    renderWidget({ empty: 'No data yet: performance rows appear here once an upload is matched.' });

    expect(
      screen.getByText('No data yet: performance rows appear here once an upload is matched.'),
    ).toBeInTheDocument();
    expect(screen.queryByText('The chart')).not.toBeInTheDocument();
  });

  it('loads and fails on its own, with Try again (FR-023)', async () => {
    const onRetry = vi.fn();
    const { user } = renderWidget({ failed: true, onRetry, note: 'Covers 1 of 2 plans' });

    expect(screen.getByRole('alert')).toHaveTextContent(
      "This part of the report couldn't be loaded",
    );
    expect(screen.getByText('Covers 1 of 2 plans')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Try again' }));
    expect(onRetry).toHaveBeenCalled();
  });
  it('exports what the address shows: the report’s own sort, search and grid filters (FR-022)', async () => {
    URL.createObjectURL = () => 'blob:export';
    URL.revokeObjectURL = () => {};
    vi.spyOn(HTMLAnchorElement.prototype, 'click').mockImplementation(() => {});
    const api = mockApi([reply('GET /api/v1/reports/{report}/export', () => new Response('file'))]);
    const { user } = renderWidget(
      {},
      '/reports/plan-details?sort=amount&order=asc&q=fleet&recordType=activity&recordType=purchaseRequest&outcome=approved',
    );

    await user.click(screen.getByRole('button', { name: 'Export' }));
    await user.click(await screen.findByRole('menuitem', { name: 'CSV (.csv)' }));

    await waitFor(() =>
      expect(api.requests.some((r) => r.path === '/api/v1/reports/plan-details/export')).toBe(true),
    );
    const request = api.requests.find((r) => r.path === '/api/v1/reports/plan-details/export');
    const sent = new URLSearchParams(request?.search);
    expect([sent.get('widget'), sent.get('sort'), sent.get('order'), sent.get('q')]).toEqual([
      'attr12Count',
      'amount',
      'asc',
      'fleet',
    ]);
    expect(sent.getAll('recordType')).toEqual(['activity', 'purchaseRequest']);
    expect(sent.getAll('outcome')).toEqual(['approved']);
  });
});
