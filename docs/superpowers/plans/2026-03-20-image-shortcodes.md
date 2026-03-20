# Image Shortcodes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add two custom Hugo shortcodes (`movies` and `mid-img`) with auto-detection in the add-image script, replacing the built-in `figure` usage.

**Architecture:** Two shortcode templates render `<figure>` elements with CSS classes. The shell script detects aspect ratio via `sips` and picks the right shortcode. CSS (written by the user) styles each class differently.

**Tech Stack:** Hugo shortcodes (Go templates), Bash, CSS

**Spec:** `docs/superpowers/specs/2026-03-20-image-shortcodes-design.md`

---

### Task 1: Create shortcode templates

**Files:**
- Create: `layouts/shortcodes/movies.html`
- Create: `layouts/shortcodes/mid-img.html`

- [ ] **Step 1: Create the shortcodes directory**

```bash
mkdir -p layouts/shortcodes
```

- [ ] **Step 2: Create `layouts/shortcodes/movies.html`**

```html
<figure class="movies">
  <img src="{{ .Get "src" }}" alt="{{ with .Get "caption" }}{{ . }}{{ else }}{{ .Get "src" }}{{ end }}">
  {{ with .Get "caption" }}<figcaption>{{ . }}</figcaption>{{ end }}
</figure>
```

- [ ] **Step 3: Create `layouts/shortcodes/mid-img.html`**

```html
<figure class="mid-img">
  <img src="{{ .Get "src" }}" alt="{{ with .Get "caption" }}{{ . }}{{ else }}{{ .Get "src" }}{{ end }}">
  {{ with .Get "caption" }}<figcaption>{{ . }}</figcaption>{{ end }}
</figure>
```

- [ ] **Step 4: Verify shortcodes render correctly**

Run: `hugo server -D`

Visit a post that uses `figure` — it should still work. We'll swap the shortcodes in Task 3.

- [ ] **Step 5: Commit**

```bash
git add layouts/shortcodes/movies.html layouts/shortcodes/mid-img.html
git commit -m "feat: add movies and mid-img shortcode templates"
```

---

### Task 2: Rewrite `add-image.sh` with auto-detection

**Files:**
- Modify: `scripts/add-image.sh`

- [ ] **Step 1: Add `sips` guard (after the argument check)**

Add after line 9 (after the `fi` that closes the usage check):

```bash
if ! command -v sips &>/dev/null; then
  echo "Error: sips not found. This script requires macOS."
  exit 1
fi
```

- [ ] **Step 2: Add aspect ratio detection inside the `for` loop**

Add after line 41 (the blank `echo ""` after the move output), before the shortcode collection block:

```bash
  # Detect dimensions and pick tag
  dims=$(sips -g pixelWidth -g pixelHeight "static/images/${cleaned}")
  width=$(echo "$dims" | awk '/pixelWidth/{print $2}')
  height=$(echo "$dims" | awk '/pixelHeight/{print $2}')

  if [ "$((width))" -gt "$((height * 3 / 2))" ]; then
    tag="movies"
  else
    tag="mid-img"
  fi
  echo "Detected: ${width}x${height} -> ${tag}"
```

Note: uses `height * 3 / 2` (integer math) instead of `height * 1.5` since bash doesn't do floating point.

- [ ] **Step 3: Update shortcode output to use the detected tag**

Replace the existing shortcode collection block (current lines 43-48):

```bash
  # Collect shortcode
  if [ -n "$caption" ]; then
    shortcodes+=("{{< ${tag} src=\"/images/${cleaned}\" caption=\"${caption}\" >}}")
  else
    shortcodes+=("{{< ${tag} src=\"/images/${cleaned}\" >}}")
  fi
```

- [ ] **Step 4: Test the script manually**

```bash
# Create a test wide image
sips -z 100 200 static/images/tokyo-fist.jpg --out /tmp/test-wide.jpg
./scripts/add-image.sh /tmp/test-wide.jpg
# Expected: Detected as "movies"

# Create a test square image
sips -z 100 100 static/images/tokyo-fist.jpg --out /tmp/test-square.jpg
./scripts/add-image.sh /tmp/test-square.jpg
# Expected: Detected as "mid-img"
```

Clean up test files from `static/images/` after verifying.

- [ ] **Step 5: Commit**

```bash
git add scripts/add-image.sh
git commit -m "feat: auto-detect image type in add-image script"
```

---

### Task 3: Update existing posts

**Files:**
- Modify: `content/posts/how-i-built-this-website.md:7`
- Modify: `content/posts/building-a-slack-bot.md:8,28`
- Rename: `static/images/geronimo (1).png` → `static/images/geronimo-1.png`

- [ ] **Step 1: Rename the malformed filename**

```bash
mv "static/images/geronimo (1).png" static/images/geronimo-1.png
```

- [ ] **Step 2: Update `how-i-built-this-website.md` line 7**

Change:
```
{{< figure src="/images/tokyo-fist.jpg" caption="Tokyo Fist (1995)" >}}
```
To:
```
{{< movies src="/images/tokyo-fist.jpg" caption="Tokyo Fist (1995)" >}}
```

- [ ] **Step 3: Update `building-a-slack-bot.md` line 8**

Change:
```
{{< figure src="/images/the_presidents_cake-roof_scene.webp" caption="The Presidents Cake (2025)" >}}
```
To:
```
{{< movies src="/images/the_presidents_cake-roof_scene.webp" caption="The Presidents Cake (2025)" >}}
```

- [ ] **Step 4: Update `building-a-slack-bot.md` line 28**

Change:
```
{{< figure src="/images/geronimo (1).png" caption="Jira-nimo Stilton Bot" >}}
```
To:
```
{{< mid-img src="/images/geronimo-1.png" caption="Jira-nimo Stilton Bot" >}}
```

- [ ] **Step 5: Verify with hugo server**

Run: `hugo server -D`

Check both posts — images should render (unstyled until CSS is added).

- [ ] **Step 6: Commit**

```bash
git add content/posts/how-i-built-this-website.md content/posts/building-a-slack-bot.md static/images/geronimo-1.png
git rm "static/images/geronimo (1).png"
git commit -m "feat: migrate existing posts to movies/mid-img shortcodes"
```

---

### Task 4: User writes CSS

**Files:**
- Modify: `static/css/custom.css`

This task is for the user to complete (learning CSS per CLAUDE.md).

**Guidance:**

- [ ] **Step 1: Add `.movies` styles to `static/css/custom.css`**

Target: `.movies` figure should be full content width. The `img` inside should be `width: 100%`. Caption centered below.

- [ ] **Step 2: Add `.mid-img` styles to `static/css/custom.css`**

Target: `.mid-img` figure should be centered (e.g. `margin: 0 auto`). The `img` inside should have `max-height: 350px` and `width: auto`. Caption centered below.

- [ ] **Step 3: Verify both image types look right**

Run: `hugo server -D`

Check both posts. Movies should span full width, mid-img should be centered and capped at 350px height.

- [ ] **Step 4: Commit**

```bash
git add static/css/custom.css
git commit -m "feat: add CSS for movies and mid-img image styles"
```
