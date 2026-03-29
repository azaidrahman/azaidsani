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
		wantBody  string
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
	tmpDir := t.TempDir()
	src := "../testdata/content/posts/test-post-one.md"
	srcBytes, _ := os.ReadFile(src)
	tmpFile := filepath.Join(tmpDir, "test-post-one.md")
	os.WriteFile(tmpFile, srcBytes, 0644)

	post, _ := ParsePost(tmpFile)
	post.Tags = []string{"go", "testing", "new-tag"}
	post.Title = "Updated Title"

	err := WriteFrontmatter(tmpFile, post)
	if err != nil {
		t.Fatalf("WriteFrontmatter error: %v", err)
	}

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
