# Zaid's Personal Website

A writing-focused personal site.

## Plan

### Phase 1 — Hugo + Cloudflare Pages (current)

- **Hugo** static site generator with the [risotto](https://github.com/joeroe/risotto) theme
- **Cloudflare Pages** for hosting (auto-deploys from `main`)
- **Cloudflare Registrar** for domain
- Homepage inspired by brennan.io — clean, text-forward landing page with short intro and highlighted posts

### Phase 2 — Self-hosted Go server

Migrate to a custom Go server (details TBD).

## Project Structure (Phase 1)

```
content/_index.md  → homepage content (short bio/intro)
content/posts/     → markdown blog posts
content/about/     → optional about page
static/            → static assets
themes/risotto/    → Hugo theme (git submodule)
```

## Local Development

```bash
hugo server -D
```

## Production Build

```bash
hugo --minify
```

Output goes to `public/`.

## Deploy

Push to `main`. Cloudflare Pages builds automatically.
