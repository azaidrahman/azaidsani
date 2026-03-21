# Swup Client-Side Navigation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add swup client-side navigation so page changes swap content without full reloads, while keeping the sidebar (activity calendar, social links) persistent.

**Architecture:** Dual-container swap — `#swup-body` for main content and `#swup-aside` for the dynamic aside block. Swup core + Head Plugin + Scripts Plugin loaded via CDN. A small init script configures containers, body class toggling, and scroll-to-top behavior.

**Tech Stack:** swup 4.x (CDN), Hugo templates, Playwright for testing

**Spec:** `docs/superpowers/specs/2026-03-22-swup-client-side-nav-design.md`

---

### Task 1: Add swap container IDs to baseof.html

**Files:**
- Modify: `layouts/_default/baseof.html:15` (add `id="swup-body"`)
- Modify: `layouts/_default/baseof.html:24` (add `id="swup-aside"`)

- [ ] **Step 1: Add `id="swup-body"` to the page body section**

In `layouts/_default/baseof.html`, change line 15 from:
```html
<section class="page__body">
```
to:
```html
<section class="page__body" id="swup-body">
```

- [ ] **Step 2: Add `id="swup-aside"` to the aside content div**

In `layouts/_default/baseof.html`, change line 24 from:
```html
<div class="aside__content">
```
to:
```html
<div class="aside__content" id="swup-aside">
```

**IMPORTANT:** The `id="swup-aside"` goes on `aside__content`, NOT on `aside__about`. The about section (calendar, social links, email button) must remain outside the swap boundary.

- [ ] **Step 3: Commit**

```bash
git add layouts/_default/baseof.html
git commit -m "feat: add swup swap container IDs to base layout"
```

---

### Task 2: Add swup CDN scripts and create init script

**Files:**
- Modify: `layouts/partials/head.html` (add CDN scripts + init script reference)
- Create: `static/js/swup-init.js`

- [ ] **Step 1: Add CDN scripts to head.html**

Add the following at the end of `layouts/partials/head.html` (after the OG meta block, before the closing of the file):

```html
<!-- swup: client-side navigation -->
<script defer src="https://unpkg.com/swup@4"></script>
<script defer src="https://unpkg.com/@swup/head-plugin@2"></script>
<script defer src="https://unpkg.com/@swup/scripts-plugin@2"></script>
<script defer src="{{ "js/swup-init.js" | absURL }}"></script>
```

All scripts use `defer` so they load in order and execute after DOM parsing.

- [ ] **Step 2: Create `static/js/swup-init.js`**

Note: The persistent zone (`aside__about`) contains three scripts that must NOT be re-executed by the Scripts Plugin:
- Email copy button inline script (`about.html:25-36`)
- `window.acalData` inline script (`about.html:54-63`)
- `activity-cal.js` script tag (`about.html:64`)

All three are outside both swap containers (`#swup-body` and `#swup-aside`), so the Scripts Plugin will not touch them. Their event listeners and DOM state persist across navigations.

```javascript
document.addEventListener('DOMContentLoaded', function () {
  if (typeof Swup === 'undefined') {
    console.warn('swup: library not loaded, falling back to standard navigation');
    return;
  }

  var swup = new Swup({
    containers: ['#swup-body', '#swup-aside'],
    plugins: [
      new SwupHeadPlugin(),
      new SwupScriptsPlugin()
    ]
  });

  // Toggle body class for homepage vs non-homepage styling
  function updateBodyClass() {
    var isHome = window.location.pathname === '/' || window.location.pathname === '/index.html';
    if (isHome) {
      document.body.classList.remove('not-home');
    } else {
      document.body.classList.add('not-home');
    }
  }

  // Scroll to top and update body class on forward navigation
  swup.hooks.on('page:view', function () {
    updateBodyClass();
    window.scrollTo(0, 0);
  });
});
```

- [ ] **Step 3: Verify Hugo builds without errors**

Run: `cd /Users/abdullahzaidas-sani/conductor/workspaces/website/davis && hugo --gc --minify`
Expected: Build succeeds with no errors

- [ ] **Step 4: Commit**

```bash
git add layouts/partials/head.html static/js/swup-init.js
git commit -m "feat: add swup client-side navigation with head and scripts plugins"
```

---

### Task 3: Fix resume-viewer.js keydown listener leak

**Files:**
- Modify: `static/js/resume-viewer.js`

When swup's Scripts Plugin re-executes `resume-viewer.js` on repeated visits to `/resume`, the IIFE creates duplicate `document` keydown listeners with stale closure references. Fix by using a named function stored on `window` so the old listener can be removed before adding a new one.

- [ ] **Step 1: Refactor resume-viewer.js to clean up previous keydown listener**

Replace the entire contents of `static/js/resume-viewer.js` with:

```javascript
(function () {
  var overlay = document.getElementById("resume-popup");
  var openBtn = document.getElementById("download-btn");
  var closeBtn = document.getElementById("popup-close");

  function openPopup() {
    overlay.style.display = "flex";
  }

  function closePopup() {
    overlay.style.display = "none";
  }

  openBtn.addEventListener("click", openPopup);
  closeBtn.addEventListener("click", closePopup);

  overlay.addEventListener("click", function (e) {
    if (e.target === overlay) {
      closePopup();
    }
  });

  // --- Click-to-zoom (desktop only) ---
  var zoomOverlay = document.getElementById("resume-zoom");
  var zoomImg = document.getElementById("resume-zoom-img");
  var isDesktop = window.matchMedia("(hover: hover) and (pointer: fine)").matches;

  function openZoom(src, alt) {
    zoomImg.src = src;
    zoomImg.alt = alt;
    zoomOverlay.classList.add("active");
    document.body.style.overflow = "hidden";
    zoomImg.onload = function () {
      zoomOverlay.scrollLeft = (zoomOverlay.scrollWidth - zoomOverlay.clientWidth) / 2;
      zoomOverlay.scrollTop = 0;
    };
  }

  function closeZoom() {
    zoomOverlay.classList.remove("active");
    zoomImg.src = "";
    document.body.style.overflow = "";
  }

  if (isDesktop) {
    var pageImages = document.querySelectorAll(".resume-page-img");
    for (var i = 0; i < pageImages.length; i++) {
      pageImages[i].addEventListener("click", function () {
        openZoom(this.src, this.alt);
      });
    }
  }

  // Close on click, but not if user was scrolling/dragging
  var didDrag = false;
  zoomOverlay.addEventListener("mousedown", function () { didDrag = false; });
  zoomOverlay.addEventListener("mousemove", function () { didDrag = true; });
  zoomOverlay.addEventListener("mouseup", function (e) {
    if (!didDrag) closeZoom();
  });

  // Remove previous keydown handler if it exists (prevents leak on swup re-execution)
  if (window._resumeKeydown) {
    document.removeEventListener("keydown", window._resumeKeydown);
  }

  window._resumeKeydown = function (e) {
    if (e.key === "Escape") {
      if (zoomOverlay.classList.contains("active")) {
        closeZoom();
      } else if (overlay.style.display === "flex") {
        closePopup();
      }
    }
  };

  document.addEventListener("keydown", window._resumeKeydown);
})();
```

The only change: the anonymous keydown handler is now stored as `window._resumeKeydown` and any previous instance is removed before adding the new one.

- [ ] **Step 2: Verify resume page still works on direct load**

Run: `cd /Users/abdullahzaidas-sani/conductor/workspaces/website/davis && hugo server -D --port 1313 &`
Open `http://localhost:1313/resume/` in a browser. Verify Download button popup and Escape key work.

- [ ] **Step 3: Commit**

```bash
git add static/js/resume-viewer.js
git commit -m "fix: prevent keydown listener leak on swup re-execution of resume-viewer"
```

---

### Task 4: Write Playwright tests for client-side navigation

**Files:**
- Create: `tests/navigation.spec.ts`

- [ ] **Step 1: Write the navigation test file**

Create `tests/navigation.spec.ts`:

```typescript
import { test, expect } from '@playwright/test';

test.describe('Client-Side Navigation (swup)', () => {
  test('navigating between pages does not trigger full reload', async ({ page }) => {
    await page.goto('/');

    // Mark the window to detect if a full reload happens
    await page.evaluate(() => {
      (window as any).__swupLoaded = true;
    });

    // Navigate to posts via sidebar/header link
    await page.click('a[href="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();

    // If swup worked, our window marker should still be present
    const markerSurvived = await page.evaluate(() => (window as any).__swupLoaded === true);
    expect(markerSurvived).toBe(true);
  });

  test('page title updates on navigation', async ({ page }) => {
    await page.goto('/');
    const homeTitle = await page.title();

    await page.click('a[href="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();
    const postsTitle = await page.title();

    expect(postsTitle).not.toBe(homeTitle);
  });

  test('body class toggles between home and non-home', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('body:not(.not-home)')).toBeAttached();

    await page.click('a[href="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();
    await expect(page.locator('body.not-home')).toBeAttached();

    // Navigate back to home
    await page.click('a[href="/"]');
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
    await page.click('a[href="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();

    // Calendar should still be visible and unchanged (persistent zone)
    await expect(calGrid).toBeVisible();
    const afterNavLabel = await page.locator('#acal-label').textContent();
    expect(afterNavLabel).toBe(initialLabel);
  });

  test('resume viewer works after client-side navigation', async ({ page }) => {
    await page.goto('/');

    // Navigate to resume via swup
    await page.click('a[href="/resume/"]');
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
    await page.click('a[href="/resume/"]');
    await expect(page.locator('.resume-viewer')).toBeVisible();

    // Navigate away
    await page.click('a[href="/"]');
    await expect(page.locator('.recent-posts')).toBeVisible();

    // Second visit (Scripts Plugin re-executes resume-viewer.js)
    await page.click('a[href="/resume/"]');
    await expect(page.locator('.resume-viewer')).toBeVisible();

    // Verify popup still works (would fail with stale listeners)
    await page.click('#download-btn');
    await expect(page.locator('#resume-popup')).toBeVisible();
    await page.keyboard.press('Escape');
    await expect(page.locator('#resume-popup')).not.toBeVisible();
  });

  test('email copy button works after navigation', async ({ page }) => {
    await page.goto('/');
    const copyBtn = page.locator('.email-copy-btn');
    await expect(copyBtn).toBeVisible();

    // Navigate away and back
    await page.click('a[href="/posts/"]');
    await expect(page.locator('.post-header__title')).toBeVisible();
    await page.click('a[href="/"]');
    await expect(page.locator('.recent-posts')).toBeVisible();

    // Email copy button should still be functional (persistent zone)
    await expect(copyBtn).toBeVisible();
    await copyBtn.click();
    // Button should show "copied" state
    await expect(copyBtn).toHaveClass(/copied/);
  });

  test('browser back/forward works', async ({ page }) => {
    await page.goto('/');

    await page.click('a[href="/posts/"]');
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
    await page.click('a[href="/"]');
    await expect(page.locator('.recent-posts')).toBeVisible();

    const homeAsideText = await asideContent.textContent();
    // Aside content should have changed (post had content, home is empty)
    expect(homeAsideText?.trim()).not.toBe(postAsideText?.trim());
  });
});
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `cd /Users/abdullahzaidas-sani/conductor/workspaces/website/davis && npx playwright test tests/navigation.spec.ts --project=desktop`
Expected: All tests pass

- [ ] **Step 3: Run the full test suite to check for regressions**

Run: `cd /Users/abdullahzaidas-sani/conductor/workspaces/website/davis && npx playwright test --project=desktop`
Expected: All existing tests still pass

- [ ] **Step 4: Commit**

```bash
git add tests/navigation.spec.ts
git commit -m "test: add Playwright tests for swup client-side navigation"
```

---

### Task 5: Manual smoke test and final verification

- [ ] **Step 1: Start Hugo dev server**

Run: `cd /Users/abdullahzaidas-sani/conductor/workspaces/website/davis && hugo server -D --port 1313`

- [ ] **Step 2: Verify these manually in a browser**

1. Navigate Home -> Posts -> single post -> Resume -> Home (no full reloads, content swaps smoothly)
2. Activity calendar remains interactive throughout (click prev/next months)
3. Email copy button works after navigating away and back to home
4. Resume Download popup opens and Escape closes it (after navigating to /resume via link)
5. Browser back/forward navigates correctly
6. External links (GitHub, LinkedIn, Letterboxd) open normally in new context
7. Page title in browser tab updates on each navigation

- [ ] **Step 3: Run full Playwright suite across all viewports**

Run: `cd /Users/abdullahzaidas-sani/conductor/workspaces/website/davis && npx playwright test`
Expected: All tests pass across desktop, tablet, and mobile projects
