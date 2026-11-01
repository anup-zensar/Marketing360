import { expect, test } from '@playwright/test';
import { expectNoAxeViolations, signInAs } from './fixtures';

// Plan details (043 US1): the Activity level of the demo year inside the report frame, its widgets with their tables,
// and a value that opens All work filtered by it.

const PRIYA = 'priya.raman@shell.com';

test('043 US1: activities spread over each Shell attribute, and what is not filled in (FR-001 – FR-017)', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await page.goto('/reports');
  await page.getByRole('link', { name: 'Plan details' }).click();

  await expect(page.getByRole('heading', { level: 1, name: 'Plan details' })).toBeVisible();
  await expect(page.getByText(/^In view: /)).toBeVisible();
  await expect(
    page.getByRole('button', { name: 'Activity', exact: true, pressed: true }),
  ).toBeVisible();
  await expect(page.getByText('All filled in', { exact: true })).toBeVisible();
  const statuses = page.getByRole('list', { name: 'Activities by status' });
  await expect(statuses).toBeVisible();
  await expectNoAxeViolations(page);

  const status = page.getByRole('region', { name: 'Built-in fields' });
  await status.getByRole('button', { name: 'Table' }).first().click();
  await expect(page).toHaveURL(/table=activityTypeCount/);
  await expect(page.getByRole('table', { name: 'Activities by activity type' })).toBeVisible();

  await statuses.getByRole('link').first().click();
  await expect(page).toHaveURL(/\/all-work\?.*activityStatus=/);
});

test('043 US3: a widget exports as CSV and Excel, and the report as one workbook (FR-017; 039 FR-022)', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await page.goto('/reports/plan-details');
  const builtIn = page.getByRole('region', { name: 'Built-in fields' });
  await expect(builtIn).toBeVisible();

  for (const [format, extension] of [
    ['CSV (.csv)', '.csv'],
    ['Excel (.xlsx)', '.xlsx'],
  ] as const) {
    await builtIn.getByRole('button', { name: 'Export' }).first().click();
    const downloading = page.waitForEvent('download');
    await page.getByRole('menuitem', { name: format }).click();
    const file = await downloading;
    expect(file.suggestedFilename()).toMatch(
      new RegExp(
        `^SP&A One – Plan details – Activities by activity type – \\d{4} – .+\\${extension}$`,
      ),
    );
  }

  // The page header's Export comes before every widget's.
  await page.getByRole('button', { name: 'Export' }).first().click();
  const downloading = page.waitForEvent('download');
  await page.getByRole('menuitem', { name: 'Excel (.xlsx)' }).click();
  expect((await downloading).suggestedFilename()).toMatch(
    /^SP&A One – Plan details – \d{4} – .+\.xlsx$/,
  );
});

test('043 US2: the level is kept in the address and Back returns to it (FR-001)', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await page.goto('/reports/plan-details');
  await expect(page.getByRole('heading', { level: 1, name: 'Plan details' })).toBeVisible();

  await page.getByRole('button', { name: 'Campaign', exact: true }).click();
  await expect(page).toHaveURL(/show=campaign/);
  await expect(page.getByRole('list', { name: 'Campaigns by approval progress' })).toBeVisible();
  await page.goBack();
  await expect(
    page.getByRole('button', { name: 'Activity', exact: true, pressed: true }),
  ).toBeVisible();
});
