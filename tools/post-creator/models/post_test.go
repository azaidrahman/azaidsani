package models

import (
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
