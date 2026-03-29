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

	names := make([]string, len(tags))
	for i, tag := range tags {
		names[i] = tag.Name
	}
	wantNames := []string{"css", "go", "htmx", "testing"}
	if !slices.Equal(names, wantNames) {
		t.Errorf("tag names = %v, want %v", names, wantNames)
	}

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
		{"GO", []string{"go", "golang"}},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			results := SearchTags(allTags, tt.query)
			got := make([]string, len(results))
			for i, r := range results {
				got[i] = r.Name
			}
			if len(tt.want) == 0 && len(got) == 0 {
				return
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

	suggestions := SuggestTags(posts[0], posts, 5)
	if len(suggestions) < 1 {
		t.Fatal("expected at least 1 suggestion")
	}
	if suggestions[0] != "htmx" {
		t.Errorf("top suggestion = %q, want 'htmx'", suggestions[0])
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

	found := false
	for _, g := range groups {
		if slices.Contains(g.Tags, "devops") && slices.Contains(g.Tags, "dev-ops") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected devops/dev-ops similarity group, got %v", groups)
	}

	for _, g := range groups {
		if slices.Contains(g.Tags, "go") && slices.Contains(g.Tags, "golang") {
			t.Error("go and golang should not be in same similarity group")
		}
	}
}

func TestRenameTag(t *testing.T) {
	tmpDir := t.TempDir()
	copyTestPosts(t, tmpDir)

	modified, err := RenameTag(tmpDir, "go", "golang")
	if err != nil {
		t.Fatalf("RenameTag error: %v", err)
	}
	if len(modified) != 2 {
		t.Errorf("modified %d files, want 2", len(modified))
	}

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

	posts, _ := ParseAllPosts(tmpDir)
	for _, p := range posts {
		if slices.Contains(p.Tags, "go") || slices.Contains(p.Tags, "testing") {
			t.Errorf("post %q still has old tags: %v", p.Filename, p.Tags)
		}
	}
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
