# Image Shortcodes Design

## Summary

Introduce two custom Hugo shortcodes — `movies` (wide/landscape) and `mid-img` (smaller/centered) — to use instead of the built-in `figure` shortcode, with auto-detection in `add-image.sh` and corresponding CSS.

## Tags

- **`movies`**: Full-width images. Used when width > 1.5x height.
- **`mid-img`**: Centered images with `max-height: 350px`. Used otherwise.

## Components

### 1. Script: `scripts/add-image.sh`

- Guard with `command -v sips` check — fail with clear error if unavailable
- Use `sips -g pixelWidth -g pixelHeight` (macOS) to read dimensions
- If `width > height * 1.5` → tag = `movies`, else → tag = `mid-img`
- Clean filename (existing logic unchanged)
- Prompt for optional caption (existing logic unchanged)
- Move file to `static/images/` (existing logic unchanged)
- Output shortcode: `{{< movies src="/images/file.ext" caption="..." >}}` or `{{< mid-img src="/images/file.ext" caption="..." >}}`
- Copy all shortcodes to clipboard (existing logic unchanged)

### 2. Shortcode: `layouts/shortcodes/movies.html`

Renders:
```html
<figure class="movies">
  <img src="{{ .Get "src" }}" alt="{{ with .Get "caption" }}{{ . }}{{ else }}{{ .Get "src" }}{{ end }}">
  {{ with .Get "caption" }}<figcaption>{{ . }}</figcaption>{{ end }}
</figure>
```

### 3. Shortcode: `layouts/shortcodes/mid-img.html`

Renders:
```html
<figure class="mid-img">
  <img src="{{ .Get "src" }}" alt="{{ with .Get "caption" }}{{ . }}{{ else }}{{ .Get "src" }}{{ end }}">
  {{ with .Get "caption" }}<figcaption>{{ . }}</figcaption>{{ end }}
</figure>
```

### 4. CSS: `static/css/custom.css`

User writes CSS targeting:
- `.movies` — full width, centered caption
- `.mid-img` — centered block, `max-height: 350px` on the img, `width: auto`, centered caption

### 5. Existing posts update

- `how-i-built-this-website.md`: tokyo-fist.jpg (650x352, ratio 1.85) → `movies`
- `building-a-slack-bot.md`: the_presidents_cake-roof_scene.webp (2500x1406, ratio 1.78) → `movies`
- `building-a-slack-bot.md`: geronimo (1).png (550x550, ratio 1.0) → `mid-img`
  - Also rename file on disk to `geronimo-1.png` and update the src path (consistent with script's cleaning logic)

## Notes

- `layouts/shortcodes/` directory needs to be created (does not exist yet)
- CSS is written by the user (learning CSS per CLAUDE.md)
- Script and shortcode templates are scaffolding — written by Claude
