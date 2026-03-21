import { test, expect } from '@playwright/test';

test.describe('Resume Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/resume/');
  });

  test('resume page has embed element', async ({ page }) => {
    const embed = page.locator('embed[type="application/pdf"]');
    await expect(embed).toBeAttached();
  });

  test('embed src points to PDF file', async ({ page }) => {
    const embed = page.locator('embed[type="application/pdf"]');
    const src = await embed.getAttribute('src');
    expect(src).toContain('Zaid-Resume.pdf');
  });

  test('PDF file is accessible (not 404)', async ({ page }) => {
    const resp = await page.request.get('/Zaid-Resume.pdf');
    expect(resp.status()).toBe(200);
  });
});
