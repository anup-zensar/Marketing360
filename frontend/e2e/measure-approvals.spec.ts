import { readFile } from 'node:fs/promises';
import { expect, test } from '@playwright/test';
import { expectNoAxeViolations, signInAs } from './fixtures';

// The Approvals report (041): the status mix and average days (US1), then the grid and its export (US2).

const ALEX = 'alex.morgan@shell.com';
const PRIYA = 'priya.raman@shell.com';
const YEAR = 2026;

const search = (page: { url(): string }) => new URL(page.url()).searchParams;

test('041 US1: the status mix and the average days, per record type, inside the report frame (FR-001 – FR-011)', async ({
  page,
}) => {
  await signInAs(page, ALEX);
  await page.goto(`/reports/approvals?year=${YEAR}`);

  await expect(page.getByRole('heading', { level: 1, name: 'Approvals' })).toBeVisible();
  const status = page.getByRole('region', { name: 'Current approval status' });
  await expect(status).toBeVisible();
  await expect(status.getByText(/\d+ waiting/)).toBeVisible();
  await expect(status.getByText('Approved automatically')).toBeVisible();
  await expect(page.getByRole('region', { name: 'Average days to a decision' })).toBeVisible();
  await expect(page.getByRole('group', { name: 'Period' })).toHaveCount(0);
  await expectNoAxeViolations(page);

  await status.getByRole('button', { name: 'Table' }).click();
  await expect(status.getByRole('table')).toContainText('Total');
  await status.getByRole('button', { name: 'Chart' }).click();

  const types = status.getByRole('group', { name: 'Record type' });
  if (await types.isVisible()) {
    await types.getByRole('button', { name: 'Purchase requests' }).click();
    await expect.poll(() => search(page).get('statusType')).toBe('purchaseRequest');
    await expect(status.getByText(/\d+ waiting/)).toBeVisible();
  }

  await signInAs(page, PRIYA);
  await page.goto(`/reports/approvals?year=${YEAR}`);
  await expect(page.getByRole('heading', { level: 1, name: 'Approvals' })).toBeVisible();
  await expect(page.getByRole('region', { name: 'Current approval status' })).toBeVisible();
});

test('041 US2: the grid sorts, searches, pages and exports every matching row (FR-012 – FR-023)', async ({
  page,
}) => {
  await signInAs(page, ALEX);
  await page.goto(`/reports/approvals?year=${YEAR}`);

  const grid = page
    .locator('section', {
      has: page.getByRole('heading', { level: 2, name: 'Approvals', exact: true }),
    })
    .last();
  await expect(grid.getByRole('table')).toBeVisible();
  await expect(grid.getByRole('columnheader').first()).toHaveText('Record ID');
  await expect(grid.getByRole('columnheader', { name: 'Comments' })).toBeVisible();
  await expectNoAxeViolations(page);

  await grid.getByRole('button', { name: 'Owner' }).click();
  await expect.poll(() => search(page).get('sort')).toBe('owner');
  await expect(grid.getByRole('columnheader', { name: 'Owner' })).toHaveAttribute(
    'aria-sort',
    'ascending',
  );

  await grid.getByRole('combobox', { name: 'Rows per page' }).selectOption('5');
  await expect.poll(() => search(page).get('pageSize')).toBe('5');
  await expect(grid.getByText(/^Rows 1–\d+ of \d+$/)).toBeVisible();

  await grid.getByRole('combobox', { name: 'Status' }).selectOption('waiting');
  await expect.poll(() => search(page).get('outcome')).toBe('waiting');
  await page.goBack();
  await expect.poll(() => search(page).get('outcome')).toBeNull();

  await grid.getByRole('searchbox', { name: 'Search approvals' }).fill('zzzz-no-such-record');
  await expect(grid.getByText('Your filters exclude everything.')).toBeVisible();
  await grid.getByRole('button', { name: 'Reset filters' }).click();
  await expect(grid.getByRole('table')).toBeVisible();

  await grid.getByRole('button', { name: 'Export' }).click();
  const downloading = page.waitForEvent('download');
  await page.getByRole('menuitem', { name: /CSV/ }).click();
  const download = await downloading;
  const file = (await readFile(await download.path())).toString('utf8');
  expect(download.suggestedFilename()).toMatch(/\.csv$/);
  expect(file).toContain('Item ID');
});
