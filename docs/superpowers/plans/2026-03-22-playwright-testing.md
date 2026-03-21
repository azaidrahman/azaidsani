# Playwright Testing Setup Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Playwright E2E tests covering responsiveness, content, calendar, and resume features, with a pre-push hook for fast smoke tests and GitHub Actions for the full suite.

**Architecture:** Playwright runs against a local Hugo dev server started automatically via `webServer` config. Three viewport projects (mobile/tablet/desktop) each run all spec files. A Husky pre-push hook runs a fast smoke subset locally; GitHub Actions runs the full suite on every push/PR.

**Tech Stack:** Playwright Test, Husky, GitHub Actions, Hugo

**Spec:** `docs/superpowers/specs/2026-03-22-playwright-testing-design.md`

---

### Task 1: Project scaffolding — package.json, Playwright config, .gitignore

**Files:**
- Create: `package.json`
- Create: `playwright.config.ts`
- Modify: `.gitignore`

- [ ] **Step 1: Create package.json**

```json
{
  "private": true,
  "scripts": {
    "test": "npx playwright test",
    "test:smoke": "npx playwright test --project=desktop --grep @smoke",
    "prepare": "husky"
  },
  "devDependencies": {
    "@playwright/test": "^1.50",
    "husky": "^9"
  }
}
```

- [ ] **Step 2: Create playwright.config.ts**

```ts
import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? 'html' : 'list',
  use: {
    baseURL: 'http://localhost:1313',
    browserName: 'chromium',
  },
  projects: [
    {
      name: 'mobile',
      use: { viewport: { width: 375, height: 667 } },
    },
    {
      name: 'tablet',
      use: { viewport: { width: 768, height: 1024 } },
    },
    {
      name: 'desktop',
      use: { viewport: { width: 1280, height: 720 } },
    },
  ],
  webServer: {
    command: 'hugo server -D --port 1313',
    url: 'http://localhost:1313',
    timeout: 30_000,
    reuseExistingServer: !process.env.CI,
  },
});
```

- [ ] **Step 3: Append to .gitignore**

Add these lines to the end of `.gitignore`:

```
node_modules/
test-results/
playwright-report/
```

- [ ] **Step 4: Install dependencies**

Run: `npm install`

Then install Chromium browser:

Run: `npx playwright install chromium`

- [ ] **Step 5: Verify Playwright runs (no tests yet)**

Run: `npx playwright test`

Expected: 0 tests found, exits cleanly (no errors)

- [ ] **Step 6: Commit**

```bash
git add package.json playwright.config.ts .gitignore package-lock.json
git commit -m "chore: scaffold Playwright testing environment"
```

---

### Task 2: Hugo build test

**Files:**
- Create: `tests/hugo-build.spec.ts`

- [ ] **Step 1: Write the test**

```ts
import { test, expect } from '@playwright/test';
import { execFileSync } from 'child_process';
import { existsSync } from 'fs';

test.describe('Hugo Build', () => {
  test('hugo --minify exits successfully @smoke', () => {
    const result = execFileSync('hugo', ['--minify'], { encoding: 'utf-8', timeout: 30_000 });
    expect(result).toBeDefined();
  });

  test('public/index.html exists after build @smoke', () => {
    execFileSync('hugo', ['--minify'], { timeout: 30_000 });
    expect(existsSync('public/index.html')).toBe(true);
  });
});
```

- [ ] **Step 2: Run the test**

Run: `npx playwright test tests/hugo-build.spec.ts --project=desktop`

Expected: 2 tests PASS

- [ ] **Step 3: Commit**

```bash
git add tests/hugo-build.spec.ts
git commit -m "test: add Hugo build verification tests"
```

---

### Task 3: Content integrity tests

**Files:**
- Create: `tests/content.spec.ts`

These tests verify all pages load, recent posts show up, links work, images resolve, and social links are present.

- [ ] **Step 1: Write the test**

```ts
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
    const entries = page.locator('.page__body ul li a');
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
```

- [ ] **Step 2: Run the test**

Run: `npx playwright test tests/content.spec.ts --project=desktop`

Expected: All tests PASS

- [ ] **Step 3: Commit**

```bash
git add tests/content.spec.ts
git commit -m "test: add content integrity tests"
```

---

### Task 4: Responsiveness tests

**Files:**
- Create: `tests/responsiveness.spec.ts`

These tests verify layout behavior across viewports. The mobile sidebar rule fires at `max-width: 45rem` (720px). The mobile project at 375px is well within that range.

- [ ] **Step 1: Write the test**

```ts
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
```

- [ ] **Step 2: Run the tests across all viewports**

Run: `npx playwright test tests/responsiveness.spec.ts`

Expected: Mobile-only tests run in `mobile` project, desktop/tablet tests run in `tablet` and `desktop` projects. Some tests skipped in non-applicable viewports. All running tests PASS.

- [ ] **Step 3: Commit**

```bash
git add tests/responsiveness.spec.ts
git commit -m "test: add responsiveness tests across viewports"
```

---

### Task 5: Activity calendar tests

**Files:**
- Create: `tests/calendar.spec.ts`

- [ ] **Step 1: Write the test**

```ts
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
```

- [ ] **Step 2: Run the tests**

Run: `npx playwright test tests/calendar.spec.ts --project=desktop`

Expected: All 5 tests PASS

- [ ] **Step 3: Commit**

```bash
git add tests/calendar.spec.ts
git commit -m "test: add activity calendar tests"
```

---

### Task 6: Resume page tests

**Files:**
- Create: `tests/resume.spec.ts`

- [ ] **Step 1: Write the test**

```ts
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
```

- [ ] **Step 2: Run the tests**

Run: `npx playwright test tests/resume.spec.ts --project=desktop`

Expected: All 3 tests PASS

- [ ] **Step 3: Commit**

```bash
git add tests/resume.spec.ts
git commit -m "test: add resume page tests"
```

---

### Task 7: Pre-push hook with Husky

**Files:**
- Create: `.husky/pre-push`

- [ ] **Step 1: Initialize Husky**

Run: `npx husky init`

This creates `.husky/` directory and a default `pre-commit` hook.

- [ ] **Step 2: Remove default pre-commit hook**

Run: `rm .husky/pre-commit` (we only want pre-push)

- [ ] **Step 3: Create .husky/pre-push**

```bash
#!/usr/bin/env sh

echo "Running pre-push checks..."

# 1. Hugo build check
echo "Checking Hugo build..."
hugo --minify || exit 1

# 2. Fast smoke tests (desktop only, @smoke tagged)
echo "Running smoke tests..."
npx playwright test --project=desktop --grep @smoke || exit 1

echo "Pre-push checks passed."
```

- [ ] **Step 4: Make it executable**

Run: `chmod +x .husky/pre-push`

- [ ] **Step 5: Verify the hook runs**

Run: `npm run test:smoke`

Expected: Runs only `@smoke` tagged tests in desktop project. All pass.

- [ ] **Step 6: Commit**

```bash
git add .husky/pre-push
git commit -m "chore: add pre-push hook for smoke tests"
```

---

### Task 8: GitHub Actions workflow

**Files:**
- Create: `.github/workflows/test.yml`

- [ ] **Step 1: Create the workflow file**

```yaml
name: Tests

on:
  push:
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive

      - name: Setup Hugo
        uses: peaceiris/actions-hugo@v3
        with:
          hugo-version: 'latest'
          extended: true

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      - name: Install Playwright browsers
        run: npx playwright install --with-deps chromium

      - name: Run Playwright tests
        run: npx playwright test

      - name: Upload test report
        uses: actions/upload-artifact@v4
        if: failure()
        with:
          name: playwright-report
          path: playwright-report/
          retention-days: 7
```

- [ ] **Step 2: Validate YAML syntax**

Run: `python3 -c "import yaml; yaml.safe_load(open('.github/workflows/test.yml'))"`

Expected: No output (valid YAML)

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/test.yml
git commit -m "ci: add GitHub Actions workflow for Playwright tests"
```

---

### Task 9: Full suite verification

Run the entire test suite across all viewports to make sure everything works together.

- [ ] **Step 1: Run full suite**

Run: `npx playwright test`

Expected: All tests pass across mobile, tablet, and desktop projects. Some responsiveness tests are skipped in non-applicable viewports (this is expected).

- [ ] **Step 2: Run smoke subset**

Run: `npm run test:smoke`

Expected: Only `@smoke` tagged tests run in desktop project. All pass. Should complete in under 10 seconds.

- [ ] **Step 3: Clean up public/ directory**

Run: `rm -rf public/`

(Hugo build test creates this; it's already in .gitignore)

- [ ] **Step 4: Final commit if any adjustments were needed**

Only commit if fixes were required during verification.
