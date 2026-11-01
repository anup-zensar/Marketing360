import { screen, waitFor, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { approvalsGrid, LONG_COMMENT } from '@/testing/fixtures/approvals-report';
import { testUser } from '@/testing/fixtures/users';
import { expectNoAxeViolations } from '@/testing/axe';
import { renderWithProviders } from '@/testing/render';
import { ApprovalsGrid } from './ApprovalsGrid';
import type { ApprovalsGridProps } from './ApprovalsGrid';
import { DEFAULT_PARAMS } from './schemas';

function renderGrid(props: Partial<ApprovalsGridProps> = {}, route = '/reports/approvals') {
  const onParams = vi.fn();
  const rendered = renderWithProviders(
    <ApprovalsGrid grid={approvalsGrid()} params={DEFAULT_PARAMS} onParams={onParams} {...props} />,
    { route, user: testUser(), path: '/reports/approvals' },
  );
  return { ...rendered, onParams };
}

describe('ApprovalsGrid (041 US2)', () => {
  it('shows the columns in order, one row per approver step, with a dash for what is missing', async () => {
    renderGrid();

    const headings = screen.getAllByRole('columnheader').map((h) => h.textContent);
    expect(headings).toEqual([
      'Record ID',
      'Record name',
      'Record type',
      'Approver',
      'Decided on',
      'Status',
      'Submitted on',
      'Level',
      'Purchase order',
      'Amount',
      'Owner',
      'Comments',
    ]);
    const waiting = screen.getByRole('row', { name: /Event stand/ });
    expect(within(waiting).getAllByText('—').length).toBeGreaterThanOrEqual(3);
    expect(within(waiting).getByText('4500001, 4500002')).toBeInTheDocument();
    const automatic = screen.getByRole('row', { name: /Webinar/ });
    expect(within(automatic).getByText('Approved automatically')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('links a record name only when the person may open it', () => {
    renderGrid();

    expect(screen.getByRole('link', { name: 'Spring launch' })).toHaveAttribute(
      'href',
      '/activities/A-1001',
    );
    expect(screen.getByText('Event stand')).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Event stand' })).toBeNull();
  });

  it('sorts by a heading, the second click turning the order round', async () => {
    const { user, onParams } = renderGrid({
      params: { ...DEFAULT_PARAMS, sort: 'amount', order: 'asc' },
    });

    expect(screen.getByRole('columnheader', { name: 'Amount' })).toHaveAttribute(
      'aria-sort',
      'ascending',
    );
    await user.click(screen.getByRole('button', { name: 'Amount' }));
    await user.click(screen.getByRole('button', { name: 'Owner' }));

    expect(onParams).toHaveBeenNthCalledWith(1, { sort: 'amount', order: 'desc' });
    expect(onParams).toHaveBeenNthCalledWith(2, { sort: 'owner', order: 'asc' });
  });

  it('filters by record type, status and approver, and clears a choice with All', async () => {
    const { user, onParams } = renderGrid({ params: { ...DEFAULT_PARAMS, outcomes: ['waiting'] } });

    await user.selectOptions(
      screen.getByRole('combobox', { name: 'Record type' }),
      'purchaseRequest',
    );
    await user.selectOptions(screen.getByRole('combobox', { name: 'Approver' }), 'U-1001');
    await user.selectOptions(screen.getByRole('combobox', { name: 'Status' }), '');

    expect(onParams).toHaveBeenCalledWith({ recordTypes: ['purchaseRequest'] });
    expect(onParams).toHaveBeenCalledWith({ approvers: ['U-1001'] });
    expect(onParams).toHaveBeenCalledWith({ outcomes: [] });
    expect(screen.getByRole('option', { name: 'Not needed' })).toBeInTheDocument();
  });

  it('searches after a pause in typing', async () => {
    const { user, onParams } = renderGrid();

    await user.type(screen.getByRole('searchbox', { name: 'Search approvals' }), 'stand');

    await waitFor(() => expect(onParams).toHaveBeenCalledWith({ q: 'stand' }));
    expect(onParams).toHaveBeenCalledTimes(1);
  });

  it('pages and changes the page size, naming the rows on screen', async () => {
    const grid = { ...approvalsGrid(), total: 120, page: 2, pageSize: 50 };
    const { user, onParams } = renderGrid({ grid });

    expect(screen.getByText('Rows 51–100 of 120')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Next' }));
    await user.click(screen.getByRole('button', { name: 'Previous' }));
    await user.selectOptions(screen.getByRole('combobox', { name: 'Rows per page' }), '10');

    expect(onParams).toHaveBeenCalledWith({ page: 3 });
    expect(onParams).toHaveBeenCalledWith({ page: 1 });
    expect(onParams).toHaveBeenCalledWith({ pageSize: 10 });
  });

  it('opens a long comment in full on the page and closes it again', async () => {
    const { user } = renderGrid();

    await user.click(screen.getByRole('button', { name: 'Read in full' }));
    const dialog = await screen.findByRole('alertdialog');

    expect(within(dialog).getByText(LONG_COMMENT)).toBeInTheDocument();
    await user.click(within(dialog).getByRole('button', { name: 'Close' }));
    await waitFor(() => expect(screen.queryByRole('alertdialog')).toBeNull());
  });

  it('says there are no approvals yet when nothing is filtered', () => {
    renderGrid({ grid: { ...approvalsGrid(), items: [], total: 0 } });

    expect(screen.getByText(/No approvals yet/)).toBeInTheDocument();
  });

  it('says the filters exclude everything, and Reset filters clears them', async () => {
    const { user, onParams } = renderGrid({
      grid: { ...approvalsGrid(), items: [], total: 0 },
      params: { ...DEFAULT_PARAMS, q: 'zzz', outcomes: ['waiting'] },
    });

    await user.click(screen.getByRole('button', { name: 'Reset filters' }));

    expect(screen.getByText('Your filters exclude everything.')).toBeInTheDocument();
    expect(onParams).toHaveBeenCalledWith({ q: '', recordTypes: [], approvers: [], outcomes: [] });
  });

  it('names the plans whose amounts the person cannot see', () => {
    renderGrid({ covers: { plans: 1, of: 2 } });
    expect(screen.getByText(/Amounts cover 1 of 2 plans/)).toBeInTheDocument();
  });

  it('says financial data is out of reach when no plan shows amounts', () => {
    renderGrid({ covers: { plans: 0, of: 2 } });

    expect(screen.getByText(/You can’t see financial data/)).toBeInTheDocument();
  });

  it('shows a loading skeleton', () => {
    renderGrid({ grid: undefined });

    expect(screen.getByText('Loading…')).toBeInTheDocument();
  });

  it('shows an error with Try again', async () => {
    const onRetry = vi.fn();
    const { user } = renderGrid({ grid: undefined, failed: true, onRetry });

    await user.click(screen.getByRole('button', { name: 'Try again' }));

    expect(onRetry).toHaveBeenCalledOnce();
  });
});
