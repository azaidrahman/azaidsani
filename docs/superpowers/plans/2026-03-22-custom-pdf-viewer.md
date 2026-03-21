# Custom Resume PDF Viewer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the browser-native PDF embed on the `/resume/` page with a custom themed viewer using pre-rendered PNG images of each resume page.

**Architecture:** Convert the 2-page PDF to high-res PNGs served as `<img>` tags inside a custom viewer container. A Hugo section template (`layouts/resume/list.html`) renders the viewer with a toolbar (Download + Open PDF buttons) and a popup modal. Minimal vanilla JS handles popup toggle. All styles added to the existing `custom.css`.

**Tech Stack:** Hugo (static site generator), vanilla HTML/CSS/JS, base16-dark theme (risotto), Font Awesome 6.6.0 (already loaded)

**Spec:** `docs/superpowers/specs/2026-03-22-custom-pdf-viewer-design.md`

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `static/images/resume-page-1.png` | Create | Pre-rendered page 1 image |
| `static/images/resume-page-2.png` | Create | Pre-rendered page 2 image |
| `layouts/resume/list.html` | Create | Viewer template (toolbar, images, popup) |
| `static/js/resume-viewer.js` | Create | Popup open/close/escape logic |
| `static/css/custom.css` | Modify | Viewer, toolbar, popup, mobile styles |
| `content/resume/_index.md` | Modify | Simplify to frontmatter only |

---

### Task 1: Export PDF pages as PNG images

**Files:**
- Source: `static/Zaid-Resume.pdf`
- Create: `static/images/resume-page-1.png`
- Create: `static/images/resume-page-2.png`

- [ ] **Step 1: Create images directory**

```bash
mkdir -p static/images
```

- [ ] **Step 2: Convert PDF pages to PNG using `sips` (macOS built-in)**

macOS doesn't have a simple CLI for multi-page PDF-to-PNG. Use Python with the `pdf2image` library or the `convert` command if available. Simplest approach with built-in tools:

```bash
# Check if ImageMagick is available
which magick || which convert
```

If `magick` (ImageMagick 7) is available:
```bash
magick -density 200 static/Zaid-Resume.pdf -quality 100 static/images/resume-tmp.png
# This creates resume-tmp-0.png and resume-tmp-1.png (0-indexed)
mv static/images/resume-tmp-0.png static/images/resume-page-1.png
mv static/images/resume-tmp-1.png static/images/resume-page-2.png
```

If ImageMagick is not available, install it:
```bash
brew install imagemagick
```

Then run the conversion above.

- [ ] **Step 3: Verify images look correct**

```bash
# Check file sizes and dimensions
file static/images/resume-page-1.png static/images/resume-page-2.png
```

Expected: PNG images approximately 1700px wide, reasonable file sizes (100-500KB each).

Open them visually to confirm text is crisp:
```bash
open static/images/resume-page-1.png
open static/images/resume-page-2.png
```

- [ ] **Step 4: Commit**

```bash
git add static/images/resume-page-1.png static/images/resume-page-2.png
git commit -m "feat: add pre-rendered resume page images"
```

---

### Task 2: Create the Hugo layout template

**Files:**
- Create: `layouts/resume/list.html`

This template defines the `"main"` and `"aside"` blocks consumed by `layouts/_default/baseof.html`. The base template renders the page structure — this template just fills in the main content area.

- [ ] **Step 1: Create the resume layout directory**

```bash
mkdir -p layouts/resume
```

- [ ] **Step 2: Write the template**

Create `layouts/resume/list.html` with this exact content:

```html
{{ define "main" }}
<div class="resume-viewer">
  <div class="resume-toolbar">
    <span class="toolbar-title">Zaid-Resume.pdf</span>
    <div class="toolbar-actions">
      <button class="toolbar-btn" id="download-btn" type="button">
        <i class="fa-solid fa-download"></i> Download
      </button>
      <a class="toolbar-btn" href="/Zaid-Resume.pdf" target="_blank" rel="noopener">
        <i class="fa-solid fa-arrow-up-right-from-square"></i> Open PDF
      </a>
    </div>
  </div>
  <div class="resume-pages">
    <img src="/images/resume-page-1.png" alt="Resume page 1" class="resume-page-img" loading="lazy">
    <div class="page-separator">page 1 of 2</div>
    <img src="/images/resume-page-2.png" alt="Resume page 2" class="resume-page-img" loading="lazy">
    <div class="page-separator">page 2 of 2</div>
  </div>
</div>

<!-- Download popup -->
<div class="resume-popup-overlay" id="resume-popup" role="dialog" aria-modal="true" aria-label="Download resume">
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
    <button class="popup-close" id="popup-close" type="button">esc to close</button>
  </div>
</div>

<script src="/js/resume-viewer.js"></script>
{{ end }}

{{ define "aside" }}{{ end }}
```

Key details:
- `{{ define "main" }}` and `{{ define "aside" }}` match the blocks in `layouts/_default/baseof.html:16,25`
- No `<h1>` page heading — the toolbar title "Zaid-Resume.pdf" serves as the page identifier. This is a deliberate choice: the viewer should feel like a document viewer, not a blog post.
- Aside is intentionally empty (no sidebar on resume page)
- Font Awesome icons (`fa-solid fa-download`, `fa-solid fa-arrow-up-right-from-square`) are already loaded via `layouts/partials/head.html:15`
- `loading="lazy"` on images for performance
- `type="button"` on buttons to prevent form submission behavior
- Script loaded at bottom so DOM elements exist when it runs

- [ ] **Step 3: Commit**

```bash
git add layouts/resume/list.html
git commit -m "feat: add custom resume viewer layout template"
```

---

### Task 3: Create the popup toggle JavaScript

**Files:**
- Create: `static/js/resume-viewer.js`

- [ ] **Step 1: Write the script**

Create `static/js/resume-viewer.js` with this exact content:

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

  document.addEventListener("keydown", function (e) {
    if (e.key === "Escape" && overlay.style.display === "flex") {
      closePopup();
    }
  });
})();
```

Key details:
- IIFE to avoid polluting global scope
- Three close mechanisms: close button, overlay click, Escape key
- Uses `style.display = "flex"` to show (matching the CSS flexbox centering)
- Checks `e.target === overlay` so clicking inside the popup card doesn't close it

- [ ] **Step 2: Commit**

```bash
git add static/js/resume-viewer.js
git commit -m "feat: add resume popup toggle script"
```

---

### Task 4: Add viewer styles to custom.css

**Files:**
- Modify: `static/css/custom.css` (append after the mobile section, before the closing of the file)

The existing file ends with a `@media (max-width: 45rem)` block at line 308-321. Add the new styles after line 321.

- [ ] **Step 1: Add resume viewer CSS**

Append to the end of `static/css/custom.css`:

```css
/* --- Resume Viewer --- */

.resume-viewer {
    background: var(--off-bg);
    border-radius: 4px;
    overflow: hidden;
}

.resume-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.6rem 1rem;
    background: var(--base01);
    border-bottom: 1px solid var(--base02);
}

.toolbar-title {
    font-family: "Fira Mono", monospace;
    font-size: 0.85rem;
    color: var(--base03);
}

.toolbar-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.35rem 0.75rem;
    background: var(--base02);
    border: 1px solid var(--base03);
    border-radius: 3px;
    color: var(--base05);
    font-family: "Fira Mono", monospace;
    font-size: 0.78rem;
    cursor: pointer;
    transition: background 0.15s ease, color 0.15s ease;
    text-decoration: none;
}

.toolbar-btn:hover {
    background: var(--base03);
    color: var(--base07);
}

.toolbar-actions {
    display: flex;
    gap: 0.5rem;
}

.resume-pages {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 1.5rem;
    gap: 0.5rem;
}

.resume-page-img {
    width: 100%;
    max-width: 680px;
    border-radius: 2px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.4);
}

.page-separator {
    font-family: "Fira Mono", monospace;
    font-size: 0.7rem;
    color: var(--base03);
    text-align: center;
    padding: 0.5rem 0;
}

/* --- Resume Download Popup --- */

.resume-popup-overlay {
    display: none;
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    z-index: 100;
    align-items: center;
    justify-content: center;
}

.resume-popup {
    background: var(--base01);
    border: 1px solid var(--base02);
    border-radius: 6px;
    padding: 1.5rem 2rem;
    max-width: 360px;
    width: 90%;
    text-align: center;
}

.popup-title {
    font-family: "Fira Mono", monospace;
    font-size: 1rem;
    color: var(--base0A);
    margin-bottom: 0.5rem;
}

.popup-subtitle {
    font-size: 0.85rem;
    color: var(--base04);
    margin-bottom: 1.2rem;
}

.popup-actions {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
}

.popup-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 0.6rem 1rem;
    border-radius: 4px;
    font-family: "Fira Mono", monospace;
    font-size: 0.82rem;
    cursor: pointer;
    text-decoration: none;
    transition: background 0.15s ease, filter 0.15s ease;
    border: none;
}

.popup-btn-primary {
    background: var(--base0A);
    color: var(--base00);
}

.popup-btn-primary:hover {
    filter: brightness(1.1);
}

.popup-btn-secondary {
    background: var(--base02);
    color: var(--base05);
    border: 1px solid var(--base03);
}

.popup-btn-secondary:hover {
    background: var(--base03);
}

.popup-close {
    margin-top: 0.8rem;
    font-size: 0.75rem;
    color: var(--base03);
    cursor: pointer;
    background: none;
    border: none;
    font-family: "Fira Mono", monospace;
}

.popup-close:hover {
    color: var(--base05);
}
```

- [ ] **Step 2: Add mobile overrides for the resume viewer**

Insert these rules before the closing `}` of the existing `@media (max-width: 45rem)` block in `static/css/custom.css` (the block that contains `.page__body`, `.page__aside`, and `.not-home .page__aside` rules):

```css
    .resume-toolbar {
        flex-direction: column;
        gap: 0.5rem;
        text-align: center;
    }

    .resume-pages {
        padding: 1rem;
    }
```

**Important:** Do this step BEFORE Step 1 (appending), or search for the `@media` block by content rather than line number, since appending in Step 1 shifts line numbers.

- [ ] **Step 3: Commit**

```bash
git add static/css/custom.css
git commit -m "feat: add resume viewer and popup styles"
```

---

### Task 5: Simplify the resume content file

**Files:**
- Modify: `content/resume/_index.md`

- [ ] **Step 1: Replace content with frontmatter only**

Replace the entire contents of `content/resume/_index.md` with:

```markdown
---
title: "Resume"
---
```

This removes the old `<embed>` tag and download link. The `layouts/resume/list.html` template handles all rendering now. Hugo's template lookup resolves `content/resume/_index.md` -> `layouts/resume/list.html` automatically.

- [ ] **Step 2: Commit**

```bash
git add content/resume/_index.md
git commit -m "refactor: simplify resume page to use custom viewer template"
```

---

### Task 6: Verify everything works end-to-end

- [ ] **Step 1: Start Hugo dev server**

```bash
hugo server -D
```

Expected: Server starts successfully with no errors, serves at `http://localhost:1313`.

- [ ] **Step 2: Verify the resume page renders correctly**

Open `http://localhost:1313/resume/` in a browser. Check:
- Both page images display correctly with box shadows
- Toolbar shows "Zaid-Resume.pdf" label, Download button, and Open PDF link
- "page 1 of 2" and "page 2 of 2" separators appear between and after images
- Styling matches the site's dark theme (dark background, monospace fonts)
- No sidebar appears on the resume page

- [ ] **Step 3: Verify the download popup works**

Click the "Download" button in the toolbar. Check:
- Popup overlay appears with dark backdrop
- Popup card shows "Resume" title, filename, and two buttons
- "Download PDF" button triggers a file download
- "Open in New Tab" opens the PDF in a new browser tab
- Clicking outside the popup closes it
- Pressing Escape closes it
- "esc to close" text button closes it

- [ ] **Step 4: Verify mobile responsiveness**

Resize browser to narrow width (< 720px). Check:
- Toolbar stacks vertically
- Images scale to full width
- Popup fits within the viewport

- [ ] **Step 5: Verify the "Open PDF" toolbar button**

Click the "Open PDF" button in the toolbar. Check:
- Opens `/Zaid-Resume.pdf` in a new browser tab with the browser's native PDF viewer
