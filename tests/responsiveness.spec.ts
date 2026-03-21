import { test, expect } from '@playwright/test';

test.describe('Responsiveness', () => {
  test('mobile: sidebar hidden on non-homepage', async ({ page, viewport }) => {
    test.skip(viewport!.width > 720, 'Mobile-only test');
    await page.goto('/posts/');
    const aside = page.locator('.page__aside');
    await expect(aside).toBeHidden();
  });

  test('mobile: sidebar visible on homepage', async ({ page, viewport }) => {
    test.skip(viewport!.width > 720, 'Mobile-only test');
    await page.goto('/');
    const aside = page.locator('.page__aside');
    await expect(aside).toBeVisible();
  });

  test('desktop/tablet: sidebar visible on all pages', async ({ page, viewport }) => {
    test.skip(viewport!.width <= 720, 'Desktop/tablet-only test');
    for (const path of ['/', '/posts/', '/resume/']) {
      await page.goto(path);
      const aside = page.locator('.page__aside');
      await expect(aside).toBeVisible();
    }
  });

  test('images do not overflow container', async ({ page }) => {
    await page.goto('/posts/building-a-slack-bot/');
    const images = page.locator('.page__body img');
    const count = await images.count();
    for (let i = 0; i < count; i++) {
      const img = images.nth(i);
      const imgBox = await img.boundingBox();
      if (!imgBox) continue;
      const bodyBox = await page.locator('.page__body').boundingBox();
      expect(imgBox.width, `Image ${i} overflows container`).toBeLessThanOrEqual(bodyBox!.width + 1);
    }
  });

  test('mobile: activity calendar does not overflow', async ({ page, viewport }) => {
    test.skip(viewport!.width > 720, 'Mobile-only test');
    await page.goto('/');
    const grid = page.locator('#acal-grid');
    const gridBox = await grid.boundingBox();
    expect(gridBox).toBeTruthy();
    expect(gridBox!.width).toBeLessThanOrEqual(viewport!.width + 1);
  });
});
