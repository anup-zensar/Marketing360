import type { Locator, Page } from '@playwright/test';
import { expectNoAxeViolations, expectToast, signInAs, uniqueName } from './fixtures';
import { createPlanningRecords } from './planning';
import { expect, test } from './support/no-csp-violations';

// Feature 051 US1 against the real API: Priya builds, checks and saves a dashboard; a widget's figure equals
// the activity page; leaving with unsaved changes asks first; a Viewer gets no builder.

const PRIYA = 'priya.raman@shell.com';
const MEI = 'mei.lin@shell.com';
const SARA = 'sara.novak@shell.com';

function cell(page: Page, name: string): Locator {
  return page.getByRole('group', { name: new RegExp(`^Widget: ${name}, column \\d+, row \\d+$`) });
}

function inspector(page: Page): Locator {
  return page.getByRole('complementary', { name: 'Inspector' });
}

async function chooseData(page: Page, source: string, groupBy: string, measure: string) {
  const panel = inspector(page);
  await panel.getByLabel(/^Source/).selectOption({ label: source });
  await panel.getByLabel(/^Group by/).selectOption({ label: groupBy });
  await panel.getByRole('button', { name: 'Add measure' }).click();
  await panel.getByLabel('What to measure').selectOption({ label: measure });
}

for (const colorScheme of ['light', 'dark'] as const) {
  test(`051 SC-006: the builder with a saved dashboard passes axe (${colorScheme})`, async ({
    page,
  }) => {
    await page.emulateMedia({ colorScheme });
    await signInAs(page, PRIYA);
    await page.goto('/builder/DB-1001');
    await expect(
      page.getByRole('heading', { level: 1, name: "Edit 'Lubricants Europe overview'" }),
    ).toBeVisible();
    await expect(page.locator('html')).toHaveAttribute('data-theme', colorScheme);
    await expect(cell(page, 'Planned cost by activity type').getByRole('img')).toBeVisible();
    await expectNoAxeViolations(page);

    await cell(page, 'Planned cost by activity type')
      .getByRole('button', { name: 'Planned cost by activity type', exact: true })
      .click();
    await expect(inspector(page).getByRole('tab', { name: 'Data' })).toBeVisible();
    await expectNoAxeViolations(page);
  });
}

test('051 US1: Priya builds a dashboard — palette, Data tab, keyboard, undo, problems, save, reload (FR-026 – FR-035)', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await page.goto('/');
  await page.getByRole('link', { name: 'Dashboard builder' }).click();
  await expect(page.getByRole('heading', { level: 1, name: 'New dashboard' })).toBeVisible();
  await expect(page.getByRole('button', { name: 'Save' })).toBeVisible();
  await expectNoAxeViolations(page);

  await page.getByRole('button', { name: 'Add bar' }).click();
  const bar = cell(page, 'Bar chart');
  await expect(bar).toHaveAccessibleName('Widget: Bar chart, column 1, row 1');
  await expect(inspector(page).getByRole('tab', { name: 'Data' })).toHaveAttribute(
    'aria-selected',
    'true',
  );
  await chooseData(page, 'Activities', 'Activity type', 'Planned cost');
  await expect(bar.getByRole('img', { name: /^Bar chart\./ })).toBeVisible();

  const handle = bar.getByRole('button', { name: 'Bar chart', exact: true });
  await handle.focus();
  await page.keyboard.press('ArrowRight');
  await expect(bar).toHaveAccessibleName('Widget: Bar chart, column 2, row 1');
  await expect(page.getByText('Moved to column 2, row 1.')).toBeAttached();
  await page.keyboard.press('Shift+ArrowRight');
  await expect(bar).toHaveAttribute('style', /grid-column: 2 \/ span 7/);
  await page.getByRole('button', { name: 'Undo' }).click();
  await expect(bar).toHaveAttribute('style', /grid-column: 2 \/ span 6/);

  await page
    .getByRole('button', { name: 'Add table' })
    .dragTo(page.getByRole('region', { name: 'Canvas' }), { targetPosition: { x: 40, y: 440 } });
  const table = cell(page, 'Table');
  await expect(table).toBeVisible();
  await expect(table.getByText('Choose what to measure')).toBeVisible();

  await page.getByRole('button', { name: 'Save' }).click();
  const summary = page.getByRole('alert').filter({ hasText: 'problems need your attention' });
  await expect(summary).toContainText('2 problems need your attention');
  await expect(summary.getByRole('link', { name: 'Enter a name.' })).toBeVisible();
  await expect(summary.getByRole('link', { name: 'Table: Choose a source.' })).toBeVisible();
  await expect(table.getByText('Choose a source.')).toBeVisible();
  await expect(page).toHaveURL(/\/builder$/);
  await expectNoAxeViolations(page);

  await summary.getByRole('link', { name: 'Table: Choose a source.' }).click();
  await chooseData(page, 'Plans', 'Plan', 'Planned cost');
  const name = uniqueName('Lubricants check');
  await page.getByRole('button', { name: 'Dashboard details' }).click();
  await inspector(page).getByLabel(/^Name/).fill(name);
  await page.getByRole('button', { name: 'Save' }).click();

  await expectToast(page, 'Dashboard saved');
  await expect(page).toHaveURL(/\/builder\/DB-\d+$/);
  await page.reload();
  await expect(page.getByRole('heading', { level: 1, name: `Edit '${name}'` })).toBeVisible();
  await expect(cell(page, 'Bar chart')).toHaveAccessibleName('Widget: Bar chart, column 2, row 1');
  await expect(
    cell(page, 'Table')
      .getByRole('region', { name: 'Values, table' })
      .getByRole('rowheader', { name: 'FY2026 Lubricants Europe' }),
  ).toBeVisible();

  await page.getByRole('link', { name: 'View', exact: true }).click();
  await expect(page.getByRole('heading', { level: 1, name })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Edit' })).toBeVisible();
});

test('051 US1-10, FR-034: leaving the builder with unsaved changes asks "Discard changes?"', async ({
  page,
}) => {
  await signInAs(page, PRIYA);
  await page.goto('/builder');
  await page.getByRole('button', { name: 'Add text note' }).click();
  await page.getByRole('link', { name: 'Dashboards', exact: true }).first().click();

  const dialog = page.getByRole('alertdialog', { name: 'Discard changes?' });
  await expect(dialog.getByRole('button', { name: 'Keep editing' })).toBeFocused();
  await dialog.getByRole('button', { name: 'Keep editing' }).click();
  await expect(page).toHaveURL(/\/builder$/);
  await expect(cell(page, 'Text note')).toBeVisible();

  await page.getByRole('link', { name: 'Dashboards', exact: true }).first().click();
  await dialog.getByRole('button', { name: 'Discard changes' }).click();
  await expect(page.getByRole('heading', { level: 1, name: 'Dashboards' })).toBeVisible();
});

test('051 US1-3, SC-001: a widget adds up the planned cost the activity page shows', async ({
  page,
}) => {
  await signInAs(page, MEI);
  const records = await createPlanningRecords(page, 'Builder figures', {
    withWorkflow: false,
    leaveEmpty: [],
  });
  await page.goto('/builder');
  await page.getByRole('button', { name: 'Add table' }).click();
  await chooseData(page, 'Activities', 'Activity', 'Planned cost');
  const name = uniqueName('Mobility check');
  await page.getByRole('button', { name: 'Dashboard details' }).click();
  await inspector(page).getByLabel(/^Name/).fill(name);
  await page.getByRole('button', { name: 'Save' }).click();
  await expectToast(page, 'Dashboard saved');
  const dashboardId = /\/builder\/(DB-\d+)$/.exec(page.url())?.[1] ?? '';

  await page.goto(`/dashboards/${dashboardId}?plan=${records.planId}`);
  const row = page
    .getByRole('region', { name: 'Values, table' })
    .getByRole('row', { name: new RegExp(records.activityName) });
  await expect(row).toHaveCount(1);
  const shown = /[\d,]+\.\d{2}/.exec((await row.getByRole('cell').first().textContent()) ?? '');
  expect(shown?.[0]).toBe('20,000.00');

  await page.goto(records.activityPath);
  await expect(page.getByLabel('Summary')).toContainText(`${shown?.[0] ?? ''} USD`);
});

test('051 US1-12, FR-002: a Viewer has no Dashboard builder and is refused /builder', async ({
  page,
}) => {
  await signInAs(page, SARA);
  await page.goto('/');
  const sidebar = page.getByRole('navigation', { name: 'Main' });
  await expect(sidebar.getByRole('link', { name: 'Dashboards', exact: true })).toBeVisible();
  await expect(sidebar.getByRole('link', { name: 'Dashboard builder' })).toHaveCount(0);

  await page.goto('/builder');
  await expect(page.getByRole('heading', { level: 1, name: 'No access' })).toBeVisible();
  await expectNoAxeViolations(page);
});
