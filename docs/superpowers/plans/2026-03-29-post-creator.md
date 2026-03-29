# Post Creator Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> **Learning project:** Per CLAUDE.md, Zaid is learning Go and CSS. This plan provides complete scaffolding, type definitions, test code, and templates. Implementation steps show function signatures and describe expected behavior — Zaid writes the function bodies. Handler and model implementations are marked with **[YOU WRITE]**.

> **Frontmatter format:** Existing posts use YAML (`---` delimiters), not TOML. The spec mentioned TOML but the plan follows what actually exists. New posts will also use YAML for consistency.

**Goal:** Build a local Go + htmx companion tool for managing blog posts (preview, images, tags) and add public tag browsing to the Hugo site.

**Architecture:** Standalone Go HTTP server using `html/template` + htmx for interactivity. Server reads/writes directly to Hugo project directories. No database — filesystem is truth. Hugo template additions for public tag pages are independent of the Go tool.

**Tech Stack:** Go 1.26 stdlib (`net/http`, `html/template`, `embed`), htmx, goldmark, gopkg.in/yaml.v3

---

## File Structure

### Go Tool (`tools/post-creator/`)

```
tools/post-creator/
  main.go                        → entry point: flags, router, server startup
  server.go                      → Server struct, route registration, template loading
  models/
    post.go                      → Post struct, ParsePost, ParseAllPosts, WriteFrontmatter, CreatePost, Slugify
    post_test.go                 → table-driven tests for all post operations
    preview.go                   → ReplaceShortcodes, RenderPreview (goldmark)
    preview_test.go              → tests for shortcode replacement and rendering
    image.go                     → CleanFilename, DetectDimensions, RecommendShortcode, GenerateShortcode
    image_test.go                → tests for image processing
    tag.go                       → CollectAllTags, SearchTags, SuggestTags, FindSimilarTags, RenameTag, MergeTags
    tag_test.go                  → tests for tag operations
  handlers/
    posts.go                     → PostList, PostDetail, Preview, UpdateFrontmatter, CreatePost handlers
    tags.go                      → TagSearch, TagSuggestions, TagDashboard, RenameTag, MergeTags handlers
    images.go                    → ImageUpload handler
  templates/
    layout.html                  → base layout with nav + htmx head
    post-list.html               → post list page
    post-detail.html             → companion view (preview + frontmatter + drop zone)
    tag-dashboard.html           → tag management page
    partials/
      preview.html               → rendered markdown preview fragment
      tag-search.html            → autocomplete dropdown fragment
      tag-suggest.html           → suggested tags fragment
      post-filter.html           → filtered post list fragment
  static/
    htmx.min.js                  → htmx library (embedded)
    style.css                    → tool CSS (approximates site look)
    drag-drop.js                 → image drag-and-drop + clipboard
  testdata/
    content/posts/
      test-post-one.md           → sample published post
      test-post-two.md           → sample draft post
      test-post-three.md         → sample post sharing tags with one
  go.mod
  go.sum
```

### Hugo Changes (project root)

```
layouts/
  _default/taxonomy.html         → NEW: /tags page (all tags with counts)
  _default/term.html             → NEW: /tags/{tag}/ page (posts for one tag)
  _default/single.html           → MODIFY: add tag pills to post header
  partials/about.html            → MODIFY: add tag pills to sidebar
static/
  css/custom.css                 → MODIFY: add pill badge CSS
```

---

## Task 1: Project Scaffolding & Test Data

**Files:**
- Create: `tools/post-creator/go.mod`
- Create: `tools/post-creator/main.go`
- Create: `tools/post-creator/server.go`
- Create: `tools/post-creator/static/htmx.min.js`
- Create: `tools/post-creator/testdata/content/posts/test-post-one.md`
- Create: `tools/post-creator/testdata/content/posts/test-post-two.md`
- Create: `tools/post-creator/testdata/content/posts/test-post-three.md`

- [ ] **Step 1: Create directory structure**

```bash
cd /Users/abdullahzaidas-sani/projects/personal/website/.worktrees/post-creator
mkdir -p tools/post-creator/{models,handlers,templates/partials,static,testdata/content/posts,testdata/static/images}
```

- [ ] **Step 2: Initialize Go module**

```bash
cd tools/post-creator
go mod init post-creator
go get github.com/yuin/goldmark
go get gopkg.in/yaml.v3
```

- [ ] **Step 3: Download htmx**

```bash
curl -o static/htmx.min.js https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js
```

- [ ] **Step 4: Create test data files**

`tools/post-creator/testdata/content/posts/test-post-one.md`:
```markdown
---
title: "Test Post One"
date: 2026-01-15
draft: false
tags: ["go", "testing"]
---

This is the body of test post one.

{{< movies src="/images/test-landscape.jpg" caption="A Landscape Image" >}}

More content after the image.
```

`tools/post-creator/testdata/content/posts/test-post-two.md`:
```markdown
---
title: "Test Post Two"
date: 2026-02-20
draft: true
tags: ["go", "htmx"]
---

Body of test post two.

{{< mid-img src="/images/test-portrait.png" caption="A Portrait Image" >}}
```

`tools/post-creator/testdata/content/posts/test-post-three.md`:
```markdown
---
title: "Test Post Three"
date: 2026-03-10
draft: false
tags: ["testing", "htmx", "css"]
---

Body of test post three with multiple shared tags.
```

- [ ] **Step 5: Create `main.go` (entry point)**

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	port := flag.Int("port", 3333, "server port")
	project := flag.String("project", "../..", "path to Hugo project root")
	flag.Parse()

	// Resolve project root to absolute path
	projectRoot, err := filepath.Abs(*project)
	if err != nil {
		log.Fatalf("invalid project path: %v", err)
	}

	// Verify project root looks like a Hugo project
	if _, err := os.Stat(filepath.Join(projectRoot, "hugo.toml")); err != nil {
		log.Fatalf("no hugo.toml found at %s — is --project pointing to your Hugo project root?", projectRoot)
	}

	srv, err := NewServer(projectRoot)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Post Creator running at http://localhost:%d\n", *port)
	fmt.Printf("Hugo project: %s\n", projectRoot)
	log.Fatal(http.ListenAndServe(addr, srv.Router()))
}
```

- [ ] **Step 6: Create `server.go` (router + template loading)**

```go
package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	ProjectRoot string
	Templates   *template.Template
}

func NewServer(projectRoot string) (*Server, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		ProjectRoot: projectRoot,
		Templates:   tmpl,
	}, nil
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	// Static assets (htmx, css, js) from embedded FS
	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// Serve Hugo project images so they display in preview
	imagesDir := filepath.Join(s.ProjectRoot, "static", "images")
	mux.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir(imagesDir))))

	// Pages
	mux.HandleFunc("GET /", s.PostList)
	mux.HandleFunc("GET /posts/{filename}", s.PostDetail)
	mux.HandleFunc("GET /tags", s.TagDashboard)

	// API — posts
	mux.HandleFunc("GET /api/posts/{filename}/preview", s.Preview)
	mux.HandleFunc("POST /api/posts/{filename}/frontmatter", s.UpdateFrontmatter)
	mux.HandleFunc("POST /api/posts/create", s.CreatePost)

	// API — tags
	mux.HandleFunc("GET /api/tags/search", s.TagSearch)
	mux.HandleFunc("GET /api/posts/{filename}/tag-suggestions", s.TagSuggestions)
	mux.HandleFunc("POST /api/tags/rename", s.RenameTag)
	mux.HandleFunc("POST /api/tags/merge", s.MergeTags)

	// API — images
	mux.HandleFunc("POST /api/images/upload", s.ImageUpload)

	return mux
}

// Stub handlers — each will be implemented in later tasks

func (s *Server) PostList(w http.ResponseWriter, r *http.Request)        {}
func (s *Server) PostDetail(w http.ResponseWriter, r *http.Request)      {}
func (s *Server) Preview(w http.ResponseWriter, r *http.Request)         {}
func (s *Server) UpdateFrontmatter(w http.ResponseWriter, r *http.Request) {}
func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request)      {}
func (s *Server) TagSearch(w http.ResponseWriter, r *http.Request)       {}
func (s *Server) TagSuggestions(w http.ResponseWriter, r *http.Request)  {}
func (s *Server) TagDashboard(w http.ResponseWriter, r *http.Request)    {}
func (s *Server) RenameTag(w http.ResponseWriter, r *http.Request)       {}
func (s *Server) MergeTags(w http.ResponseWriter, r *http.Request)       {}
func (s *Server) ImageUpload(w http.ResponseWriter, r *http.Request)     {}
```

- [ ] **Step 7: Create placeholder template so it compiles**

`tools/post-creator/templates/layout.html`:
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Post Creator</title>
    <link rel="stylesheet" href="/static/style.css">
    <script src="/static/htmx.min.js"></script>
</head>
<body>
    <nav>
        <a href="/">Posts</a>
        <a href="/tags">Tags</a>
    </nav>
    <main>
        {{block "content" .}}{{end}}
    </main>
</body>
</html>
```

`tools/post-creator/static/style.css` (minimal placeholder):
```css
/* Post Creator styles — Zaid will flesh this out */
body { font-family: 'Source Sans Pro', sans-serif; }
```

- [ ] **Step 8: Verify it compiles and runs**

```bash
cd tools/post-creator
go build .
```

Expected: binary compiles with no errors. Server starts and shows empty pages.

- [ ] **Step 9: Commit**

```bash
git add tools/
git commit -m "feat: scaffold post-creator Go project with routing and test data"
```

---

## Task 2: Post Model — Parsing

**Files:**
- Create: `tools/post-creator/models/post.go`
- Create: `tools/post-creator/models/post_test.go`

- [ ] **Step 1: Write the Post struct and function signatures**

`tools/post-creator/models/post.go`:
```go
package models

import "time"

// Post represents a Hugo blog post parsed from a markdown file.
type Post struct {
	Filename string    // just the filename, e.g. "my-post.md"
	Title    string    `yaml:"title"`
	Date     time.Time `yaml:"date"`
	Draft    bool      `yaml:"draft"`
	Tags     []string  `yaml:"tags"`
	Body     string    // raw markdown body (everything after second ---)
}

// ParsePost reads a markdown file and returns a Post.
// It splits the file on "---" delimiters, parses the YAML frontmatter,
// and stores the remaining content as Body.
func ParsePost(filepath string) (Post, error) {
	// [YOU WRITE]
	// 1. Read the file with os.ReadFile
	// 2. Convert to string, split on "---" (the file starts with ---, so split gives ["", frontmatter, body])
	// 3. yaml.Unmarshal the frontmatter section into a Post struct
	// 4. Set Body to everything after the second ---
	// 5. Set Filename to just the base filename (filepath.Base)
	return Post{}, nil
}

// ParseAllPosts scans a directory for .md files and parses each one.
// Returns posts sorted by date descending (newest first).
// Skips files that fail to parse (logs a warning).
func ParseAllPosts(dir string) ([]Post, error) {
	// [YOU WRITE]
	// 1. filepath.Glob(dir + "/*.md")
	// 2. Parse each file with ParsePost
	// 3. Sort by Date descending (sort.Slice)
	return nil, nil
}
```

- [ ] **Step 2: Write the tests**

`tools/post-creator/models/post_test.go`:
```go
package models

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestParsePost(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		wantTitle string
		wantDraft bool
		wantTags  []string
		wantDate  time.Time
		wantBody  string // substring check
	}{
		{
			name:      "published post with images",
			file:      "../testdata/content/posts/test-post-one.md",
			wantTitle: "Test Post One",
			wantDraft: false,
			wantTags:  []string{"go", "testing"},
			wantDate:  time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			wantBody:  "body of test post one",
		},
		{
			name:      "draft post",
			file:      "../testdata/content/posts/test-post-two.md",
			wantTitle: "Test Post Two",
			wantDraft: true,
			wantTags:  []string{"go", "htmx"},
			wantDate:  time.Date(2026, 2, 20, 0, 0, 0, 0, time.UTC),
			wantBody:  "Body of test post two",
		},
		{
			name:      "post with multiple shared tags",
			file:      "../testdata/content/posts/test-post-three.md",
			wantTitle: "Test Post Three",
			wantDraft: false,
			wantTags:  []string{"testing", "htmx", "css"},
			wantDate:  time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			wantBody:  "multiple shared tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			post, err := ParsePost(tt.file)
			if err != nil {
				t.Fatalf("ParsePost(%q) error: %v", tt.file, err)
			}
			if post.Title != tt.wantTitle {
				t.Errorf("Title = %q, want %q", post.Title, tt.wantTitle)
			}
			if post.Draft != tt.wantDraft {
				t.Errorf("Draft = %v, want %v", post.Draft, tt.wantDraft)
			}
			if !slices.Equal(post.Tags, tt.wantTags) {
				t.Errorf("Tags = %v, want %v", post.Tags, tt.wantTags)
			}
			if !post.Date.Equal(tt.wantDate) {
				t.Errorf("Date = %v, want %v", post.Date, tt.wantDate)
			}
			if !containsSubstring(post.Body, tt.wantBody) {
				t.Errorf("Body does not contain %q, got %q", tt.wantBody, post.Body)
			}
		})
	}
}

func TestParsePost_InvalidFile(t *testing.T) {
	_, err := ParsePost("nonexistent.md")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestParseAllPosts(t *testing.T) {
	posts, err := ParseAllPosts("../testdata/content/posts")
	if err != nil {
		t.Fatalf("ParseAllPosts error: %v", err)
	}
	if len(posts) != 3 {
		t.Fatalf("got %d posts, want 3", len(posts))
	}
	// Should be sorted newest first
	if posts[0].Title != "Test Post Three" {
		t.Errorf("first post = %q, want 'Test Post Three' (newest)", posts[0].Title)
	}
	if posts[2].Title != "Test Post One" {
		t.Errorf("last post = %q, want 'Test Post One' (oldest)", posts[2].Title)
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

- [ ] **Step 3: Run tests — verify they fail**

```bash
cd tools/post-creator
go test ./models/ -v -run TestParsePost
```

Expected: tests fail because `ParsePost` returns an empty Post.

- [ ] **Step 4: Implement ParsePost and ParseAllPosts** **[YOU WRITE]**

Implement the function bodies in `models/post.go`. Key Go APIs you'll use:
- `os.ReadFile(filepath)` to read the file
- `strings.SplitN(content, "---", 3)` to split frontmatter from body (3 parts: before first `---`, frontmatter, body)
- `yaml.Unmarshal([]byte(frontmatter), &post)` to parse YAML
- `filepath.Base(path)` for the filename
- `filepath.Glob(pattern)` to find .md files
- `sort.Slice` to sort by date

- [ ] **Step 5: Run tests — verify they pass**

```bash
cd tools/post-creator
go test ./models/ -v -run TestParse
```

Expected: all ParsePost and ParseAllPosts tests pass.

- [ ] **Step 6: Commit**

```bash
git add tools/post-creator/models/post.go tools/post-creator/models/post_test.go
git commit -m "feat: add Post model with YAML frontmatter parsing"
```

---

## Task 3: Post Model — Writing & Creation

**Files:**
- Modify: `tools/post-creator/models/post.go`
- Modify: `tools/post-creator/models/post_test.go`

- [ ] **Step 1: Add function signatures to `post.go`**

Add to `models/post.go`:
```go
// Slugify converts a title to a URL-friendly filename slug.
// "My Cool Post!" -> "my-cool-post"
func Slugify(title string) string {
	// [YOU WRITE]
	// 1. strings.ToLower
	// 2. Replace spaces with hyphens
	// 3. Remove anything that isn't a-z, 0-9, or hyphen (use regexp)
	// 4. Collapse multiple hyphens into one
	// 5. Trim leading/trailing hyphens
	return ""
}

// WriteFrontmatter updates only the YAML frontmatter of a post file,
// preserving the body byte-for-byte. Writes atomically (temp file + rename).
func WriteFrontmatter(filepath string, post Post) error {
	// [YOU WRITE]
	// 1. Read the existing file
	// 2. Split on "---" to isolate the body
	// 3. Marshal the Post struct fields (Title, Date, Draft, Tags) to YAML
	// 4. Reconstruct: "---\n" + yaml + "---\n" + body
	// 5. Write to a temp file in the same directory (os.CreateTemp)
	// 6. os.Rename temp file to the original path (atomic swap)
	return nil
}

// CreatePost scaffolds a new post file with the given title.
// Returns the created filename (e.g., "my-cool-post.md").
func CreatePost(postsDir, title string) (string, error) {
	// [YOU WRITE]
	// 1. Slugify the title to get the filename
	// 2. Check the file doesn't already exist
	// 3. Build YAML frontmatter with title, date=today, draft=true, tags=[]
	// 4. Write to postsDir/slug.md
	return "", nil
}
```

- [ ] **Step 2: Write the tests**

Add to `models/post_test.go`:
```go
func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Cool Post", "my-cool-post"},
		{"Hello, World!", "hello-world"},
		{"  spaces  everywhere  ", "spaces-everywhere"},
		{"Already-Slugged", "already-slugged"},
		{"UPPER CASE", "upper-case"},
		{"special!@#chars$%^", "specialchars"},
		{"multiple---hyphens", "multiple-hyphens"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestWriteFrontmatter_PreservesBody(t *testing.T) {
	// Copy test file to temp directory so we don't modify test data
	tmpDir := t.TempDir()
	src := "../testdata/content/posts/test-post-one.md"
	srcBytes, _ := os.ReadFile(src)
	tmpFile := filepath.Join(tmpDir, "test-post-one.md")
	os.WriteFile(tmpFile, srcBytes, 0644)

	// Parse, modify tags, write back
	post, _ := ParsePost(tmpFile)
	post.Tags = []string{"go", "testing", "new-tag"}
	post.Title = "Updated Title"

	err := WriteFrontmatter(tmpFile, post)
	if err != nil {
		t.Fatalf("WriteFrontmatter error: %v", err)
	}

	// Re-parse and verify
	updated, err := ParsePost(tmpFile)
	if err != nil {
		t.Fatalf("ParsePost after write error: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("Title = %q, want 'Updated Title'", updated.Title)
	}
	if !slices.Equal(updated.Tags, []string{"go", "testing", "new-tag"}) {
		t.Errorf("Tags = %v, want [go testing new-tag]", updated.Tags)
	}
	// Body must be preserved exactly
	if !containsSubstring(updated.Body, "body of test post one") {
		t.Errorf("Body was modified: %q", updated.Body)
	}
	if !containsSubstring(updated.Body, "movies") {
		t.Errorf("Body lost shortcode content: %q", updated.Body)
	}
}

func TestCreatePost(t *testing.T) {
	tmpDir := t.TempDir()
	filename, err := CreatePost(tmpDir, "My New Post")
	if err != nil {
		t.Fatalf("CreatePost error: %v", err)
	}
	if filename != "my-new-post.md" {
		t.Errorf("filename = %q, want 'my-new-post.md'", filename)
	}

	// Verify file exists and has correct frontmatter
	post, err := ParsePost(filepath.Join(tmpDir, filename))
	if err != nil {
		t.Fatalf("ParsePost on new file error: %v", err)
	}
	if post.Title != "My New Post" {
		t.Errorf("Title = %q, want 'My New Post'", post.Title)
	}
	if !post.Draft {
		t.Error("new post should be draft=true")
	}
}

func TestCreatePost_DuplicateFilename(t *testing.T) {
	tmpDir := t.TempDir()
	CreatePost(tmpDir, "Duplicate")
	_, err := CreatePost(tmpDir, "Duplicate")
	if err == nil {
		t.Error("expected error for duplicate filename, got nil")
	}
}
```

- [ ] **Step 3: Run tests — verify they fail**

```bash
go test ./models/ -v -run "TestSlugify|TestWrite|TestCreate"
```

Expected: all fail (functions return zero values).

- [ ] **Step 4: Implement Slugify, WriteFrontmatter, CreatePost** **[YOU WRITE]**

Key Go APIs:
- `regexp.MustCompile("[^a-z0-9-]")` for slug cleaning
- `yaml.Marshal(&frontmatter)` to serialize YAML
- `os.CreateTemp(dir, pattern)` for atomic writes
- `os.Rename(tmp, target)` for atomic swap
- `time.Now().Format("2006-01-02")` for today's date

- [ ] **Step 5: Run tests — verify they pass**

```bash
go test ./models/ -v -run "TestSlugify|TestWrite|TestCreate"
```

- [ ] **Step 6: Commit**

```bash
git add tools/post-creator/models/
git commit -m "feat: add frontmatter writing, post creation, and slugify"
```

---

## Task 4: Preview Rendering

**Files:**
- Create: `tools/post-creator/models/preview.go`
- Create: `tools/post-creator/models/preview_test.go`

- [ ] **Step 1: Write function signatures**

`tools/post-creator/models/preview.go`:
```go
package models

// ReplaceShortcodes converts Hugo shortcodes to HTML so goldmark can render them.
// {{< movies src="/images/foo.jpg" caption="Bar" >}} becomes a <figure class="movies">...</figure>
// {{< mid-img src="/images/foo.jpg" caption="Bar" >}} becomes a <figure class="mid-img">...</figure>
func ReplaceShortcodes(markdown string) string {
	// [YOU WRITE]
	// Use regexp to find {{< movies src="..." caption="..." >}} and {{< mid-img ... >}}
	// Replace with <figure class="TYPE"><img src="SRC" alt="CAPTION"><figcaption>CAPTION</figcaption></figure>
	// Handle case where caption is absent
	return ""
}

// RenderPreview converts a markdown string (with shortcodes already replaced)
// to HTML using goldmark.
func RenderPreview(markdown string) (string, error) {
	// [YOU WRITE]
	// 1. Call ReplaceShortcodes on the input
	// 2. Use goldmark to convert markdown to HTML
	//    md := goldmark.New()
	//    var buf bytes.Buffer
	//    md.Convert([]byte(markdown), &buf)
	// 3. Return buf.String()
	return "", nil
}
```

- [ ] **Step 2: Write the tests**

`tools/post-creator/models/preview_test.go`:
```go
package models

import (
	"strings"
	"testing"
)

func TestReplaceShortcodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string // substring that must appear
	}{
		{
			name:  "movies shortcode",
			input: `Some text {{< movies src="/images/foo.jpg" caption="Bar" >}} more text`,
			want:  `<figure class="movies"><img src="/images/foo.jpg" alt="Bar"><figcaption>Bar</figcaption></figure>`,
		},
		{
			name:  "mid-img shortcode",
			input: `{{< mid-img src="/images/baz.png" caption="Qux" >}}`,
			want:  `<figure class="mid-img"><img src="/images/baz.png" alt="Qux"><figcaption>Qux</figcaption></figure>`,
		},
		{
			name:  "no caption",
			input: `{{< movies src="/images/no-cap.jpg" >}}`,
			want:  `<figure class="movies"><img src="/images/no-cap.jpg"`,
		},
		{
			name:  "no shortcodes",
			input: `Just plain markdown text`,
			want:  `Just plain markdown text`,
		},
		{
			name:  "multiple shortcodes",
			input: `{{< movies src="/images/a.jpg" caption="A" >}} text {{< mid-img src="/images/b.png" caption="B" >}}`,
			want:  `class="movies"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReplaceShortcodes(tt.input)
			if !strings.Contains(got, tt.want) {
				t.Errorf("ReplaceShortcodes result does not contain %q\ngot: %q", tt.want, got)
			}
		})
	}
}

func TestReplaceShortcodes_MultipleInOutput(t *testing.T) {
	input := `{{< movies src="/images/a.jpg" caption="A" >}} text {{< mid-img src="/images/b.png" caption="B" >}}`
	got := ReplaceShortcodes(input)
	if !strings.Contains(got, `class="movies"`) {
		t.Error("missing movies figure")
	}
	if !strings.Contains(got, `class="mid-img"`) {
		t.Error("missing mid-img figure")
	}
}

func TestRenderPreview(t *testing.T) {
	input := "# Hello\n\nSome **bold** text.\n"
	html, err := RenderPreview(input)
	if err != nil {
		t.Fatalf("RenderPreview error: %v", err)
	}
	if !strings.Contains(html, "<h1>Hello</h1>") {
		t.Errorf("missing h1, got: %s", html)
	}
	if !strings.Contains(html, "<strong>bold</strong>") {
		t.Errorf("missing bold, got: %s", html)
	}
}

func TestRenderPreview_WithShortcodes(t *testing.T) {
	input := "# Post\n\n{{< movies src=\"/images/test.jpg\" caption=\"Test\" >}}\n\nMore text.\n"
	html, err := RenderPreview(input)
	if err != nil {
		t.Fatalf("RenderPreview error: %v", err)
	}
	if !strings.Contains(html, `class="movies"`) {
		t.Errorf("shortcode not replaced, got: %s", html)
	}
	if !strings.Contains(html, "More text") {
		t.Errorf("text after shortcode missing, got: %s", html)
	}
}
```

- [ ] **Step 3: Run tests — verify they fail**

```bash
go test ./models/ -v -run "TestReplace|TestRender"
```

- [ ] **Step 4: Implement ReplaceShortcodes and RenderPreview** **[YOU WRITE]**

Key Go APIs:
- `regexp.MustCompile` with named capture groups: `(?P<type>movies|mid-img)`
- `regexp.ReplaceAllStringFunc` or `regexp.ReplaceAllString` for substitution
- `goldmark.New()` and `md.Convert()` for markdown rendering
- `bytes.Buffer` for the goldmark output

- [ ] **Step 5: Run tests — verify they pass**

```bash
go test ./models/ -v -run "TestReplace|TestRender"
```

- [ ] **Step 6: Commit**

```bash
git add tools/post-creator/models/preview.go tools/post-creator/models/preview_test.go
git commit -m "feat: add preview rendering with shortcode replacement"
```

---

## Task 5: Image Processing

**Files:**
- Create: `tools/post-creator/models/image.go`
- Create: `tools/post-creator/models/image_test.go`

- [ ] **Step 1: Write function signatures**

`tools/post-creator/models/image.go`:
```go
package models

import "fmt"

// CleanFilename sanitizes an image filename for web use.
// "My Photo (1).JPG" becomes "my-photo-1.jpg"
func CleanFilename(name string) string {
	// [YOU WRITE]
	// 1. Lowercase
	// 2. Remove brackets and parens
	// 3. Replace spaces with hyphens
	// 4. Remove anything not a-z, 0-9, hyphen, or dot
	// 5. Collapse multiple hyphens
	return ""
}

// DetectDimensions reads an image file and returns its width and height.
func DetectDimensions(filepath string) (width, height int, err error) {
	// [YOU WRITE]
	// 1. os.Open the file
	// 2. image.DecodeConfig (import _ "image/jpeg" and _ "image/png" for format support)
	// 3. Return config.Width, config.Height
	return 0, 0, nil
}

// RecommendShortcode returns "movies" if width > 1.5*height, else "mid-img".
func RecommendShortcode(width, height int) string {
	// [YOU WRITE]
	return ""
}

// GenerateShortcode builds the Hugo shortcode string.
func GenerateShortcode(shortcodeType, filename, caption string) string {
	src := fmt.Sprintf("/images/%s", filename)
	// [YOU WRITE]
	// Return: {{< TYPE src="SRC" caption="CAPTION" >}}
	// If caption is empty, omit the caption attribute
	return ""
}
```

- [ ] **Step 2: Write the tests**

`tools/post-creator/models/image_test.go`:
```go
package models

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestCleanFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Photo.JPG", "my-photo.jpg"},
		{"screenshot (1).png", "screenshot-1.png"},
		{"hello [world].jpeg", "hello-world.jpeg"},
		{"UPPER CASE.PNG", "upper-case.png"},
		{"already-clean.jpg", "already-clean.jpg"},
		{"multiple   spaces.jpg", "multiple-spaces.jpg"},
		{"special!@#chars.png", "specialchars.png"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := CleanFilename(tt.input)
			if got != tt.want {
				t.Errorf("CleanFilename(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectDimensions(t *testing.T) {
	// Create a test JPEG image (800x400 = landscape)
	tmpDir := t.TempDir()
	landscapePath := filepath.Join(tmpDir, "landscape.jpg")
	createTestImage(t, landscapePath, 800, 400, "jpeg")

	w, h, err := DetectDimensions(landscapePath)
	if err != nil {
		t.Fatalf("DetectDimensions error: %v", err)
	}
	if w != 800 || h != 400 {
		t.Errorf("got %dx%d, want 800x400", w, h)
	}

	// Create a test PNG image (400x600 = portrait)
	portraitPath := filepath.Join(tmpDir, "portrait.png")
	createTestImage(t, portraitPath, 400, 600, "png")

	w, h, err = DetectDimensions(portraitPath)
	if err != nil {
		t.Fatalf("DetectDimensions error: %v", err)
	}
	if w != 400 || h != 600 {
		t.Errorf("got %dx%d, want 400x600", w, h)
	}
}

func TestRecommendShortcode(t *testing.T) {
	tests := []struct {
		w, h int
		want string
	}{
		{800, 400, "movies"},   // 2.0 ratio > 1.5
		{600, 400, "mid-img"},  // 1.5 ratio — not strictly greater
		{400, 600, "mid-img"},  // portrait
		{1920, 1080, "movies"}, // widescreen
		{500, 500, "mid-img"},  // square
	}
	for _, tt := range tests {
		got := RecommendShortcode(tt.w, tt.h)
		if got != tt.want {
			t.Errorf("RecommendShortcode(%d, %d) = %q, want %q", tt.w, tt.h, got, tt.want)
		}
	}
}

func TestGenerateShortcode(t *testing.T) {
	tests := []struct {
		scType, filename, caption string
		want                      string
	}{
		{"movies", "my-image.jpg", "A Caption", `{{< movies src="/images/my-image.jpg" caption="A Caption" >}}`},
		{"mid-img", "photo.png", "Photo", `{{< mid-img src="/images/photo.png" caption="Photo" >}}`},
		{"movies", "no-cap.jpg", "", `{{< movies src="/images/no-cap.jpg" >}}`},
	}
	for _, tt := range tests {
		got := GenerateShortcode(tt.scType, tt.filename, tt.caption)
		if got != tt.want {
			t.Errorf("got %q, want %q", got, tt.want)
		}
	}
}

// createTestImage generates a solid-color test image at the given path.
func createTestImage(t *testing.T, path string, width, height int, format string) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{100, 100, 100, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	switch format {
	case "jpeg":
		jpeg.Encode(f, img, nil)
	case "png":
		png.Encode(f, img)
	}
}
```

- [ ] **Step 3: Run tests — verify they fail**

```bash
go test ./models/ -v -run "TestClean|TestDetect|TestRecommend|TestGenerate"
```

- [ ] **Step 4: Implement all image functions** **[YOU WRITE]**

Key Go APIs:
- `strings.ToLower`, `regexp.MustCompile` for filename cleaning
- `os.Open` + `image.DecodeConfig` for dimensions (import `_ "image/jpeg"` and `_ "image/png"`)
- `fmt.Sprintf` for shortcode string building

- [ ] **Step 5: Run tests — verify they pass**

```bash
go test ./models/ -v -run "TestClean|TestDetect|TestRecommend|TestGenerate"
```

- [ ] **Step 6: Commit**

```bash
git add tools/post-creator/models/image.go tools/post-creator/models/image_test.go
git commit -m "feat: add image processing — filename cleaning, dimensions, shortcode generation"
```

---

## Task 6: Tag Operations

**Files:**
- Create: `tools/post-creator/models/tag.go`
- Create: `tools/post-creator/models/tag_test.go`

- [ ] **Step 1: Write function signatures**

`tools/post-creator/models/tag.go`:
```go
package models

// TagInfo holds a tag name and its usage count.
type TagInfo struct {
	Name  string
	Count int
}

// CollectAllTags scans all posts and returns unique tags with counts,
// sorted alphabetically.
func CollectAllTags(posts []Post) []TagInfo {
	// [YOU WRITE]
	// 1. Build a map[string]int counting occurrences
	// 2. Convert to []TagInfo slice
	// 3. Sort alphabetically by Name
	return nil
}

// SearchTags filters tags by prefix (case-insensitive).
func SearchTags(allTags []TagInfo, query string) []TagInfo {
	// [YOU WRITE]
	// Filter where strings.HasPrefix(strings.ToLower(tag.Name), strings.ToLower(query))
	return nil
}

// SuggestTags recommends tags for a post based on what similar posts use.
// Returns up to maxResults tags the post doesn't already have.
func SuggestTags(targetPost Post, allPosts []Post, maxResults int) []string {
	// [YOU WRITE]
	// 1. Find posts that share at least one tag with targetPost
	// 2. Collect all tags from those posts
	// 3. Exclude tags targetPost already has
	// 4. Rank by frequency (most common first)
	// 5. Return top maxResults
	return nil
}

// SimilarGroup represents a group of tags that look like duplicates.
type SimilarGroup struct {
	Tags []string // e.g., ["devops", "dev-ops"]
}

// FindSimilarTags detects tag names that are likely duplicates.
// Normalizes by lowercasing and removing hyphens/underscores before comparing.
func FindSimilarTags(tags []TagInfo) []SimilarGroup {
	// [YOU WRITE]
	// 1. Build map of normalized form -> []original names
	// 2. Groups with 2+ entries are similar
	return nil
}

// RenameTag replaces oldTag with newTag in all posts' frontmatter.
// Returns the filenames of modified posts.
func RenameTag(postsDir, oldTag, newTag string) ([]string, error) {
	// [YOU WRITE]
	// 1. ParseAllPosts
	// 2. For each post with oldTag, replace it with newTag in Tags slice
	// 3. WriteFrontmatter for each modified post
	// 4. Return list of modified filenames
	return nil, nil
}

// MergeTags replaces all sourceTags with targetTag in all posts.
// Returns the filenames of modified posts.
func MergeTags(postsDir string, sourceTags []string, targetTag string) ([]string, error) {
	// [YOU WRITE]
	// 1. ParseAllPosts
	// 2. For each post, replace any sourceTag with targetTag (deduplicate)
	// 3. WriteFrontmatter for each modified post
	// 4. Return list of modified filenames
	return nil, nil
}
```

- [ ] **Step 2: Write the tests**

`tools/post-creator/models/tag_test.go`:
```go
package models

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestCollectAllTags(t *testing.T) {
	posts := []Post{
		{Tags: []string{"go", "testing"}},
		{Tags: []string{"go", "htmx"}},
		{Tags: []string{"testing", "htmx", "css"}},
	}
	tags := CollectAllTags(posts)

	// Should be sorted alphabetically
	names := make([]string, len(tags))
	for i, tag := range tags {
		names[i] = tag.Name
	}
	wantNames := []string{"css", "go", "htmx", "testing"}
	if !slices.Equal(names, wantNames) {
		t.Errorf("tag names = %v, want %v", names, wantNames)
	}

	// Check counts
	for _, tag := range tags {
		switch tag.Name {
		case "go":
			if tag.Count != 2 {
				t.Errorf("go count = %d, want 2", tag.Count)
			}
		case "css":
			if tag.Count != 1 {
				t.Errorf("css count = %d, want 1", tag.Count)
			}
		}
	}
}

func TestSearchTags(t *testing.T) {
	allTags := []TagInfo{
		{Name: "go", Count: 2},
		{Name: "golang", Count: 1},
		{Name: "htmx", Count: 2},
		{Name: "hugo", Count: 1},
	}

	tests := []struct {
		query string
		want  []string
	}{
		{"go", []string{"go", "golang"}},
		{"h", []string{"htmx", "hugo"}},
		{"hu", []string{"hugo"}},
		{"xyz", nil},
		{"GO", []string{"go", "golang"}}, // case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := SearchTags(allTags, tt.query)
			got := make([]string, len(results))
			for i, r := range results {
				got[i] = r.Name
			}
			if len(tt.want) == 0 && len(got) == 0 {
				return // both empty
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("SearchTags(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestSuggestTags(t *testing.T) {
	posts := []Post{
		{Filename: "a.md", Tags: []string{"go", "testing"}},
		{Filename: "b.md", Tags: []string{"go", "htmx"}},
		{Filename: "c.md", Tags: []string{"testing", "htmx", "css"}},
	}

	// Post A has "go" and "testing"
	// Posts sharing tags: B (go), C (testing)
	// Tags from B+C not in A: htmx(2), css(1)
	suggestions := SuggestTags(posts[0], posts, 5)

	if len(suggestions) < 1 {
		t.Fatal("expected at least 1 suggestion")
	}
	if suggestions[0] != "htmx" {
		t.Errorf("top suggestion = %q, want 'htmx' (appears in both B and C)", suggestions[0])
	}
}

func TestFindSimilarTags(t *testing.T) {
	tags := []TagInfo{
		{Name: "devops"},
		{Name: "dev-ops"},
		{Name: "go"},
		{Name: "golang"},
		{Name: "css"},
	}
	groups := FindSimilarTags(tags)

	// "devops" and "dev-ops" normalize to "devops" — should be grouped
	found := false
	for _, g := range groups {
		if slices.Contains(g.Tags, "devops") && slices.Contains(g.Tags, "dev-ops") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected devops/dev-ops similarity group, got %v", groups)
	}

	// "go" and "golang" should NOT be grouped (different normalized forms)
	for _, g := range groups {
		if slices.Contains(g.Tags, "go") && slices.Contains(g.Tags, "golang") {
			t.Error("go and golang should not be in same similarity group")
		}
	}
}

func TestRenameTag(t *testing.T) {
	// Set up temp directory with test posts
	tmpDir := t.TempDir()
	copyTestPosts(t, tmpDir)

	modified, err := RenameTag(tmpDir, "go", "golang")
	if err != nil {
		t.Fatalf("RenameTag error: %v", err)
	}
	if len(modified) != 2 {
		t.Errorf("modified %d files, want 2 (post one and two have 'go' tag)", len(modified))
	}

	// Verify the tag was renamed
	posts, _ := ParseAllPosts(tmpDir)
	for _, p := range posts {
		if slices.Contains(p.Tags, "go") {
			t.Errorf("post %q still has 'go' tag after rename", p.Filename)
		}
	}
}

func TestMergeTags(t *testing.T) {
	tmpDir := t.TempDir()
	copyTestPosts(t, tmpDir)

	modified, err := MergeTags(tmpDir, []string{"go", "testing"}, "development")
	if err != nil {
		t.Fatalf("MergeTags error: %v", err)
	}
	if len(modified) < 1 {
		t.Error("expected at least 1 modified file")
	}

	// Verify: no post should have "go" or "testing", but should have "development"
	posts, _ := ParseAllPosts(tmpDir)
	for _, p := range posts {
		if slices.Contains(p.Tags, "go") || slices.Contains(p.Tags, "testing") {
			t.Errorf("post %q still has old tags: %v", p.Filename, p.Tags)
		}
	}
	// At least one post should have "development"
	hasDev := false
	for _, p := range posts {
		if slices.Contains(p.Tags, "development") {
			hasDev = true
			break
		}
	}
	if !hasDev {
		t.Error("no post has 'development' tag after merge")
	}
}

// copyTestPosts copies the testdata posts to a temp directory for mutation tests.
func copyTestPosts(t *testing.T, dstDir string) {
	t.Helper()
	files := []string{"test-post-one.md", "test-post-two.md", "test-post-three.md"}
	for _, f := range files {
		src := filepath.Join("..", "testdata", "content", "posts", f)
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("copy %s: %v", f, err)
		}
		os.WriteFile(filepath.Join(dstDir, f), data, 0644)
	}
}
```

- [ ] **Step 3: Run tests — verify they fail**

```bash
go test ./models/ -v -run "TestCollect|TestSearch|TestSuggest|TestFindSimilar|TestRename|TestMerge"
```

- [ ] **Step 4: Implement all tag functions** **[YOU WRITE]**

Key Go APIs:
- `map[string]int` for counting
- `sort.Slice` for sorting
- `strings.ToLower`, `strings.ReplaceAll` for normalization
- `strings.HasPrefix` for search
- `slices.Contains` for membership checks

- [ ] **Step 5: Run tests — verify they pass**

```bash
go test ./models/ -v -run "TestCollect|TestSearch|TestSuggest|TestFindSimilar|TestRename|TestMerge"
```

- [ ] **Step 6: Run all model tests to verify nothing is broken**

```bash
go test ./models/ -v
```

Expected: all tests pass.

- [ ] **Step 7: Commit**

```bash
git add tools/post-creator/models/tag.go tools/post-creator/models/tag_test.go
git commit -m "feat: add tag operations — collect, search, suggest, similarity, rename, merge"
```

---

## Task 7: Post List View

**Files:**
- Modify: `tools/post-creator/server.go` (implement PostList handler)
- Create: `tools/post-creator/templates/post-list.html`
- Create: `tools/post-creator/templates/partials/post-filter.html`

- [ ] **Step 1: Create `post-list.html` template**

`tools/post-creator/templates/post-list.html`:
```html
{{define "content"}}
<div class="post-list-header">
    <h1>Posts</h1>
    <form hx-post="/api/posts/create" hx-target="body" hx-swap="outerHTML">
        <input type="text" name="title" placeholder="New post title..." required>
        <button type="submit">New Post</button>
    </form>
</div>

<div id="post-list">
    {{template "post-filter" .}}
</div>
{{end}}
```

- [ ] **Step 2: Create `post-filter.html` partial**

This is the htmx fragment returned when filtering by tag.

`tools/post-creator/templates/partials/post-filter.html`:
```html
{{define "post-filter"}}
<div class="tag-filter">
    {{range .AllTags}}
    <button class="pill {{if eq .Name $.ActiveTag}}active{{end}}"
            hx-get="/?tag={{.Name}}"
            hx-target="#post-list"
            hx-swap="innerHTML"
            hx-push-url="true">
        {{.Name}} <span class="count">{{.Count}}</span>
    </button>
    {{end}}
    {{if .ActiveTag}}
    <button class="pill clear" hx-get="/" hx-target="#post-list" hx-swap="innerHTML" hx-push-url="true">
        clear
    </button>
    {{end}}
</div>

<table class="post-table">
    <thead>
        <tr><th>Title</th><th>Date</th><th>Status</th><th>Tags</th></tr>
    </thead>
    <tbody>
        {{range .Posts}}
        <tr>
            <td><a href="/posts/{{.Filename}}">{{.Title}}</a></td>
            <td>{{.Date.Format "2006-01-02"}}</td>
            <td>{{if .Draft}}<span class="badge draft">draft</span>{{else}}<span class="badge published">published</span>{{end}}</td>
            <td>
                {{range .Tags}}
                <span class="pill small">{{.}}</span>
                {{end}}
            </td>
        </tr>
        {{end}}
    </tbody>
</table>
{{end}}
```

- [ ] **Step 3: Implement PostList handler in `server.go`** **[YOU WRITE]**

Replace the stub `PostList` method. It should:
```go
func (s *Server) PostList(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. Call models.ParseAllPosts(filepath.Join(s.ProjectRoot, "content", "posts"))
	// 2. Call models.CollectAllTags(posts)
	// 3. Read r.URL.Query().Get("tag") for optional filter
	// 4. If tag filter is set, filter posts to only those with that tag
	// 5. Build template data: Posts, AllTags, ActiveTag
	// 6. If request has "HX-Request" header, render just "post-filter" template
	//    Otherwise render "layout.html" with "content" block from "post-list.html"
	// 7. s.Templates.ExecuteTemplate(w, templateName, data)
}
```

Check the `HX-Request` header to decide whether to return the full page or just the htmx fragment:
```go
if r.Header.Get("HX-Request") == "true" {
    s.Templates.ExecuteTemplate(w, "post-filter", data)
} else {
    s.Templates.ExecuteTemplate(w, "layout.html", data)
}
```

- [ ] **Step 4: Implement CreatePost handler** **[YOU WRITE]**

```go
func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. r.ParseForm()
	// 2. title := r.FormValue("title")
	// 3. Call models.CreatePost(postsDir, title)
	// 4. Redirect to /posts/{filename} with http.Redirect
}
```

- [ ] **Step 5: Update template loading in `server.go`**

The `NewServer` function's `template.ParseFS` call needs to pick up nested partials. Ensure the glob patterns are:
```go
tmpl, err := template.ParseFS(templateFS, "templates/*.html", "templates/partials/*.html")
```

- [ ] **Step 6: Manual test**

```bash
cd tools/post-creator
go run . --project ../..
# Open http://localhost:3333 — should see post list with your 2 real posts
```

- [ ] **Step 7: Commit**

```bash
git add tools/post-creator/
git commit -m "feat: add post list view with tag filtering and post creation"
```

---

## Task 8: Post Companion View (Preview + Frontmatter Editing)

**Files:**
- Modify: `tools/post-creator/server.go` (implement PostDetail, Preview, UpdateFrontmatter)
- Create: `tools/post-creator/templates/post-detail.html`
- Create: `tools/post-creator/templates/partials/preview.html`
- Create: `tools/post-creator/templates/partials/tag-search.html`
- Create: `tools/post-creator/templates/partials/tag-suggest.html`

- [ ] **Step 1: Create `post-detail.html` template**

`tools/post-creator/templates/post-detail.html`:
```html
{{define "content"}}
<div class="companion">
    <div class="companion-header">
        <h1>{{.Post.Title}}</h1>
        <a href="/" class="back-link">Back to posts</a>
    </div>

    <!-- Frontmatter controls -->
    <form class="frontmatter-form" hx-post="/api/posts/{{.Post.Filename}}/frontmatter" hx-swap="none">
        <div class="form-row">
            <label>Title</label>
            <input type="text" name="title" value="{{.Post.Title}}">
        </div>
        <div class="form-row">
            <label>Date</label>
            <input type="date" name="date" value="{{.Post.Date.Format "2006-01-02"}}">
        </div>
        <div class="form-row">
            <label>Draft</label>
            <input type="checkbox" name="draft" {{if .Post.Draft}}checked{{end}}>
        </div>
        <div class="form-row">
            <label>Tags</label>
            <div class="tag-editor">
                <div class="current-tags" id="current-tags">
                    {{range .Post.Tags}}
                    <span class="pill editable" data-tag="{{.}}">
                        {{.}}
                        <input type="hidden" name="tags" value="{{.}}">
                        <button type="button" class="remove-tag" onclick="this.parentElement.remove()">x</button>
                    </span>
                    {{end}}
                </div>
                <input type="text" id="tag-input" placeholder="Add tag..."
                       hx-get="/api/tags/search"
                       hx-trigger="keyup changed delay:200ms"
                       hx-target="#tag-dropdown"
                       hx-include="[name='q']"
                       name="q"
                       autocomplete="off">
                <div id="tag-dropdown"></div>
            </div>
        </div>
        <div class="form-row"
             hx-get="/api/posts/{{.Post.Filename}}/tag-suggestions"
             hx-trigger="load"
             hx-target="#tag-suggestions">
            <label>Suggested</label>
            <div id="tag-suggestions"></div>
        </div>
        <button type="submit" class="save-btn">Save Frontmatter</button>
    </form>

    <!-- Image drop zone -->
    <div class="drop-zone" id="drop-zone">
        Drop images here or click to upload
        <input type="file" id="image-input" accept="image/*" hidden>
    </div>

    <!-- Image modal (hidden by default, shown by drag-drop.js) -->
    <div class="image-modal" id="image-modal" style="display:none;">
        <div class="modal-content">
            <img id="modal-preview" src="" alt="Preview">
            <div class="modal-controls">
                <label>Type</label>
                <select id="shortcode-type">
                    <option value="movies">movies (full-width)</option>
                    <option value="mid-img">mid-img (centered)</option>
                </select>
                <label>Caption</label>
                <input type="text" id="caption-input" placeholder="Image caption...">
                <button id="copy-shortcode">Copy to Clipboard</button>
                <button id="close-modal">Close</button>
            </div>
        </div>
    </div>

    <!-- Live preview -->
    <div class="preview-panel"
         hx-get="/api/posts/{{.Post.Filename}}/preview"
         hx-trigger="every 1s"
         hx-swap="innerHTML">
        {{template "preview" .PreviewHTML}}
    </div>
</div>

<script src="/static/drag-drop.js"></script>
{{end}}
```

- [ ] **Step 2: Create preview partial**

`tools/post-creator/templates/partials/preview.html`:
```html
{{define "preview"}}
<div class="preview-content">
    {{.}}
</div>
{{end}}
```

- [ ] **Step 3: Create tag-search partial**

`tools/post-creator/templates/partials/tag-search.html`:
```html
{{define "tag-search"}}
{{range .}}
<div class="tag-option" data-tag="{{.Name}}">
    {{.Name}} <span class="count">{{.Count}}</span>
</div>
{{end}}
{{end}}
```

- [ ] **Step 4: Create tag-suggest partial**

`tools/post-creator/templates/partials/tag-suggest.html`:
```html
{{define "tag-suggest"}}
{{range .}}
<button type="button" class="pill suggestion" data-tag="{{.}}">+ {{.}}</button>
{{end}}
{{end}}
```

- [ ] **Step 5: Implement PostDetail handler** **[YOU WRITE]**

```go
func (s *Server) PostDetail(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. filename := r.PathValue("filename")
	// 2. Parse the post from filepath.Join(s.ProjectRoot, "content", "posts", filename)
	// 3. Render preview HTML with models.RenderPreview(post.Body)
	// 4. Build template data: Post, PreviewHTML (template.HTML type for unescaped HTML)
	// 5. Execute "layout.html" template with data
}
```

- [ ] **Step 6: Implement Preview handler (with content hashing)** **[YOU WRITE]**

```go
func (s *Server) Preview(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. filename := r.PathValue("filename")
	// 2. Read the file, compute hash (md5 or sha256 of content)
	// 3. Compare with If-None-Match header from request
	// 4. If same hash, return 304 Not Modified
	// 5. Otherwise, parse post, render preview, set ETag header, return preview fragment
	//
	// Key: use crypto/md5 or crypto/sha256 for hashing
	// Set w.Header().Set("ETag", hash)
	// Check r.Header.Get("If-None-Match") == hash then w.WriteHeader(304)
}
```

Note: htmx respects 304 responses and won't swap content — this prevents flicker.

- [ ] **Step 7: Implement UpdateFrontmatter handler** **[YOU WRITE]**

```go
func (s *Server) UpdateFrontmatter(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. filename := r.PathValue("filename")
	// 2. r.ParseForm()
	// 3. Build Post from form values: title, date (parse), draft (checkbox), tags (multi-value)
	// 4. Call models.WriteFrontmatter(filepath, post)
	// 5. Return 200 OK (hx-swap="none" means no DOM update needed)
}
```

- [ ] **Step 8: Implement TagSearch handler** **[YOU WRITE]**

```go
func (s *Server) TagSearch(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. query := r.URL.Query().Get("q")
	// 2. Parse all posts, collect all tags
	// 3. models.SearchTags(allTags, query)
	// 4. Render "tag-search" template with results
}
```

- [ ] **Step 9: Implement TagSuggestions handler** **[YOU WRITE]**

```go
func (s *Server) TagSuggestions(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. filename := r.PathValue("filename")
	// 2. Parse the target post + all posts
	// 3. models.SuggestTags(targetPost, allPosts, 5)
	// 4. Render "tag-suggest" template with results
}
```

- [ ] **Step 10: Manual test**

```bash
cd tools/post-creator
go run . --project ../..
# Open http://localhost:3333
# Click on a post — should see preview panel, frontmatter controls, tag autocomplete
# Edit a post in nvim in another terminal — preview should update within 1 second
```

- [ ] **Step 11: Commit**

```bash
git add tools/post-creator/
git commit -m "feat: add post companion view with live preview and frontmatter editing"
```

---

## Task 9: Tag Dashboard

**Files:**
- Modify: `tools/post-creator/server.go` (implement TagDashboard, RenameTag, MergeTags)
- Create: `tools/post-creator/templates/tag-dashboard.html`

- [ ] **Step 1: Create `tag-dashboard.html` template**

`tools/post-creator/templates/tag-dashboard.html`:
```html
{{define "content"}}
<h1>Tag Dashboard</h1>

{{if .SimilarGroups}}
<div class="similar-tags">
    <h2>Possible Duplicates</h2>
    {{range .SimilarGroups}}
    <div class="similar-group">
        {{range .Tags}}
        <span class="pill warning">{{.}}</span>
        {{end}}
        <form class="inline-merge" hx-post="/api/tags/merge" hx-target="body" hx-confirm="Merge these tags?">
            {{range .Tags}}
            <input type="hidden" name="sources" value="{{.}}">
            {{end}}
            <input type="text" name="target" placeholder="Merge into..." class="merge-input" required>
            <button type="submit" class="btn-small">Merge</button>
        </form>
    </div>
    {{end}}
</div>
{{end}}

<div class="tag-grid">
    {{range .Tags}}
    <div class="tag-card">
        <div class="tag-header">
            <span class="pill">{{.Name}} <span class="count">{{.Count}}</span></span>
            <form class="inline-rename" hx-post="/api/tags/rename" hx-target="body" hx-confirm="Rename this tag?">
                <input type="hidden" name="old" value="{{.Name}}">
                <input type="text" name="new" placeholder="New name..." class="rename-input" required>
                <button type="submit" class="btn-small">Rename</button>
            </form>
        </div>
    </div>
    {{end}}
</div>
{{end}}
```

- [ ] **Step 2: Implement TagDashboard handler** **[YOU WRITE]**

```go
func (s *Server) TagDashboard(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. Parse all posts
	// 2. Collect all tags with models.CollectAllTags
	// 3. Find similar tags with models.FindSimilarTags
	// 4. Build template data: Tags, SimilarGroups
	// 5. Execute "layout.html" template
}
```

- [ ] **Step 3: Implement RenameTag handler** **[YOU WRITE]**

```go
func (s *Server) RenameTag(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. r.ParseForm()
	// 2. old := r.FormValue("old"), new := r.FormValue("new")
	// 3. Call models.RenameTag(postsDir, old, new)
	// 4. Redirect back to /tags
}
```

- [ ] **Step 4: Implement MergeTags handler** **[YOU WRITE]**

```go
func (s *Server) MergeTags(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. r.ParseForm()
	// 2. sources := r.Form["sources"], target := r.FormValue("target")
	// 3. Call models.MergeTags(postsDir, sources, target)
	// 4. Redirect back to /tags
}
```

- [ ] **Step 5: Manual test**

```bash
cd tools/post-creator
go run . --project ../..
# Open http://localhost:3333/tags
# Should see all tags with counts
# Test rename and merge with test data (use a copy to be safe)
```

- [ ] **Step 6: Commit**

```bash
git add tools/post-creator/
git commit -m "feat: add tag dashboard with rename, merge, and similarity detection"
```

---

## Task 10: Image Upload & Drop Zone

**Files:**
- Modify: `tools/post-creator/server.go` (implement ImageUpload)
- Create: `tools/post-creator/static/drag-drop.js`

- [ ] **Step 1: Create `drag-drop.js`**

`tools/post-creator/static/drag-drop.js`:
```javascript
(function() {
    var dropZone = document.getElementById('drop-zone');
    var fileInput = document.getElementById('image-input');
    var modal = document.getElementById('image-modal');

    if (!dropZone) return;

    // Click to open file picker
    dropZone.addEventListener('click', function() { fileInput.click(); });
    fileInput.addEventListener('change', function(e) { handleFiles(e.target.files); });

    // Drag and drop
    dropZone.addEventListener('dragover', function(e) {
        e.preventDefault();
        dropZone.classList.add('drag-over');
    });
    dropZone.addEventListener('dragleave', function() {
        dropZone.classList.remove('drag-over');
    });
    dropZone.addEventListener('drop', function(e) {
        e.preventDefault();
        dropZone.classList.remove('drag-over');
        handleFiles(e.dataTransfer.files);
    });

    function handleFiles(files) {
        if (files.length === 0) return;
        var file = files[0];
        if (!file.type.startsWith('image/')) return;

        var formData = new FormData();
        formData.append('image', file);

        fetch('/api/images/upload', { method: 'POST', body: formData })
            .then(function(r) { return r.json(); })
            .then(function(data) { showModal(data); })
            .catch(function(err) { console.error('Upload failed:', err); });
    }

    function showModal(data) {
        document.getElementById('modal-preview').src = '/images/' + data.filename;
        document.getElementById('shortcode-type').value = data.recommended_shortcode;
        document.getElementById('caption-input').value = '';
        modal.dataset.filename = data.filename;
        modal.style.display = 'flex';
    }

    // Copy shortcode to clipboard
    document.getElementById('copy-shortcode').addEventListener('click', function() {
        var type = document.getElementById('shortcode-type').value;
        var caption = document.getElementById('caption-input').value;
        var filename = modal.dataset.filename;
        var src = '/images/' + filename;

        var shortcode;
        if (caption) {
            shortcode = '{{< ' + type + ' src="' + src + '" caption="' + caption + '" >}}';
        } else {
            shortcode = '{{< ' + type + ' src="' + src + '" >}}';
        }

        var btn = document.getElementById('copy-shortcode');
        navigator.clipboard.writeText(shortcode).then(function() {
            btn.textContent = 'Copied!';
            setTimeout(function() { btn.textContent = 'Copy to Clipboard'; }, 1500);
        });
    });

    // Close modal
    document.getElementById('close-modal').addEventListener('click', function() {
        modal.style.display = 'none';
    });

    // Add tag helper for autocomplete (used by tag-search and tag-suggest partials)
    document.addEventListener('click', function(e) {
        // Handle tag option clicks (autocomplete dropdown)
        var tagOption = e.target.closest('.tag-option');
        if (tagOption) {
            addTag(tagOption.dataset.tag);
            return;
        }
        // Handle suggestion pill clicks
        var suggestion = e.target.closest('.suggestion');
        if (suggestion) {
            addTag(suggestion.dataset.tag);
            return;
        }
    });

    function addTag(tagName) {
        var container = document.getElementById('current-tags');
        var input = document.getElementById('tag-input');
        var dropdown = document.getElementById('tag-dropdown');

        // Check if tag already exists
        var existing = container.querySelectorAll('input[name="tags"]');
        for (var i = 0; i < existing.length; i++) {
            if (existing[i].value === tagName) return;
        }

        // Build the pill using safe DOM methods
        var pill = document.createElement('span');
        pill.className = 'pill editable';
        pill.dataset.tag = tagName;

        var text = document.createTextNode(tagName + ' ');
        pill.appendChild(text);

        var hidden = document.createElement('input');
        hidden.type = 'hidden';
        hidden.name = 'tags';
        hidden.value = tagName;
        pill.appendChild(hidden);

        var removeBtn = document.createElement('button');
        removeBtn.type = 'button';
        removeBtn.className = 'remove-tag';
        removeBtn.textContent = 'x';
        removeBtn.addEventListener('click', function() { pill.remove(); });
        pill.appendChild(removeBtn);

        container.appendChild(pill);

        if (input) input.value = '';
        if (dropdown) dropdown.textContent = '';
    }
})();
```

- [ ] **Step 2: Implement ImageUpload handler** **[YOU WRITE]**

```go
func (s *Server) ImageUpload(w http.ResponseWriter, r *http.Request) {
	// [YOU WRITE]
	// 1. r.ParseMultipartForm(10 << 20) — 10MB max
	// 2. file, header, _ := r.FormFile("image")
	// 3. Clean the filename: models.CleanFilename(header.Filename)
	// 4. Save to filepath.Join(s.ProjectRoot, "static", "images", cleanedName)
	//    — io.Copy from uploaded file to destination
	// 5. Detect dimensions: models.DetectDimensions(destPath)
	// 6. Recommend shortcode: models.RecommendShortcode(w, h)
	// 7. Return JSON response:
	//    {
	//      "filename": cleanedName,
	//      "width": w,
	//      "height": h,
	//      "recommended_shortcode": recommended,
	//      "shortcode_text": models.GenerateShortcode(recommended, cleanedName, "")
	//    }
	// Use encoding/json and w.Header().Set("Content-Type", "application/json")
}
```

- [ ] **Step 3: Manual test**

```bash
cd tools/post-creator
go run . --project ../..
# Open a post companion view
# Drag an image onto the drop zone
# Verify: image appears in modal, shortcode type is pre-selected
# Add a caption, click "Copy to Clipboard"
# Paste in terminal — verify shortcode format is correct
```

- [ ] **Step 4: Commit**

```bash
git add tools/post-creator/
git commit -m "feat: add image drag-and-drop upload with shortcode clipboard copy"
```

---

## Task 11: Hugo Public Tag Pages

**Files:**
- Create: `layouts/_default/taxonomy.html`
- Create: `layouts/_default/term.html`
- Modify: `layouts/_default/single.html`
- Modify: `layouts/partials/about.html`
- Modify: `static/css/custom.css`

- [ ] **Step 1: Create `/tags` page template**

`layouts/_default/taxonomy.html`:
```html
{{ define "main" }}
<header>
    <h1>Tags</h1>
</header>
<div class="tag-page">
    {{ range .Data.Terms.Alphabetical }}
    <a href="{{ .Page.Permalink }}" class="tag-pill">
        {{ .Page.Title }} <span class="tag-count">{{ .Count }}</span>
    </a>
    {{ end }}
</div>
{{ end }}
```

- [ ] **Step 2: Create `/tags/{tag}/` page template**

`layouts/_default/term.html`:
```html
{{ define "main" }}
<header>
    <h1>Posts tagged "{{ .Title }}"</h1>
    <p><a href="/tags/">All tags</a></p>
</header>
<div class="term-posts">
    {{ range .Pages }}
    <article class="term-post-item">
        <time>{{ .Date.Format "2006-01-02" }}</time>
        <a href="{{ .Permalink }}">{{ .Title }}</a>
    </article>
    {{ end }}
</div>
{{ end }}
```

- [ ] **Step 3: Add tag pills to post header**

Modify `layouts/_default/single.html` — insert after the `post-header__meta` paragraph (after line 7).

The full header section becomes:
```html
    <header class="post-header">
        <h1 class="post-header__title">{{ .Title | markdownify }}</h1>
        <p class="post-header__meta">
            {{ if .Date }}{{ .Date.Format "2 January 2006" }}{{ end }}
            {{ if .ReadingTime }} · {{ .ReadingTime }} min read{{ end }}
        </p>
        {{ with .Params.tags }}
        <div class="post-header__tags">
            {{ range . }}
            <a href="/tags/{{ . | urlize }}/" class="tag-pill">{{ . }}</a>
            {{ end }}
        </div>
        {{ end }}
    </header>
```

- [ ] **Step 4: Add tag pills to sidebar**

Modify `layouts/partials/about.html` — insert after the email copy `</script>` tag and before the `<!-- Activity Calendar -->` comment. **[YOU STYLE]**

```html
<!-- Tags -->
{{ $tags := $.Site.Taxonomies.tags }}
{{ if $tags }}
<div class="sidebar-tags">
    <h2 class="sidebar-tags__title">Tags</h2>
    <div class="sidebar-tags__list">
        {{ range $tags.Alphabetical }}
        <a href="{{ .Page.Permalink }}" class="tag-pill small">{{ .Page.Title }} <span class="tag-count">{{ .Count }}</span></a>
        {{ end }}
    </div>
</div>
{{ end }}
```

- [ ] **Step 5: Add CSS for tag pills** **[YOU STYLE]**

Add to `static/css/custom.css`. Here's the structure — you fill in the visual values (colors, spacing) to match your site's aesthetic:

```css
/* Tag pills — shared style for sidebar, posts, and /tags page */
.tag-pill {
    display: inline-block;
    font-size: 0.75rem;
    padding: 0.1rem 0.45rem;
    border: 1px solid var(--base03);
    border-radius: 3px;
    color: var(--base0);
    text-decoration: none;
    /* Add hover, transition, etc. as you like */
}

.tag-pill .tag-count {
    font-size: 0.65rem;
    color: var(--base01);
}

.tag-pill.small {
    font-size: 0.7rem;
    padding: 0.05rem 0.35rem;
}

/* Sidebar tags section */
.sidebar-tags__title {
    /* Match your existing sidebar heading styles */
}

.sidebar-tags__list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
}

/* /tags page */
.tag-page {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
}

/* /tags/{tag} post list */
.term-post-item {
    display: flex;
    gap: 1rem;
    /* Match your existing post list item styles */
}

.term-post-item time {
    color: var(--base01);
    font-size: 0.85rem;
}

/* Post header tags */
.post-header__tags {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
    margin-top: 0.5rem;
}
```

- [ ] **Step 6: Test with Hugo**

```bash
cd /Users/abdullahzaidas-sani/projects/personal/website/.worktrees/post-creator
hugo server -D
# Check:
# - http://localhost:1313/tags/ — should show all tags as pills
# - http://localhost:1313/tags/ai/ — should show posts tagged "ai"
# - http://localhost:1313/posts/how-i-built-this-website/ — should show tag pills in header
# - Sidebar should show tag pills below social links
```

- [ ] **Step 7: Commit**

```bash
git add layouts/ static/css/custom.css
git commit -m "feat: add public tag pages, post tag pills, and sidebar tags"
```

---

## Completion

After all tasks are done:

1. Run the full model test suite: `cd tools/post-creator && go test ./models/ -v`
2. Run the Go tool and verify all views work: `go run . --project ../..`
3. Run Hugo and verify tag pages: `hugo server -D`
4. Use the `superpowers:finishing-a-development-branch` skill to decide on merge/PR strategy
