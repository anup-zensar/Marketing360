import { fileURLToPath } from 'node:url';
import type { Locator, Page } from '@playwright/test';
import { expectNoAxeViolations, expectToast, signInAs, uniqueName } from './fixtures';
import { expect, test } from './support/no-csp-violations';

// Feature 051 US3 against the real API: Priya uploads the agency's event leads, corrects a column, charts
// the dataset beside SP&A One data and deletes it with Undo; files that break the rules store nothing.

const PRIYA = 'priya.raman@shell.com';
const XLSX = fileURLToPath(new URL('./fixtures/event-leads.xlsx', import.meta.url));
const CSV_TYPE = 'text/csv';
// The fixture's Leads column adds up to 1,355 (backend/seed/datasets.go, EventLeads).
const LEADS_TOTAL = '1,355';

function panel(page: Page): Locator {
  return page.getByRole('dialog', { name: 'Upload data' });
}

async function openUpload(page: Page) {
  await page.goto('/datasets');
  await expect(page.getByRole('heading', { level: 1, name: 'Datasets' })).toBeVisible();
  await page.getByRole('button', { name: 'Upload data' }).first().click();
  await expect(panel(page)).toBeVisible();
}

async function checkFile(page: Page, file: string | { name: string; buffer: Buffer }) {
  const chosen = typeof file === 'string' ? file : { ...file, mimeType: CSV_TYPE };
  await panel(page).getByLabel('Choose a file').setInputFiles(chosen);
  await panel(page).getByRole('button', { name: 'Check the file' }).click();
}

function column(page: Page, position: number): Locator {
  return panel(page).getByRole('group', { name: `Column ${position}` });
}

test('051 US3: Priya uploads event leads, corrects a column and charts them (FR-040 – FR-045)', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await page.goto('/dashboards');
  await page.getByRole('link', { name: 'Datasets', exact: true }).click();
  await expect(page.getByRole('heading', { level: 1, name: 'Datasets' })).toBeVisible();
  await expectNoAxeViolations(page);

  await page.getByRole('button', { name: 'Upload data' }).first().click();
  await checkFile(page, XLSX);
  await expect(panel(page)).toContainText('Step 2 of 2: Check the columns');
  await expect(panel(page).getByText(/^event-leads\.xlsx · Sheet '.+' · 12 rows$/)).toBeVisible();
  await expect(column(page, 1).getByLabel(/^Name/)).toHaveValue('Event');
  await expect(column(page, 3).getByLabel(/^Type/)).toHaveValue('number');
  await expect(panel(page).getByRole('region', { name: 'First rows' })).toContainText(
    'Lubricants Expo Hamburg',
  );

  const name = uniqueName('Event leads');
  await panel(page)
    .getByLabel(/^Dataset name/)
    .fill(name);
  await column(page, 4).getByLabel(/^Name/).fill('Agency cost');
  await column(page, 4).getByLabel(/^Type/).selectOption('money');
  await column(page, 4)
    .getByLabel(/^Currency/)
    .selectOption('EUR');
  await expectNoAxeViolations(page);
  await panel(page).getByRole('button', { name: 'Save dataset' }).click();

  await expectToast(page, `Dataset '${name}' saved`);
  await expect(page.getByRole('heading', { level: 1, name })).toBeVisible();
  await expect(page).toHaveURL(/\/datasets\/DS-\d+$/);
  const datasetPath = new URL(page.url()).pathname;
  const columns = page.getByRole('region', { name: 'Columns' });
  await expect(columns.getByRole('row', { name: /Agency cost Money EUR/ })).toBeVisible();
  await expectNoAxeViolations(page);

  await page.goto('/builder');
  await page.getByRole('button', { name: 'Add key figure' }).click();
  const inspector = page.getByRole('complementary', { name: 'Inspector' });
  await inspector.getByLabel(/^Source/).selectOption({ label: name });
  await inspector.getByRole('button', { name: 'Add measure' }).click();
  await inspector.getByLabel('What to measure').selectOption({ label: 'Leads' });
  await inspector.getByLabel('Aggregation').selectOption({ label: 'Sum' });
  const keyFigure = page.getByRole('group', { name: /^Widget: Key figure,/ });
  await expect(keyFigure.getByText(LEADS_TOTAL, { exact: true })).toBeVisible();

  await page.getByRole('button', { name: 'Dashboard details' }).click();
  const dashboard = uniqueName('Leads check');
  await inspector.getByLabel(/^Name/).fill(dashboard);
  await page.getByRole('button', { name: 'Save' }).click();
  await expectToast(page, 'Dashboard saved');

  await page.goto(datasetPath);
  await expect(page.getByRole('link', { name: dashboard })).toBeVisible();
  await page.getByRole('main').getByRole('button', { name: 'More', exact: true }).click();
  await page.getByRole('menuitem', { name: 'Delete' }).click();
  const confirm = page.getByRole('alertdialog', { name: `Delete dataset '${name}'?` });
  await expect(confirm).toContainText(
    "1 of your dashboards uses it; their widgets will show 'This data is no longer available'.",
  );
  await confirm.getByRole('button', { name: 'Delete dataset' }).click();
  await expect(page.getByRole('heading', { level: 1, name: 'Datasets' })).toBeVisible();
  await page.getByRole('button', { name: 'Undo' }).click();
  await expectToast(page, new RegExp(`${name}.* restored`));
  await expect(page.getByRole('link', { name })).toBeVisible();
});

test('051 US3-2: every cell is checked against its type and nothing is stored while one is wrong', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await openUpload(page);
  const rows = ['Event,Leads', 'Expo Hamburg,140', 'Fleet Day,n/a', 'Workshop Lyon,62'];
  await checkFile(page, { name: 'leads-with-gap.csv', buffer: Buffer.from(rows.join('\n')) });
  await expect(panel(page)).toContainText('Step 2 of 2: Check the columns');

  const name = uniqueName('Leads with a gap');
  await panel(page)
    .getByLabel(/^Dataset name/)
    .fill(name);
  await column(page, 2).getByLabel(/^Type/).selectOption('number');
  await panel(page).getByRole('button', { name: 'Save dataset' }).click();

  await expect(
    panel(page)
      .getByText(/Row \d+, Leads: 'n\/a' is not a number/)
      .first(),
  ).toBeVisible();
  await expect(panel(page)).toBeVisible();
  await expectNoAxeViolations(page);
});

test('051 US3-3: a file over a limit is refused with the reason and the limit', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await openUpload(page);
  const header = Array.from({ length: 51 }, (_, index) => `Column ${index + 1}`).join(',');
  const values = Array.from({ length: 51 }, (_, index) => String(index)).join(',');
  await checkFile(page, { name: 'too-wide.csv', buffer: Buffer.from(`${header}\n${values}`) });

  await expect(panel(page)).toContainText('Step 1 of 2: Choose a file');
  await expect(
    panel(page)
      .getByText(/50 columns/)
      .last(),
  ).toBeVisible();
  await expect(panel(page).getByLabel('Choose a file')).toHaveAttribute('aria-invalid', 'true');
});
