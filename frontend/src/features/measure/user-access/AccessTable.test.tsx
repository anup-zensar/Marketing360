import { screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { testUser } from '@/testing/fixtures/users';
import { renderWithProviders } from '@/testing/render';
import { AccessTable } from './AccessTable';
import { userAccessReport } from './user-access-fixture';

function renderTable(total = 3) {
  const onChange = vi.fn();
  const rendered = renderWithProviders(
    <AccessTable
      rows={userAccessReport().rows}
      total={total}
      table={{ q: '', teams: [], status: 'active', sort: 'name', page: 1, pageSize: 20 }}
      onChange={onChange}
    />,
    { route: '/reports/user-access', user: testUser(), path: '/reports/user-access' },
  );
  return { onChange, ...rendered };
}

describe('AccessTable (045 FR-006)', () => {
  it('turns the sorted column round, and sorts last sign-in latest first', async () => {
    const { user, onChange } = renderTable();

    expect(screen.getByRole('columnheader', { name: 'Name' })).toHaveAttribute(
      'aria-sort',
      'ascending',
    );
    await user.click(screen.getByRole('button', { name: 'Name' }));
    expect(onChange).toHaveBeenLastCalledWith({ sort: '-name' });
    await user.click(screen.getByRole('button', { name: 'Last sign-in' }));
    expect(onChange).toHaveBeenLastCalledWith({ sort: '-lastSignedIn' });
  });

  it('pages 5 to 500 rows', async () => {
    const { user, onChange } = renderTable(45);

    expect(screen.getByText('1–20 of 45')).toBeInTheDocument();
    await user.selectOptions(screen.getByLabelText('Per page'), '50');
    expect(onChange).toHaveBeenLastCalledWith({ pageSize: 50 });
    await user.click(screen.getByRole('button', { name: 'Next' }));
    expect(onChange).toHaveBeenLastCalledWith({ page: 2 });
  });
});
