import { screen, waitFor, within } from '@testing-library/react';
import type { HttpHandler } from 'msw';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { dataset, READER_ACCESS, rowPage } from '@/testing/fixtures/datasets';
import { deleted } from '@/testing/fixtures/settings';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { findMessages, renderWithProviders } from '@/testing/render';
import { apiError, pending, reply } from '@/testing/replies';
import type { Answer } from '@/testing/replies';
import { DatasetPage } from './DatasetPage';

function renderPage(
  answer: Answer<'GET /api/v1/datasets/{datasetId}'> = { body: dataset() },
  answers: HttpHandler[] = [],
) {
  const api = mockApi([
    reply('GET /api/v1/datasets/{datasetId}', answer),
    reply('GET /api/v1/datasets/{datasetId}/rows', (request) => {
      const offset = Number(new URL(request.url).searchParams.get('offset'));
      return { body: rowPage(offset, 60, offset === 0 ? 50 : 10) };
    }),
    ...answers,
  ]);
  const view = renderWithProviders(<DatasetPage />, {
    route: '/datasets/DS-1001',
    path: '/datasets/:datasetId',
    user: testUser(),
  });
  return { ...view, api };
}

describe('051 FR-040, FR-044: a dataset’s page (US3 scenario 5)', () => {
  it('shows its details, columns, rows and the dashboards that use it', async () => {
    renderPage();

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Event leads 2026' }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("By Priya Raman · Uploaded from event-leads.xlsx, sheet 'Leads' · 12 rows"),
    ).toBeInTheDocument();
    const columns = screen.getByRole('table', { name: 'Columns of the dataset' });
    expect(within(columns).getAllByRole('row')[4]).toHaveTextContent('Agency costMoneyEUR');
    expect(within(columns).getAllByRole('row')[2]).toHaveTextContent('DateDate—');
    expect(screen.getByRole('link', { name: 'Lubricants Europe overview' })).toHaveAttribute(
      'href',
      '/dashboards/DB-1001',
    );
    expect(await screen.findByText('Rows 1–50 of 60')).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('pages through the rows', async () => {
    const { user, api } = renderPage();

    await user.click(await screen.findByRole('button', { name: 'Next' }));

    expect(await screen.findByText('Rows 51–60 of 60')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Next' })).not.toBeInTheDocument();
    expect(api.requests.map((r) => r.search)).toContain('?offset=50&limit=50');
    await user.click(screen.getByRole('button', { name: 'Previous' }));
    expect(await screen.findByText('Rows 1–50 of 60')).toBeInTheDocument();
  });

  it('shows a shared dataset view-only: no Edit, Sharing or Delete', async () => {
    renderPage({ body: dataset({ isOwner: false, access: READER_ACCESS, shareableTeams: [] }) });

    await screen.findByRole('heading', { level: 1, name: 'Event leads 2026' });
    expect(screen.queryByRole('button', { name: 'Edit' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Sharing' })).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'More' })).not.toBeInTheDocument();
  });

  it('says "Dataset not found" for one the person may not see', async () => {
    renderPage({ status: 404, body: apiError('not_found', 'Not found') });

    expect(await screen.findByText('Dataset not found')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Go to Datasets' })).toHaveAttribute(
      'href',
      '/datasets',
    );
  });

  it('shows the loading state, and an error with Try again', async () => {
    let calls = 0;
    const { user } = renderPage(() => {
      calls += 1;
      return calls === 1
        ? { status: 500, body: apiError('internal_error', 'Broken') }
        : { body: dataset() };
    });

    expect(screen.getByLabelText('Loading the dataset')).toBeInTheDocument();
    await user.click(await screen.findByRole('button', { name: 'Try again' }));

    expect(
      await screen.findByRole('heading', { level: 1, name: 'Event leads 2026' }),
    ).toBeInTheDocument();
  });

  it('shows the rows loading in their own place', async () => {
    renderPage(undefined, [reply('GET /api/v1/datasets/{datasetId}/rows', pending)]);

    expect(await screen.findByLabelText('Loading rows')).toBeInTheDocument();
  });
});

describe('051 FR-040, FR-044: changing a dataset', () => {
  it('renames it and its columns; types stay as saved', async () => {
    const { user, api } = renderPage(undefined, [
      reply('PATCH /api/v1/datasets/{datasetId}', { body: dataset({ name: 'Leads 2026' }) }),
    ]);

    await user.click(await screen.findByRole('button', { name: 'Edit' }));
    const panel = await screen.findByRole('dialog', { name: 'Edit dataset' });
    expect(panel).toHaveTextContent('Column types are fixed once a dataset is saved.');
    await user.clear(within(panel).getByLabelText('Name'));
    await user.type(within(panel).getByLabelText('Name'), 'Leads 2026');
    await user.clear(within(panel).getByLabelText('Column 3'));
    await user.type(within(panel).getByLabelText('Column 3'), 'Leads count');
    await user.click(within(panel).getByRole('button', { name: 'Save' }));

    expect(await screen.findByText("Dataset 'Leads 2026' saved")).toBeInTheDocument();
    expect(api.bodiesOf('PATCH /api/v1/datasets/DS-1001')).toEqual([
      {
        version: 2,
        name: 'Leads 2026',
        description: 'Leads from agency events.',
        columns: [{ position: 2, label: 'Leads count' }],
      },
    ]);
  });

  it('shares it with the owner’s teams (FR-046)', async () => {
    const { user, api } = renderPage(undefined, [
      reply('PUT /api/v1/datasets/{datasetId}/sharing', { body: dataset() }),
    ]);

    await user.click(await screen.findByRole('button', { name: 'Sharing' }));
    const dialog = await screen.findByRole('dialog', { name: "Share 'Event leads 2026'" });
    expect(dialog).toHaveTextContent('People in these teams can view this dataset.');
    await user.click(within(dialog).getByRole('checkbox', { name: 'Lubricants Europe' }));
    await user.click(within(dialog).getByRole('button', { name: 'Save' }));

    expect(await screen.findByText('Sharing saved')).toBeInTheDocument();
    expect(api.bodiesOf('PUT /api/v1/datasets/DS-1001/sharing')).toEqual([
      { version: 2, teamIds: ['T-1001'] },
    ]);
  });

  it('keeps the sharing dialog open with the refusal', async () => {
    const { user } = renderPage(undefined, [
      reply('PUT /api/v1/datasets/{datasetId}/sharing', {
        status: 400,
        body: apiError('validation_failed', 'You can share only with your teams.'),
      }),
    ]);

    await user.click(await screen.findByRole('button', { name: 'Sharing' }));
    await user.click(await screen.findByRole('button', { name: 'Save' }));

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'You can share only with your teams.',
    );
  });

  it('deletes it after saying how many dashboards use it, with Undo (US3 scenario 6, FR-045)', async () => {
    const { user, router } = renderPage(undefined, [
      reply('DELETE /api/v1/datasets/{datasetId}', {
        body: deleted('dataset', 'DS-1001', 'Event leads 2026'),
      }),
    ]);

    await user.click(await screen.findByRole('button', { name: 'More' }));
    await user.click(await screen.findByRole('menuitem', { name: 'Delete' }));
    const dialog = await screen.findByRole('alertdialog', {
      name: "Delete dataset 'Event leads 2026'?",
    });
    expect(dialog).toHaveTextContent(
      "1 of your dashboards uses it; their widgets will show 'This data is no longer available'. It moves to the recycle bin; you can restore it.",
    );
    await user.click(within(dialog).getByRole('button', { name: 'Delete dataset' }));

    await waitFor(() => expect(router.state.location.pathname).toBe('/datasets'));
    expect(await screen.findByText("Dataset 'Event leads 2026' deleted")).toBeInTheDocument();
    expect(within(await findMessages()).getByRole('button', { name: 'Undo' })).toBeVisible();
  });
});
