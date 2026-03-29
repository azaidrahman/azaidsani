# Post Creator — Design Spec

## Overview

A local Go + htmx companion tool for composing, previewing, and managing blog posts on zaidsani.com. Writing happens in nvim — this tool handles everything around the writing: live preview, image management, and tag operations. Additionally, a set of Hugo template changes bring tag browsing to the public site.

**Two deliverables:**

1. **Post Creator Tool** — standalone Go binary in `tools/post-creator/`
2. **Public Tag Pages** — Hugo templates + CSS for tag browsing on the live site

---

## Part 1: Post Creator Tool

### Stack

- **Go** (stdlib `net/http`, `html/template`, `os`, `path/filepath`, `image`)
- **htmx** (single JS file, embedded via `embed.FS`)
- **goldmark** (markdown-to-HTML rendering for preview)
- Polling via htmx `hx-trigger="every 1s"` (file change detection for live preview)
- Vanilla JS only for drag-and-drop image handling

### Project Structure

```
tools/post-creator/
  main.go              → entry point, HTTP server, flag parsing
  handlers/            → HTTP handlers
    posts.go           → post list, post detail, preview, frontmatter update
    tags.go            → tag search, rename, merge, recommendations, dashboard
    images.go          → image upload, processing, shortcode generation
  templates/           → Go html/template files
    layout.html        → base layout (nav, structure)
    post-list.html     → post list page
    post-detail.html   → post companion view (preview + tags + images)
    tag-dashboard.html → tag management dashboard
    partials/          → htmx response fragments
      preview.html     → rendered markdown preview
      tag-search.html  → autocomplete dropdown
      tag-suggest.html → recommended tags
      post-filter.html → filtered post list
  static/              → embedded static assets
    htmx.min.js
    style.css          → tool styling (approximates site look)
    drag-drop.js       → image drag-and-drop + clipboard logic
  go.mod
  go.sum
```

### Running

```bash
cd tools/post-creator
go run . --project ../..
# Opens http://localhost:3333
```

The `--project` flag points to the Hugo project root. Defaults to `../..` (assumes `tools/post-creator/` is two levels deep). The tool reads/writes directly to `content/posts/` and `static/images/` relative to the project root.

---

### View 1: Post List (`/`)

The home page of the tool. Lists all posts from `content/posts/`.

**Displays for each post:**
- Title
- Date
- Draft status (badge)
- Tags (as clickable pills)

**Features:**
- Sorted by date, newest first
- Filter by tag — click a tag pill to filter the list (htmx partial update)
- "New Post" button — opens a form for title input, scaffolds the file from the archetype (`archetypes/default.md`), sets date to today, `draft: true`, and creates the file in `content/posts/{slug}.md`
- Click a post title to open the companion view

---

### View 2: Post Companion (`/posts/{filename}`)

The main working view. You write in nvim; this shows you everything else.

**Layout:**
- Top bar: frontmatter controls
- Main area: live preview panel
- Overlay: image upload modal (on drag-and-drop)

#### Frontmatter Controls (top bar)

Form inputs that modify only the frontmatter block of the markdown file, preserving the body exactly:

- **Title** — text input
- **Date** — date picker
- **Draft toggle** — checkbox
- **Tags** — input with autocomplete + removable pill badges for current tags
  - On keystroke (debounced 200ms): `hx-get="/api/tags/search?q=..."` returns matching tags as a dropdown
  - Click or Enter to add a tag
  - Click the `x` on a pill to remove a tag
  - **Suggested tags** row below: when the page loads, `hx-get="/api/posts/{file}/tag-suggestions"` returns recommended tags based on what similar posts use. Click to add.
- **Save** button — writes frontmatter changes back to the file

#### Live Preview Panel

- Renders the full post as HTML, styled to approximate the site (Source Sans Pro font, similar heading sizes, image shortcode classes `.movies` and `.mid-img`)
- Polls for changes: `hx-get="/api/posts/{file}/preview" hx-trigger="every 1s"`
- Server computes a content hash — returns `304 Not Modified` if unchanged, avoiding unnecessary re-renders
- Images referenced in shortcodes are served by the tool from `static/images/` so they display in preview

#### Image Drop Zone

1. Drag an image file anywhere onto the editor view (or a dedicated drop area)
2. Vanilla JS intercepts the drop, uploads via `POST /api/images/upload`
3. Go handler:
   - Cleans filename: lowercase, remove brackets/parens, spaces → hyphens
   - Reads image dimensions via Go `image` package
   - Copies file to `static/images/{cleaned-name}`
   - Returns: `{ filename, width, height, recommended_shortcode, shortcode_text }`
4. Modal appears with:
   - Image preview thumbnail
   - Shortcode type selector (`movies` / `mid-img`) — pre-selected based on aspect ratio (width > 1.5x height → `movies`, else `mid-img`)
   - Caption text input
   - "Copy to Clipboard" button
5. Clicking "Copy to Clipboard" copies the full shortcode string (e.g., `{{< movies src="/images/cleaned-name.jpg" caption="Your caption" >}}`) to the system clipboard
6. You paste it into nvim at the position you want

---

### View 3: Tag Dashboard (`/tags`)

Management view for all tags across all posts.

**Tag list:**
- All unique tags with post count, displayed as a card grid of pill badges (consistent with sidebar style)
- Click a tag to expand/see which posts use it
- Sorted alphabetically

**Rename:**
- Click "rename" next to a tag → inline edit field
- On submit: `POST /api/tags/rename` with `{old, new}`
- Handler scans all `.md` files, replaces the tag in frontmatter, rewrites affected files
- Shows confirmation with list of affected posts before executing

**Merge:**
- Select two or more tags → "Merge" button
- Choose the target tag name
- `POST /api/tags/merge` with `{sources: [...], target}`
- Same file-rewriting logic as rename
- Confirmation with affected posts list

**Similarity detection:**
- On page load, the handler computes string similarity between all tag pairs
- Flags groups that look like duplicates (e.g., `devops` / `dev-ops`) after normalizing hyphens, underscores, and case
- Displayed as a "Possible duplicates" section at the top of the dashboard

---

### Data Flow

**Reading posts:**
- Scans `content/posts/*.md` on each request — no cache, always fresh from disk
- Parses TOML frontmatter (`+++` delimiters) to extract title, date, tags, draft status
- Body parsed separately by goldmark for preview rendering

**Writing frontmatter:**
- Handler reads the file, splits on `+++` delimiters
- Replaces the frontmatter section with updated values
- Preserves the body byte-for-byte
- Writes atomically: write to temp file in same directory, then `os.Rename`

**Creating new posts:**
- Reads `archetypes/default.md` for template
- Generates slug from title (lowercase, spaces → hyphens, strip special chars)
- Writes to `content/posts/{slug}.md`

**Image upload:**
- Receives multipart file upload
- Cleans filename, detects dimensions, copies to `static/images/`
- Returns JSON with shortcode details

**Tag search:**
- Scans all posts, collects unique tags, filters by query prefix
- Returns HTML fragment for htmx to swap into the dropdown

**Tag recommendations:**
- For a given post, find its current tags
- Find all other posts sharing at least one tag
- Collect tags from those posts that the current post doesn't have
- Rank by frequency (most commonly co-occurring tags first)
- Return top 5 as suggestions

---

## Part 2: Public Tag Pages (Hugo)

Hugo template and CSS changes to bring tag browsing to the live site.

### `/tags` Page

A taxonomy list page showing all tags with post counts.

**Layout:**
- Page title: "Tags"
- Tags displayed as pill badges (matching sidebar style) with post count inside each pill
- Click a tag → goes to `/tags/{tag}/`
- Alphabetically sorted

### `/tags/{tag}/` Pages

A taxonomy term page showing all posts for a specific tag.

**Layout:**
- Page title: the tag name
- Posts listed below, same style as the main posts list (date + title)
- Grouped by year if enough posts exist, otherwise a flat list
- Link back to `/tags`

### Tags on Individual Posts

On `single.html` (post pages), display the post's tags as clickable pill badges. Position: below the post title/date in the post header area.

### Sidebar Tag Element

Added to the sidebar (`partials/about.html`), below the existing bio and social links.

**Style: compact pill badges**
- Small "Tags" label (uppercase, muted)
- Tags as small bordered pills with post count, wrapping naturally
- Pill style: `font-size: 0.75rem`, subtle border (`1px solid`), small border-radius, muted text color
- Count displayed inside the pill in slightly smaller/dimmer text
- Each pill links to `/tags/{tag}/`
- Non-distracting — blends with existing sidebar aesthetic

### Hugo Template Files

- `layouts/_default/taxonomy.html` — the `/tags` page (list of all tags)
- `layouts/_default/term.html` — the `/tags/{tag}/` page (posts for one tag)
- Modify `layouts/_default/single.html` — add tag pills to post header
- Modify `layouts/partials/about.html` — add tag section to sidebar
- CSS additions in `static/css/custom.css` — pill badge styles, tag page layout

---

## What This Is NOT

- Not a text editor — you write in nvim
- Not deployed — the Go tool is local-only
- No database — filesystem is the source of truth
- No authentication — it's a local dev tool
- No auto-save — explicit save for frontmatter changes only
- No analytics, comments, or dynamic features on the public site
