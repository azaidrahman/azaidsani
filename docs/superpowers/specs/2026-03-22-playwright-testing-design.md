# Playwright Testing Setup for Personal Website

## Summary

Add a Playwright-based end-to-end testing environment to verify responsiveness, content integrity, JavaScript features, and Hugo build health. Tests run locally via a pre-push git hook (fast smoke subset) and in CI via GitHub Actions (full suite across 3 viewports).

## Stack

- **Playwright Test** (`@playwright/test`) — browser-based E2E testing
- **Husky** — git hook management for pre-push checks
- **GitHub Actions** — CI pipeline for full test suite
- **Hugo dev server** — Playwright's `webServer` config starts it automatically

## File Structure

```
package.json                    # Playwright + Husky, scripts
playwright.config.ts            # Viewports, server config, projects
tests/
  hugo-build.spec.ts            # Hugo compiles without errors
  content.spec.ts               # Pages load, links work, images render
  responsiveness.spec.ts        # Layout checks at mobile/tablet/desktop
  calendar.spec.ts              # Activity calendar renders, nav works
  animation.spec.ts             # Terminal animation elements present
  resume.spec.ts                # PDF embed loads on resume page
.github/
  workflows/
    test.yml                    # GitHub Actions workflow
.husky/
  pre-push                      # Git hook: hugo build + smoke tests
```

## Viewport Projects

| Project  | Width | Height | Device Equivalent |
|----------|-------|--------|-------------------|
| mobile   | 375   | 667    | iPhone SE         |
| tablet   | 768   | 1024   | iPad              |
| desktop  | 1280  | 720    | Standard laptop   |

## Test Specifications

### hugo-build.spec.ts

- Run `hugo --minify` and assert exit code 0
- Verify `public/index.html` exists after build

### content.spec.ts

- Homepage (`/`) loads with 200 status
- Posts page (`/posts/`) loads and lists blog entries
- Each blog post page loads with title and content present
- Resume page (`/resume/`) loads
- All `<img>` tags have valid `src` attributes that resolve (not 404)
- All internal `<a>` links return 200
- Social links (GitHub, LinkedIn, Letterboxd) are present in sidebar

### responsiveness.spec.ts

- **Mobile:** sidebar is hidden on non-homepage pages, content fills width
- **Tablet/Desktop:** sidebar is visible, layout is side-by-side
- Images respect `max-width: 100%` and don't overflow their container
- Activity calendar grid doesn't overflow on mobile

### calendar.spec.ts

- Calendar container renders on homepage
- Day cells are present (grid with 7-column structure)
- Previous/next month navigation buttons work (click and verify month label changes)
- Legend is visible with correct entries

### animation.spec.ts

- Terminal text element exists in sidebar
- Blinking cursor element is present

### resume.spec.ts

- Resume page contains an `<embed>` or `<iframe>` element
- The element's `src` points to the PDF file
- The PDF source URL does not 404

## Pre-Push Hook (Fast Subset)

Runs locally before every `git push`. Target: under 10 seconds.

1. `hugo --minify` — fail fast if build is broken
2. Playwright with `--project=desktop --grep @smoke` — runs only smoke-tagged tests

Smoke-tagged tests (`@smoke`):
- Hugo build succeeds
- Homepage loads
- No broken internal links on homepage

If either step fails, the push is blocked.

## GitHub Actions Workflow

**Triggers:** `push` (all branches), `pull_request` (to `main`)

**Steps:**
1. Checkout code (with submodules for Hugo theme)
2. Install Hugo (extended edition)
3. Install Node.js and dependencies (`npm ci`)
4. Install Playwright browsers (`npx playwright install --with-deps chromium`)
5. Run full Playwright suite across all 3 viewport projects
6. Upload Playwright HTML report as artifact on failure

## Playwright Config

- `webServer`: starts `hugo server -D --port 1313`, waits for `http://localhost:1313`
- Server startup timeout: 30 seconds
- Test timeout: 30 seconds per test
- Reporter: `list` locally, `html` in CI
- `retries`: 0 locally, 1 in CI
- Only Chromium (no need for cross-browser on a personal site)

## Dependencies

```json
{
  "devDependencies": {
    "@playwright/test": "^1.50",
    "husky": "^9"
  },
  "scripts": {
    "test": "npx playwright test",
    "test:smoke": "npx playwright test --project=desktop --grep @smoke",
    "prepare": "husky"
  }
}
```

## .gitignore Additions

```
node_modules/
test-results/
playwright-report/
```
