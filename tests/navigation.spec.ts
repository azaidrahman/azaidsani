import { test, expect } from '@playwright/test';

test.describe('Client-Side Navigation (swup)', () => {
  test('navigating between pages does not trigger full reload', async ({ page }) => {
    await page.goto('/');

    // Mark the window to detect if a full reload happens
    await page.evaluate(() => {
      (window as any).__swupLoaded = true;
    });

    // Navigate to posts via sidebar/header link
    await page.click('a[href$="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();

    // If swup worked, our window marker should still be present
    const markerSurvived = await page.evaluate(() => (window as any).__swupLoaded === true);
    expect(markerSurvived).toBe(true);
  });

  test('page title updates on navigation', async ({ page }) => {
    await page.goto('/');
    const homeTitle = await page.title();

    await page.click('a[href$="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();
    // Wait for the title to update (head plugin may update asynchronously)
    await expect.poll(() => page.title()).not.toBe(homeTitle);
    const postsTitle = await page.title();

    expect(postsTitle).not.toBe(homeTitle);
  });

  test('body class toggles between home and non-home', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('body:not(.not-home)')).toBeAttached();

    await page.click('a[href$="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();
    await expect(page.locator('body.not-home')).toBeAttached();

    // Navigate back to home
    await page.click('a.page__logo-inner');
    await expect(page.locator('.recent-posts')).toBeVisible();
    await expect(page.locator('body:not(.not-home)')).toBeAttached();
  });

  test('activity calendar persists across navigations', async ({ page }) => {
    await page.goto('/');
    const calGrid = page.locator('#acal-grid');
    await expect(calGrid).toBeVisible();

    // Get initial calendar state
    const initialLabel = await page.locator('#acal-label').textContent();

    // Navigate away and back
    await page.click('a[href$="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();

    // Calendar should still be in DOM and unchanged (persistent zone)
    // Note: on mobile, sidebar is hidden on non-home pages via CSS, so check attached not visible
    await expect(calGrid).toBeAttached();
    const afterNavLabel = await page.locator('#acal-label').textContent();
    expect(afterNavLabel).toBe(initialLabel);
  });

  test('resume viewer works after client-side navigation', async ({ page }) => {
    await page.goto('/');

    // Navigate to resume via swup
    await page.click('a[href$="/resume/"]');
    await expect(page.locator('.resume-viewer')).toBeVisible();

    // Verify resume images loaded
    await expect(page.locator('.resume-page-img')).toHaveCount(2);

    // Verify download popup works
    await page.click('#download-btn');
    await expect(page.locator('#resume-popup')).toBeVisible();

    // Close with Escape
    await page.keyboard.press('Escape');
    await expect(page.locator('#resume-popup')).not.toBeVisible();
  });

  test('resume viewer works on repeated visits without duplicate listeners', async ({ page }) => {
    await page.goto('/');

    // First visit to resume
    await page.click('a[href$="/resume/"]');
    await expect(page.locator('.resume-viewer')).toBeVisible();

    // Navigate away
    await page.click('a.page__logo-inner');
    await expect(page.locator('.recent-posts')).toBeVisible();

    // Second visit (Scripts Plugin re-executes resume-viewer.js)
    await page.click('a[href$="/resume/"]');
    await expect(page.locator('.resume-viewer')).toBeVisible();

    // Verify popup still works (would fail with stale listeners)
    await page.click('#download-btn');
    await expect(page.locator('#resume-popup')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('#resume-popup')).not.toBeVisible();
  });

  test('email copy button works after navigation', async ({ page, context }) => {
    // Grant clipboard permissions so the copy handler can resolve
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);

    await page.goto('/');
    const copyBtn = page.locator('.email-copy-btn');
    await expect(copyBtn).toBeVisible();

    // Navigate away and back
    await page.click('a[href$="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();
    await page.click('a.page__logo-inner');
    await expect(page.locator('.recent-posts')).toBeVisible();

    // Email copy button should still be functional (persistent zone)
    await expect(copyBtn).toBeVisible();
    await copyBtn.click();
    // Button should show "copied" state
    await expect(copyBtn).toHaveClass(/copied/);
  });

  test('browser back/forward works', async ({ page }) => {
    await page.goto('/');

    await page.click('a[href$="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();

    await page.goBack();
    await expect(page.locator('.recent-posts')).toBeVisible();

    await page.goForward();
    await expect(page.locator('.post-header__title')).toBeVisible();
  });

  test('aside content updates on navigation', async ({ page }) => {
    // Navigate to a post with TOC/aside content
    await page.goto('/posts/how-i-built-this-website/');
    const asideContent = page.locator('#swup-aside');

    // Store whether aside has content on a post page
    const postAsideText = await asideContent.textContent();

    // Navigate to homepage (aside block is empty)
    await page.click('a.page__logo-inner');
    await expect(page.locator('.recent-posts')).toBeVisible();

    const homeAsideText = await asideContent.textContent();
    // Aside content should have changed (post had content, home is empty)
    expect(homeAsideText?.trim()).not.toBe(postAsideText?.trim());
  });
});
