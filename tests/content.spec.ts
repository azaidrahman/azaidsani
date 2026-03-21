import { test, expect } from '@playwright/test';

test.describe('Content Integrity', () => {
  test('homepage loads with 200 @smoke', async ({ page }) => {
    const response = await page.goto('/');
    expect(response?.status()).toBe(200);
  });

  test('homepage has recent-posts section with up to 3 posts', async ({ page }) => {
    await page.goto('/');
    const recentPosts = page.locator('.recent-posts');
    await expect(recentPosts).toBeVisible();
    const items = recentPosts.locator('li');
    const count = await items.count();
    expect(count).toBeGreaterThan(0);
    expect(count).toBeLessThanOrEqual(3);
  });

  test('posts page loads and lists blog entries', async ({ page }) => {
    const response = await page.goto('/posts/');
    expect(response?.status()).toBe(200);
    const entries = page.locator('.post-item__title');
    const count = await entries.count();
    expect(count).toBeGreaterThan(0);
  });

  test('blog post: how-i-built-this-website loads', async ({ page }) => {
    const response = await page.goto('/posts/how-i-built-this-website/');
    expect(response?.status()).toBe(200);
    await expect(page.locator('.post-header__title')).toBeVisible();
    await expect(page.locator('.page__body')).toContainText('Hugo');
  });

  test('blog post: building-a-slack-bot loads', async ({ page }) => {
    const response = await page.goto('/posts/building-a-slack-bot/');
    expect(response?.status()).toBe(200);
    await expect(page.locator('.post-header__title')).toBeVisible();
    await expect(page.locator('.page__body')).toContainText('Slack');
  });

  test('blog post shortcodes render figure elements', async ({ page }) => {
    // building-a-slack-bot uses both movies and mid-img shortcodes
    await page.goto('/posts/building-a-slack-bot/');
    await expect(page.locator('figure.movies')).toBeVisible();
    await expect(page.locator('figure.mid-img')).toBeVisible();

    // how-i-built-this-website uses movies shortcode
    await page.goto('/posts/how-i-built-this-website/');
    await expect(page.locator('figure.movies')).toBeVisible();
  });

  test('resume page loads', async ({ page }) => {
    const response = await page.goto('/resume/');
    expect(response?.status()).toBe(200);
  });

  test('all images have valid src attributes', async ({ page }) => {
    await page.goto('/');
    const images = page.locator('img[src]');
    const count = await images.count();
    for (let i = 0; i < count; i++) {
      const src = await images.nth(i).getAttribute('src');
      expect(src).toBeTruthy();
      const url = new URL(src!, page.url());
      const resp = await page.request.get(url.toString());
      expect(resp.status(), `Image ${src} returned ${resp.status()}`).toBeLessThan(400);
    }
  });

  test('internal links return 200 @smoke', async ({ page }) => {
    await page.goto('/');
    const links = page.locator('a[href^="/"]');
    const hrefs = new Set<string>();
    const count = await links.count();
    for (let i = 0; i < count; i++) {
      const href = await links.nth(i).getAttribute('href');
      if (href) hrefs.add(href);
    }
    for (const href of hrefs) {
      const resp = await page.request.get(href);
      expect(resp.status(), `Link ${href} returned ${resp.status()}`).toBeLessThan(400);
    }
  });

  test('social links are present in sidebar', async ({ page }) => {
    await page.goto('/');
    const socialLinks = page.locator('.aside__social-links a');
    const texts = await socialLinks.allTextContents();
    const combined = texts.join(' ');
    expect(combined).toContain('GitHub');
    expect(combined).toContain('LinkedIn');
    expect(combined).toContain('Letterboxd');
  });
});
