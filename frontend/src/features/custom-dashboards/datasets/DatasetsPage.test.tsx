import { screen, within } from '@testing-library/react';
import type { HttpHandler } from 'msw';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { datasetSummary, SHARED_DATASET } from '@/testing/fixtures/datasets';
import { testUser } from '@/testing/fixtures/users';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { apiError, pending, reply } from '@/testing/replies';
import { DatasetsPage } from './DatasetsPage';

const LIST = { body: { mine: [datasetSummary()], sharedWithMe: [SHARED_DATASET] } };

function renderPage(answers: HttpHandler[] = []) {
  mockApi([reply('GET /api/v1/datasets', LIST), ...answers]);
  return renderWithProviders(<DatasetsPage />, {
    route: '/datasets',
    path: '/datasets',
    user: testUser(),
  });
}

describe('051 FR-040: the Datasets page', () => {
  it('lists the person’s datasets and those shared with them', async () => {
    renderPage();

    expect(await screen.findByRole('link', { name: 'Event leads 2026' })).toHaveAttribute(
      'href',
      '/datasets/DS-1001',
    );
    expect(screen.getByRole('heading', { level: 1, name: 'Datasets' })).toBeInTheDocument();
    expect(document.title).toBe('Datasets · SP&A One');
    const mine = screen.getByRole('table', { name: 'My datasets' });
    expect(within(mine).getAllByRole('row')[1]).toHaveTextContent(
      /Event leads 2026124Not shared20 Oct 2026/,
    );
    const shared = screen.getByRole('table', { name: 'Shared with me' });
    expect(within(shared).getAllByRole('row')[1]).toHaveTextContent(
      /China retail panelDaniel Okoro1,2506Lubricants Europe/,
    );
    expect(screen.getByRole('button', { name: 'Upload data' })).toBeInTheDocument();
    await expectNoAxeViolations();
  });

  it('shows the loading state', () => {
    renderPage([reply('GET /api/v1/datasets', pending)]);

    expect(screen.getByLabelText('Loading datasets')).toBeInTheDocument();
  });

  it('shows an error with Try again, which loads the list', async () => {
    let calls = 0;
    const { user } = renderPage([
      reply('GET /api/v1/datasets', () => {
        calls += 1;
        return calls === 1 ? { status: 500, body: apiError('internal_error', 'Broken') } : LIST;
      }),
    ]);

    await user.click(await screen.findByRole('button', { name: 'Try again' }));

    expect(await screen.findByRole('link', { name: 'Event leads 2026' })).toBeInTheDocument();
  });

  it('says so when there are no datasets, and offers Upload data', async () => {
    renderPage([reply('GET /api/v1/datasets', { body: { mine: [], sharedWithMe: [] } })]);

    expect(await screen.findByText('No datasets yet')).toBeInTheDocument();
    expect(screen.getByText('Nothing shared with you yet.')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { name: 'Upload data' })).toHaveLength(2);
    await expectNoAxeViolations();
  });
});
