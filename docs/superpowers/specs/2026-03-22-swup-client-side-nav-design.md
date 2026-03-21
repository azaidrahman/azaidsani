# Swup Client-Side Navigation

## Goal

Eliminate full page reloads when navigating between pages on the Hugo site by integrating swup for client-side content swapping.

## Architecture: Dual-Container Swap (Approach B)

Two swup containers divide the page into swappable and persistent zones:

```
<body>
  <div class="page">
    <header>                          <- persistent (nav)
    <section class="page__body"       <- SWAPPED (add id="swup-body" here)
             id="swup-body">
    <section class="page__aside">
      <div class="aside__about">      <- persistent (calendar, social, email)
      <hr>                            <- persistent
      <div class="aside__content"     <- SWAPPED (add id="swup-aside" here)
           id="swup-aside">
    <footer>                          <- persistent
  </div>
</body>
```

**Boundary warning:** `#swup-aside` must wrap only `aside__content`, NOT `aside__about`. Placing the swap boundary too high would destroy the activity calendar and email button listeners on every navigation.

### Why dual containers

- `page__body` content is always page-specific
- `aside__content` varies per page (TOC on posts, description on lists, empty on home/resume) — must update or stale content remains visible
- `aside__about` (activity calendar, social links, email copy button) is identical on every page and expensive to re-initialize — keeping it persistent avoids that cost entirely

## Plugins

| Plugin | Version | CDN | Purpose |
|--------|---------|-----|---------|
| swup (core) | 4.x | unpkg.com/swup@4 | Link interception, content swap, history |
| Head Plugin | 2.x | unpkg.com/@swup/head-plugin@2 | Updates `<title>`, meta tags, OG tags |
| Scripts Plugin | 2.x | unpkg.com/@swup/scripts-plugin@2 | Executes `<script>` tags in swapped content |

Note: Scripts Plugin executes scripts **after** DOM insertion of new content, so `getElementById` calls in page-specific scripts (like `resume-viewer.js`) will find their target elements.

## Scroll Management

On forward navigation, scroll to top of `#swup-body`. On browser back/forward, restore the previous scroll position. Handle this via swup's `page:view` hook with `window.scrollTo(0, 0)` for forward navigation. Swup's default popstate handling preserves scroll on history navigation.

## Body Class Handling

The `<body>` has class `not-home` on non-homepage routes. Since `<body>` is outside swap containers, a swup `page:view` hook will toggle this class based on `location.pathname === '/'`.

## Script Cleanup: resume-viewer.js

`resume-viewer.js` is an IIFE that adds a `document` keydown listener via closure. When the user navigates away from `/resume` and back, the Scripts Plugin re-executes the script, creating duplicate listeners pointing at stale DOM elements.

Fix: refactor `resume-viewer.js` to remove the previous keydown listener before adding a new one (use a named function reference stored on `window` or check for existing initialization).

## Installation

CDN `<script>` tags in `head.html` — no JS build step. The site has no existing bundler and adding one is out of scope.

## Files Changed

1. **`layouts/partials/head.html`** — Add swup + plugin CDN scripts and `<script src>` for swup-init.js
2. **`layouts/_default/baseof.html`** — Add `id="swup-body"` to `<section class="page__body">` and `id="swup-aside"` to `<div class="aside__content">`
3. **`static/js/swup-init.js`** (new) — Initialize swup with containers, plugins, body class hook, and scroll-to-top behavior
4. **`static/js/resume-viewer.js`** — Refactor to clean up previous keydown listener on re-initialization

## Behavior

1. User clicks any internal `<a>` link (external links and `target="_blank"` are ignored by swup by default)
2. Swup intercepts, fetches target page via AJAX
3. Head Plugin updates `<title>` and meta tags from the fetched page's `<head>`
4. Swup swaps content of `#swup-body` and `#swup-aside`
5. Scripts Plugin executes any `<script>` tags in the new content (e.g., resume-viewer.js)
6. `page:view` hook scrolls to top and toggles `not-home` body class
7. URL updates via History API; browser back/forward works with scroll restoration

## Graceful Degradation

If JavaScript fails to load, all `<a>` links still work as normal server-side navigation. No functionality is lost — only the smooth transition enhancement.

## Testing

- Verify navigation between all page types (home, post list, single post, resume) without reload
- Verify `<title>` updates correctly on each navigation
- Verify resume viewer works when navigating to /resume (not just on direct load)
- Verify resume viewer works correctly on repeated visits (no duplicate listeners)
- Verify activity calendar persists and remains interactive across navigations
- Verify browser back/forward works correctly with scroll restoration
- Verify body class toggles correctly (homepage vs non-homepage styling)
- Verify email copy button still works after navigating away and back
- Verify external links, `target="_blank"`, and anchor links are not intercepted
- Update Playwright tests per project conventions
