# Swup Client-Side Navigation

## Goal

Eliminate full page reloads when navigating between pages on the Hugo site by integrating swup for client-side content swapping.

## Architecture: Dual-Container Swap (Approach B)

Two swup containers divide the page into swappable and persistent zones:

```
<body>
  <div class="page">
    <header>          ← persistent (nav)
    <section #swup-body>   ← SWAPPED (main content)
    <section page__aside>
      <div aside__about>   ← persistent (calendar, social, email)
      <hr>
      <div #swup-aside>    ← SWAPPED (TOC, description, etc.)
    <footer>          ← persistent
  </div>
</body>
```

### Why dual containers

- `page__body` content is always page-specific
- `aside__content` varies per page (TOC on posts, description on lists, empty on home/resume) — must update or stale content remains visible
- `aside__about` (activity calendar, social links, email copy button) is identical on every page and expensive to re-initialize — keeping it persistent avoids that cost entirely

## Plugins

| Plugin | Purpose |
|--------|---------|
| Head Plugin | Updates `<title>`, meta tags, OG tags on each navigation |
| Scripts Plugin | Executes `<script>` tags found in swapped content (e.g., `resume-viewer.js` when navigating to /resume) |

## Body Class Handling

The `<body>` has class `not-home` on non-homepage routes. Since `<body>` is outside swap containers, a swup `page:view` hook will toggle this class based on the current URL.

## Installation

CDN `<script>` tags in `head.html` — no JS build step. The site has no existing bundler and adding one is out of scope.

## Files Changed

1. **`layouts/partials/head.html`** — Add swup + plugin CDN scripts
2. **`layouts/_default/baseof.html`** — Add `id="swup-body"` and `id="swup-aside"` to the two swap containers
3. **`static/js/swup-init.js`** (new) — Initialize swup with containers, plugins, and body class hook
4. **`layouts/partials/head.html`** — Add `<script src>` for swup-init.js (deferred, after CDN scripts)

## Behavior

1. User clicks any `<a>` link
2. Swup intercepts, fetches target page via AJAX
3. Head Plugin updates `<title>` and meta tags from the fetched page's `<head>`
4. Swup swaps content of `#swup-body` and `#swup-aside`
5. Scripts Plugin executes any `<script>` tags in the new content (e.g., resume-viewer.js)
6. `page:view` hook toggles `not-home` body class
7. URL updates via History API; browser back/forward works

## Graceful Degradation

If JavaScript fails to load, all `<a>` links still work as normal server-side navigation. No functionality is lost — only the smooth transition enhancement.

## Testing

- Verify navigation between all page types (home, post list, single post, resume) without reload
- Verify `<title>` updates correctly on each navigation
- Verify resume viewer works when navigating to /resume (not just on direct load)
- Verify activity calendar persists and remains interactive across navigations
- Verify browser back/forward works correctly
- Verify body class toggles correctly (homepage vs non-homepage styling)
- Verify email copy button still works after navigating away and back
