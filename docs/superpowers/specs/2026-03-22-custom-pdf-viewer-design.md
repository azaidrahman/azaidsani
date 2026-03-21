# Custom Resume PDF Viewer

## Problem

The resume page at `/resume/` uses a bare `<embed>` tag that delegates rendering to the browser's built-in PDF viewer. This looks inconsistent across browsers, doesn't match the site's dark theme, and some mobile browsers can't display it inline at all.

## Solution

Replace the `<embed>` with pre-rendered PNG images of each resume page, displayed inside a custom themed viewer component. The original PDF stays available for download and open-in-new-tab.

## Approach: Pre-Rendered Images

Convert the 2-page PDF to high-resolution PNG images at build/export time. Serve them as `<img>` tags inside a custom viewer container styled to match the site's base16-dark theme.

### Why this approach

- Zero external JS dependencies (no PDF.js, no Node packages)
- Consistent rendering across all browsers and mobile devices
- Instant load — images are just static assets
- Full visual control — the viewer is just HTML/CSS
- Fits the existing Hugo + static asset architecture

### Trade-offs

- Text in the images is not selectable (mitigated by download/open-in-tab for the real PDF)
- Updating the resume requires re-exporting PNGs (acceptable for a resume that changes infrequently)

## Components

### 1. Pre-rendered page images

- Files: `/static/images/resume-page-1.png`, `/static/images/resume-page-2.png`
- Export at 2x resolution for crisp display on retina screens
- Format: PNG (lossless, clean text rendering)

### 2. Hugo layout: `layouts/resume/list.html`

A dedicated template for the `/resume/` section that renders the viewer. This replaces the current approach of raw HTML inside `_index.md`.

Structure:
```
{{ define "main" }}
<div class="resume-viewer">
  <div class="resume-toolbar">
    <span class="toolbar-title">Zaid-Resume.pdf</span>
    <div class="toolbar-actions">
      <button class="toolbar-btn" id="download-btn">
        <i class="fa-solid fa-download"></i> Download
      </button>
      <a class="toolbar-btn" href="/Zaid-Resume.pdf" target="_blank" rel="noopener">
        <i class="fa-solid fa-arrow-up-right-from-square"></i> Open PDF
      </a>
    </div>
  </div>
  <div class="resume-pages">
    <img src="/images/resume-page-1.png" alt="Resume page 1" class="resume-page-img">
    <div class="page-separator">page 1 of 2</div>
    <img src="/images/resume-page-2.png" alt="Resume page 2" class="resume-page-img">
    <div class="page-separator">page 2 of 2</div>
  </div>
</div>

<!-- Download popup -->
<div class="resume-popup-overlay" id="resume-popup">
  <div class="resume-popup">
    <div class="popup-title">Resume</div>
    <div class="popup-subtitle">Zaid-Resume.pdf</div>
    <div class="popup-actions">
      <a class="popup-btn popup-btn-primary" href="/Zaid-Resume.pdf" download>
        <i class="fa-solid fa-download"></i> Download PDF
      </a>
      <a class="popup-btn popup-btn-secondary" href="/Zaid-Resume.pdf" target="_blank" rel="noopener">
        <i class="fa-solid fa-arrow-up-right-from-square"></i> Open in New Tab
      </a>
    </div>
    <button class="popup-close" id="popup-close">esc to close</button>
  </div>
</div>
{{ end }}
```

### 3. Styles in `static/css/custom.css`

New CSS classes:

- `.resume-viewer` — outer container, `background: var(--off-bg)`, rounded corners
- `.resume-toolbar` — flex row, `background: var(--base01)`, bottom border
- `.toolbar-title` — Fira Mono, muted color
- `.toolbar-btn` — small button with icon, base02 background, hover state
- `.resume-pages` — flex column, centered, gap between pages, padding
- `.resume-page-img` — white background, max-width, box-shadow, border-radius
- `.page-separator` — Fira Mono, small muted text, centered
- `.resume-popup-overlay` — fixed overlay, dark semi-transparent background
- `.resume-popup` — centered card, base01 background, border
- `.popup-btn-primary` — base0A background, dark text (download action)
- `.popup-btn-secondary` — base02 background, light text (open in new tab)
- `.popup-close` — subtle text button

Mobile (`@media max-width: 45rem`):
- Images scale to 100% width
- Toolbar buttons shrink (icon-only or smaller text)
- Popup fits within viewport

### 4. JavaScript: `static/js/resume-viewer.js`

Minimal script (~20 lines):

- Open popup: click handler on `#download-btn` shows `#resume-popup`
- Close popup: click `#popup-close`, click overlay background, or press Escape
- No external dependencies

### 5. Content: `content/resume/_index.md`

Simplified to just frontmatter:
```yaml
---
title: "Resume"
layout: "list"
---
```

The layout template handles all rendering.

## Files Changed

| File | Action |
|------|--------|
| `static/images/resume-page-1.png` | Add (exported from PDF) |
| `static/images/resume-page-2.png` | Add (exported from PDF) |
| `layouts/resume/list.html` | Create (viewer template) |
| `static/js/resume-viewer.js` | Create (popup toggle logic) |
| `static/css/custom.css` | Edit (add viewer styles) |
| `content/resume/_index.md` | Edit (simplify to frontmatter only) |

## Resume Update Workflow

When the resume changes:
1. Export each page as PNG at 2x resolution (e.g., using Preview on macOS: File > Export, 300 DPI)
2. Replace `static/images/resume-page-1.png` and `resume-page-2.png`
3. Replace `static/Zaid-Resume.pdf` with the updated PDF
4. Commit and push — Cloudflare Pages auto-deploys

## Out of Scope

- PDF.js or any JS PDF rendering library
- Text selection within the viewer (available via "Open in New Tab")
- Zoom controls (browser zoom works on the images)
- Print button (browser print works, or user downloads the PDF)
