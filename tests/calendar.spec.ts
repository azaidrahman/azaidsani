import { test, expect } from '@playwright/test';

test.describe('Activity Calendar', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('calendar grid renders', async ({ page }) => {
    const grid = page.locator('#acal-grid');
    await expect(grid).toBeVisible();
  });

  test('has exactly 7 day-of-week headers', async ({ page }) => {
    const dowHeaders = page.locator('#acal-grid .acal-dow');
    await expect(dowHeaders).toHaveCount(7);
  });

  test('has day cells', async ({ page }) => {
    const cells = page.locator('#acal-grid .acal-cell');
    const count = await cells.count();
    // any month has at least 28 days + possibly empty padding cells
    expect(count).toBeGreaterThanOrEqual(28);
  });

  test('prev/next navigation changes month label', async ({ page }) => {
    const label = page.locator('#acal-label');
    const initialText = await label.textContent();
    expect(initialText).toBeTruthy();

    // click prev (should work since we allow 2 months back)
    await page.locator('#acal-prev').click();
    const afterPrev = await label.textContent();
    expect(afterPrev).not.toBe(initialText);

    // click next to go back to original
    await page.locator('#acal-next').click();
    const afterNext = await label.textContent();
    expect(afterNext).toBe(initialText);
  });

  test('legend is visible', async ({ page }) => {
    const legend = page.locator('.acal-legend');
    await expect(legend).toBeVisible();
    await expect(legend).toContainText('git');
    await expect(legend).toContainText('post');
  });
});
