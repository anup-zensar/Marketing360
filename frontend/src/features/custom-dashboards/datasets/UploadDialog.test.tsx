import { fireEvent, screen, waitFor, within } from '@testing-library/react';
import type { HttpHandler } from 'msw';
import { describe, expect, it } from 'vitest';
import { expectNoAxeViolations } from '@/testing/axe';
import { CURRENCIES, dataset, preview } from '@/testing/fixtures/datasets';
import { mockApi } from '@/testing/mock-api';
import { renderWithProviders } from '@/testing/render';
import { testUser } from '@/testing/fixtures/users';
import { apiError, reply } from '@/testing/replies';
import { DatasetsPage } from './DatasetsPage';

const XLSX = new File(['PK-sheet'], 'event-leads.xlsx', {
  type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
});

function renderUpload(answers: HttpHandler[] = []) {
  const api = mockApi([
    reply('GET /api/v1/datasets', { body: { mine: [], sharedWithMe: [] } }),
    reply('GET /api/v1/me/display-currencies', { body: CURRENCIES }),
    reply('POST /api/v1/datasets/preview', { body: preview() }),
    reply('POST /api/v1/datasets', { status: 201, body: dataset({ id: 'DS-1007' }) }),
    ...answers,
  ]);
  const view = renderWithProviders(<DatasetsPage />, {
    route: '/datasets',
    path: '/datasets',
    user: testUser(),
  });
  return { ...view, api };
}

async function checkFile(view: ReturnType<typeof renderUpload>, file: File = XLSX) {
  await view.user.click(screen.getAllByRole('button', { name: 'Upload data' })[0] ?? document.body);
  // fireEvent, not user.upload: user-event drops files the input's accept list leaves out.
  fireEvent.change(await screen.findByLabelText('Choose a file'), { target: { files: [file] } });
  await view.user.click(screen.getByRole('button', { name: 'Check the file' }));
}

async function columnsStep(view: ReturnType<typeof renderUpload>) {
  await checkFile(view);
  await screen.findByLabelText('Dataset name');
  return screen.getByRole('dialog', { name: 'Upload data' });
}

describe('051 FR-041: checking a file (US3 scenario 1)', () => {
  it('shows the file, sheet, rows, proposed column types with samples and the first rows; stores nothing', async () => {
    const view = renderUpload();

    const panel = await columnsStep(view);

    expect(within(panel).getByText("event-leads.xlsx · Sheet 'Leads' · 12 rows")).toBeVisible();
    expect(within(panel).getByLabelText('Dataset name')).toHaveValue('event-leads');
    const columns = within(panel).getByRole('list', { name: 'Columns' });
    const cost = within(columns).getAllByRole('group')[3] ?? panel;
    expect(within(cost).getByLabelText('Name')).toHaveValue('Cost');
    expect(within(cost).getByLabelText(/^Type/)).toHaveValue('money');
    expect(cost).toHaveTextContent('Samples: 12000 · 9500 (1 empty cell)');
    const rows = within(panel).getByRole('table');
    expect(within(rows).getAllByRole('row')).toHaveLength(3);
    expect(view.api.bodiesOf('POST /api/v1/datasets/preview')).toEqual([
      { fileName: 'event-leads.xlsx', content: btoa('PK-sheet') },
    ]);
    expect(view.api.bodiesOf('POST /api/v1/datasets')).toEqual([]);
    await expectNoAxeViolations(panel);
  });

  it('starts again from Choose another file', async () => {
    const view = renderUpload();
    await checkFile(view);

    await view.user.click(await screen.findByRole('button', { name: 'Choose another file' }));

    expect(await screen.findByLabelText('Choose a file')).toBeInTheDocument();
  });
});

describe('051 FR-043: files outside the limits (US3 scenario 3)', () => {
  it.each([
    [
      'a file that is not CSV or .xlsx',
      new File(['x'], 'leads.pdf'),
      'SP&A One can read .xlsx and CSV (UTF-8) files.',
    ],
    [
      'a file over 5 MB',
      new File([new Uint8Array(5 * 1024 * 1024 + 1)], 'big.csv'),
      'The file is larger than 5 MB.',
    ],
  ])('refuses %s before sending it', async (_case, file, message) => {
    const view = renderUpload();

    await checkFile(view, file);

    expect(await screen.findAllByText(message)).not.toHaveLength(0);
    expect(view.api.bodiesOf('POST /api/v1/datasets/preview')).toEqual([]);
  });

  it('shows every reason the API gives for refusing the file', async () => {
    const view = renderUpload([
      reply('POST /api/v1/datasets/preview', {
        status: 400,
        body: apiError('validation_failed', 'Check the file.', {
          fields: [
            { field: 'content', message: 'The file has more than 10,000 rows.' },
            { field: 'content', message: 'The file has more than 50 columns.' },
          ],
        }),
      }),
    ]);

    await checkFile(view);

    expect(await screen.findAllByText('The file has more than 10,000 rows.')).not.toHaveLength(0);
    expect(screen.getAllByText('The file has more than 50 columns.')).not.toHaveLength(0);
    expect(screen.getByText('2 problems need your attention')).toBeInTheDocument();
  });

  it('asks for a file when none is chosen', async () => {
    const view = renderUpload();
    await view.user.click(
      screen.getAllByRole('button', { name: 'Upload data' })[0] ?? document.body,
    );

    await view.user.click(await screen.findByRole('button', { name: 'Check the file' }));

    expect(await screen.findAllByText('Choose a CSV or Excel .xlsx file.')).not.toHaveLength(0);
  });
});

describe('051 FR-042: saving the dataset (US3 scenario 2)', () => {
  it('saves the chosen names, types and currency, then opens the dataset', async () => {
    const view = renderUpload();
    const panel = await columnsStep(view);
    const name = within(panel).getByLabelText('Dataset name');
    await view.user.clear(name);
    await view.user.type(name, 'Event leads 2026');
    const cost = within(within(panel).getByRole('list', { name: 'Columns' })).getAllByRole(
      'group',
    )[3];
    if (!cost) throw new Error('No cost column');
    await view.user.clear(within(cost).getByLabelText('Name'));
    await view.user.type(within(cost).getByLabelText('Name'), 'Agency cost');
    await view.user.selectOptions(within(cost).getByLabelText(/^Currency/), 'EUR');

    await view.user.click(within(panel).getByRole('button', { name: 'Save dataset' }));

    await waitFor(() => expect(view.router.state.location.pathname).toBe('/datasets/DS-1007'));
    expect(screen.getByText("Dataset 'Event leads 2026' saved")).toBeInTheDocument();
    expect(view.api.bodiesOf('POST /api/v1/datasets')).toEqual([
      {
        fileName: 'event-leads.xlsx',
        content: btoa('PK-sheet'),
        name: 'Event leads 2026',
        description: '',
        columns: [
          { position: 0, label: 'Event', kind: 'text', currency: null },
          { position: 1, label: 'Date', kind: 'date', currency: null },
          { position: 2, label: 'Leads', kind: 'number', currency: null },
          { position: 3, label: 'Agency cost', kind: 'money', currency: 'EUR' },
        ],
      },
    ]);
  });

  it('lists every problem of the form at once and sends nothing', async () => {
    const view = renderUpload();
    const panel = await columnsStep(view);
    await view.user.clear(within(panel).getByLabelText('Dataset name'));
    const groups = within(within(panel).getByRole('list', { name: 'Columns' })).getAllByRole(
      'group',
    );
    const leads = groups[2];
    if (!leads) throw new Error('No leads column');
    await view.user.clear(within(leads).getByLabelText('Name'));
    await view.user.type(within(leads).getByLabelText('Name'), 'event');

    await view.user.click(within(panel).getByRole('button', { name: 'Save dataset' }));

    expect(await screen.findByText('2 problems need your attention')).toBeInTheDocument();
    expect(screen.getAllByText('Enter a name.')).not.toHaveLength(0);
    expect(screen.getAllByText('Another column has this name.')).not.toHaveLength(0);
    expect(view.api.bodiesOf('POST /api/v1/datasets')).toEqual([]);
  });

  it('shows every cell problem the API finds, under its column, and stores nothing', async () => {
    const view = renderUpload([
      reply('POST /api/v1/datasets', {
        status: 400,
        body: apiError('validation_failed', 'Check the columns.', {
          fields: [
            { field: 'columns[2]', message: "Row 14: 'n/a' is not a number." },
            { field: 'columns[2]', message: '… and 37 more problems in this column.' },
            { field: 'name', message: 'You already have a dataset with this name.' },
          ],
        }),
      }),
    ]);
    await checkFile(view);

    await view.user.click(await screen.findByRole('button', { name: 'Save dataset' }));

    expect(await screen.findByText('3 problems need your attention')).toBeInTheDocument();
    const leads = within(screen.getByRole('list', { name: 'Columns' })).getAllByRole('group')[2];
    expect(leads).toHaveTextContent("Row 14: 'n/a' is not a number.");
    expect(leads).toHaveTextContent('… and 37 more problems in this column.');
    expect(view.router.state.location.pathname).toBe('/datasets');
  });
});
