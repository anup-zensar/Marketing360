import { readFile } from 'node:fs/promises';
import { expect, test, type Page } from '@playwright/test';
import { callApi, expectNoAxeViolations, expectToast, signInAs, uniqueName } from './fixtures';
import { sheetNames } from './xlsx-read';

// The Reports area's shared frame (039): the Reports page and a report's filters (US1), saved views (US2) and the
// export (US3), on Plan budget's page, which shows the frame until feature 040 fills it.

const CARLOS = 'carlos.mendes@shell.com';
const PRIYA = 'priya.raman@shell.com';
const DANIEL = 'daniel.okoro@shell.com';
const REPORT = '/reports/plan-budget?year=2026';

const inView = (page: Page) => page.getByText(/^In view: /);

async function mobilityPlanId(page: Page): Promise<string> {
  await signInAs(page, CARLOS);
  const plans = await callApi<{ items: { id: string; name: string }[] }>(
    page,
    'GET',
    '/api/v1/plans',
  );
  const mobility = plans.items.find((plan) => plan.name === 'FY2026 Mobility China');
  if (!mobility) throw new Error('the seed has no FY2026 Mobility China');
  return mobility.id;
}

test('039 US1: a report opens on the defaults, filters on Apply and keeps everything in its address (FR-001 – FR-019)', async ({
  page,
}) => {
  const outOfScope = await mobilityPlanId(page);
  await signInAs(page, PRIYA);
  await page.goto('/');
  const links = (await page.getByRole('link').allInnerTexts()).map((text) => text.trim());
  expect(links.indexOf('Reports')).toBe(links.indexOf('Dashboards') + 1);
  await page.getByRole('link', { name: 'Reports', exact: true }).click();
  await expect(page.getByRole('heading', { level: 1, name: 'Reports' })).toBeVisible();
  await expectNoAxeViolations(page);

  const allWork = await callApi<{ count: { total: number } }>(
    page,
    'GET',
    '/api/v1/all-work?year=2026',
  );
  await page.goto(REPORT);
  await expect(inView(page)).toContainText(` ${allWork.count.total} activit`);
  await expectNoAxeViolations(page);
  const levels = page.getByRole('tablist', { name: 'Shell attributes by level' });
  await levels.scrollIntoViewIfNeeded();
  for (const level of ['Plan', 'Campaign', 'Initiative', 'Activity', 'Line item']) {
    await expect(levels.getByRole('tab', { name: level })).toBeInViewport({ ratio: 1 });
  }

  const unfiltered = await inView(page).innerText();
  const status = page.getByRole('group', { name: /^Activity status/ });
  await status.getByRole('checkbox', { name: 'Created' }).uncheck();
  await expect(page.getByText('Changes not applied yet')).toBeVisible();
  await page.getByRole('button', { name: 'Apply' }).click();
  await expect(page).toHaveURL(/activityStatus=/);
  await expect(inView(page)).not.toHaveText(unfiltered);
  const filtered = await inView(page).innerText();
  await expect(page.getByRole('button', { name: /^Remove filter: Activity status/ })).toBeVisible();

  const copy = await page.context().newPage();
  await copy.goto(page.url());
  await expect(inView(copy)).toHaveText(filtered);
  await copy.close();

  await page.getByRole('combobox', { name: 'Currency' }).selectOption('EUR');
  await expect(page).toHaveURL(/currency=EUR/);
  await page.getByRole('button', { name: 'Reset filters' }).first().click();
  await expect(page).not.toHaveURL(/activityStatus=/);
  await expect(page).toHaveURL(/currency=EUR/);

  await page.goto(`${REPORT}&plan=${outOfScope}`);
  await expect(
    page.getByText(/1 filter value is no longer available and was left out/),
  ).toBeVisible();
  await expect(page.getByText('FY2026 Mobility China')).toHaveCount(0);
});

test('039 US2: saved views are personal, open on the default, and come back with Undo (FR-024 – FR-028)', async ({
  page,
}) => {
  const name = uniqueName('Approved only');
  await signInAs(page, PRIYA);
  await page.goto(`${REPORT}&activityStatus=approved`);
  await page.getByRole('button', { name: 'Saved views' }).click();
  await page.getByRole('menuitem', { name: 'Save current filters…' }).click();
  await page.getByRole('textbox', { name: /Name/ }).fill(name);
  await page.getByRole('checkbox', { name: 'Open this report on this view' }).check();
  await page.getByRole('button', { name: 'Save view' }).click();
  await expectToast(page, `View '${name}' saved`);

  await page.goto('/reports/plan-budget');
  await expect(page).toHaveURL(/activityStatus=approved/);
  const button = page.getByRole('button', { name: `View: ${name}` });
  await expect(button).toBeVisible();

  await page.getByRole('combobox', { name: 'Currency' }).selectOption('EUR');
  await page.getByRole('button', { name: `View: ${name} Modified` }).click();
  await page.getByRole('menuitem', { name: 'Update view…' }).click();
  await page.getByRole('button', { name: 'Update view' }).click();
  await expectToast(page, `View '${name}' updated`);

  await page.getByRole('button', { name: `View: ${name}` }).click();
  await page.getByRole('menuitem', { name: 'Manage views…' }).click();
  const panel = page.getByRole('dialog', { name: 'Your saved views' });
  await panel.getByRole('button', { name: `Delete ${name}` }).click();
  await expectToast(page, `View '${name}' deleted`);
  await page.getByRole('button', { name: 'Undo' }).click();
  await expectToast(page, `View '${name}' restored`);

  await signInAs(page, DANIEL);
  await page.goto(REPORT);
  await page.getByRole('button', { name: 'Saved views' }).click();
  await expect(page.getByRole('menuitemcheckbox', { name })).toHaveCount(0);
});

test('039 US3: Export downloads the report with In view and About this export (FR-022)', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await page.goto(REPORT);
  await expect(inView(page)).toBeVisible();

  await page.getByRole('button', { name: 'Export', exact: true }).first().click();
  const downloading = page.waitForEvent('download');
  await page.getByRole('menuitem', { name: 'Excel (.xlsx)' }).click();
  const download = await downloading;

  expect(download.suggestedFilename()).toMatch(/\.xlsx$/);
  const sheets = sheetNames(await readFile(await download.path()));
  expect(sheets).toEqual(expect.arrayContaining(['In view', 'About this export']));
});
