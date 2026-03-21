import { test, expect } from '@playwright/test';

test.describe('Resume Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/resume/');
  });

  test('resume viewer is present', async ({ page }) => {
    const viewer = page.locator('.resume-viewer');
    await expect(viewer).toBeAttached();
  });

  test('resume page images are visible', async ({ page }) => {
    const images = page.locator('.resume-page-img');
    await expect(images).toHaveCount(2);
  });

  test('PDF file is accessible (not 404)', async ({ page }) => {
    const resp = await page.request.get('/Zaid-Resume.pdf');
    expect(resp.status()).toBe(200);
  });
});
