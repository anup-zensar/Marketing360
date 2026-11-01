import { readFile } from 'node:fs/promises';
import type { Locator, Page } from '@playwright/test';
import { expectNoAxeViolations, signInAs } from './fixtures';
import { expect, test } from './support/no-csp-violations';

// Feature 051 US2 against the real API: Daniel reads the demo dashboard Priya shared with Lubricants
// Europe (DB-1001, seed); Mei, outside that team, is told it does not exist.

const DANIEL = 'daniel.okoro@shell.com';
const MEI = 'mei.lin@shell.com';
const DASHBOARD = 'Lubricants Europe overview';
const PNG_SIGNATURE = [0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a];

function widget(page: Page, title: string): Locator {
  return page.getByRole('region', { name: title, exact: true });
}

/** Points at the middle of the chart and reads its tooltip: one row per series, label then value. */
async function tooltipRows(
  page: Page,
  chart: Locator,
): Promise<{ labels: string[]; text: string }> {
  const box = await chart.boundingBox();
  if (!box) {
    throw new Error('The chart has no size');
  }
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
  const rows = chart.locator('li');
  await expect(rows.first()).toBeVisible();
  const labels = await rows.locator('span:nth-child(2)').allTextContents();
  const text = (await rows.allTextContents()).join(' ');
  await page.mouse.move(0, 0);
  return { labels, text };
}

async function openWidgetMenu(page: Page, title: string, item: string) {
  await widget(page, title)
    .getByRole('button', { name: `Actions for ${title}` })
    .click();
  await page.getByRole('menuitem', { name: item }).click();
}

test('051 US2: Daniel opens Dashboards, the built-ins and the dashboard shared with his team (FR-003, FR-036 – FR-039)', async ({
  page,
}) => {
  await signInAs(page, DANIEL);
  await page.goto('/');
  await page
    .getByRole('navigation', { name: 'Main' })
    .getByRole('link', { name: 'Dashboards' })
    .click();
  await expect(page.getByRole('heading', { level: 1, name: 'Dashboards' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Built-in' })).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Shared with me' })).toBeVisible();
  await expectNoAxeViolations(page);

  await page.getByRole('link', { name: 'Executive dashboard' }).click();
  await expect(page.getByRole('heading', { level: 1, name: 'Executive dashboard' })).toBeVisible();
  await expect(page).toHaveURL(/\/dashboards\/executive/);
  await page.goBack();
  await page.getByRole('link', { name: 'Regional dashboard' }).click();
  await expect(page.getByRole('heading', { level: 1, name: 'Regional dashboard' })).toBeVisible();
  await page.goBack();

  await page.getByRole('link', { name: DASHBOARD }).click();
  await expect(page.getByRole('heading', { level: 1, name: DASHBOARD })).toBeVisible();
  await expect(page).toHaveURL(/\/dashboards\/DB-\d+$/);
  await expect(page.getByText('By Priya Raman')).toBeVisible();
  await expect(page.getByRole('link', { name: 'Edit' })).toHaveCount(0);
  await expect(page.getByRole('button', { name: 'Make a copy' })).toBeVisible();
  await expect(
    widget(page, 'Event leads by month').getByText("You don't have access to this data"),
  ).toBeVisible();
  await expect(widget(page, 'Event leads by month')).not.toContainText('Leads');
  const bars = widget(page, 'Planned cost by activity type').getByRole('img');
  await expect(bars).toHaveAccessibleName(/^Bar chart: Planned cost by activity type\./);
  await expectNoAxeViolations(page);

  expect((await tooltipRows(page, bars)).text).toMatch(/\$[\d,]+\.\d{2}/);

  const lines = widget(page, 'Planned cost by month').getByRole('img');
  const [hidden = ''] = (await tooltipRows(page, lines)).labels;
  await lines.locator('svg text').filter({ hasText: hidden }).first().click();
  expect((await tooltipRows(page, lines)).labels).not.toContain(hidden);

  await openWidgetMenu(page, 'Planned cost by month', 'Show as table');
  const asTable = page.getByRole('region', { name: 'Planned cost by month, table' });
  await expect(asTable.getByRole('columnheader').first()).toBeVisible();
  await expectNoAxeViolations(page);
  await openWidgetMenu(page, 'Planned cost by month', 'Show as chart');
  await expect(asTable).toHaveCount(0);
  await expect(lines).toBeVisible();

  const downloading = page.waitForEvent('download');
  await openWidgetMenu(page, 'Planned cost by activity type', 'Download as PNG');
  const download = await downloading;
  expect(download.suggestedFilename()).toBe(
    `SP&A One – ${DASHBOARD} – Planned cost by activity type.png`,
  );
  const png = await readFile(await download.path());
  expect([...png.subarray(0, PNG_SIGNATURE.length)]).toEqual(PNG_SIGNATURE);
});

test('051 US2-7: the filters live in the address, so Back shows the same view (FR-037)', async ({
  page,
}) => {
  await signInAs(page, DANIEL);
  await page.goto('/dashboards');
  await page.getByRole('link', { name: DASHBOARD }).click();
  await expect(page.getByRole('heading', { level: 1, name: DASHBOARD })).toBeVisible();

  const filters = page.getByRole('group', { name: 'Dashboard filters' });
  await filters.getByLabel('Start date').fill('2026-07-01');
  await filters.getByLabel('End date').fill('2026-12-31');
  await expect(page).toHaveURL(/from=2026-07-01&to=2026-12-31/);
  await expect(filters.getByRole('button', { name: 'Reset filters' })).toBeVisible();
  await expect(widget(page, 'Plan budgets')).toContainText('Not filtered by date');

  await page
    .getByRole('navigation', { name: 'Breadcrumb' })
    .getByRole('link', { name: 'Dashboards' })
    .click();
  await expect(page.getByRole('heading', { level: 1, name: 'Dashboards' })).toBeVisible();
  await page.goBack();
  await expect(page).toHaveURL(/from=2026-07-01&to=2026-12-31/);
  await expect(filters.getByLabel('Start date')).toHaveValue('2026-07-01');

  await filters.getByRole('button', { name: 'Reset filters' }).click();
  await expect(page).not.toHaveURL(/from=/);
  await expect(filters.getByLabel('Start date')).toHaveValue('');
});

test('051 US2-9, FR-005: a dashboard not shared with Mei answers "Dashboard not found"', async ({
  page,
}) => {
  await signInAs(page, DANIEL);
  await page.goto('/dashboards');
  await page.getByRole('link', { name: DASHBOARD }).click();
  await expect(page).toHaveURL(/\/dashboards\/DB-\d+$/);
  const address = new URL(page.url()).pathname;

  await signInAs(page, MEI);
  await page.goto(address);
  await expect(page.getByText('Dashboard not found')).toBeVisible();
  await expect(page.getByText(DASHBOARD)).toHaveCount(0);
  await expect(page.getByRole('link', { name: 'Go to Dashboards' })).toBeVisible();
  await expectNoAxeViolations(page);
});
