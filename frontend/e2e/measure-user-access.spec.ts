import { expect, test } from '@playwright/test';
import { expectNoAxeViolations, signInAs } from './fixtures';

// User access (045 US1): a person signs in, then the administrator finds them in the report with that sign-in, in
// the report page without the plan filters.

const PRIYA = 'priya.raman@shell.com';
const ALEX = 'alex.morgan@shell.com';

test('045 US1: a sign-in shows in the report the administrator opens (FR-001 – FR-009, SC-002)', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await signInAs(page, ALEX);
  await page.goto('/reports');
  await page.getByRole('link', { name: 'User access' }).click();

  await expect(page.getByRole('heading', { level: 1, name: 'User access' })).toBeVisible();
  const table = page.getByRole('table', { name: 'People' });
  const priya = table.getByRole('row', { name: /Priya Raman/ });
  await expect(priya).toBeVisible();
  await expect(priya).not.toContainText('Never');
  await expect(page.getByRole('combobox', { name: 'Currency' })).toHaveCount(0);
  await expectNoAxeViolations(page);

  await page.getByRole('searchbox', { name: 'Search people' }).fill('priya');
  await expect(page).toHaveURL(/q=priya/);
  await expect(table.getByRole('row')).toHaveCount(2);

  await table.getByRole('link', { name: 'Priya Raman' }).click();
  await expect(page).toHaveURL(/\/settings\/people\/U-\d+$/);
});

test('045 US2: the people export as CSV and Excel, and the report as one workbook (FR-013)', async ({
  page,
}) => {
  await signInAs(page, ALEX);
  await page.goto('/reports/user-access?status=all&sort=-lastSignedIn');
  const people = page.getByRole('region', { name: 'People', exact: true });
  await expect(people.getByRole('table', { name: 'People' })).toBeVisible();

  for (const [format, extension] of [
    ['CSV (.csv)', 'csv'],
    ['Excel (.xlsx)', 'xlsx'],
  ] as const) {
    await people.getByRole('button', { name: 'Export' }).click();
    const downloading = page.waitForEvent('download');
    await page.getByRole('menuitem', { name: format }).click();
    expect((await downloading).suggestedFilename()).toMatch(
      new RegExp(`^SP&A One – User access – People – .+[.]${extension}$`),
    );
  }

  // The page header's Export comes before the table's.
  await page.getByRole('button', { name: 'Export' }).first().click();
  const downloading = page.waitForEvent('download');
  await page.getByRole('menuitem', { name: 'Excel (.xlsx)' }).click();
  expect((await downloading).suggestedFilename()).toMatch(/^SP&A One – User access – .+[.]xlsx$/);
});
