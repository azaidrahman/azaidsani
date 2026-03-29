package models

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

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
func ParsePost(path string) (Post, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Post{}, fmt.Errorf("reading post: %w", err)
	}

	content := string(data)
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return Post{}, fmt.Errorf("invalid frontmatter in %s: expected --- delimiters", path)
	}

	var post Post
	if err := yaml.Unmarshal([]byte(parts[1]), &post); err != nil {
		return Post{}, fmt.Errorf("parsing frontmatter in %s: %w", path, err)
	}

	post.Filename = filepath.Base(path)
	post.Body = parts[2]
	if strings.HasPrefix(post.Body, "\n") {
		post.Body = post.Body[1:]
	}

	return post, nil
}

// ParseAllPosts scans a directory for .md files and parses each one.
// Returns posts sorted by date descending (newest first).
func ParseAllPosts(dir string) ([]Post, error) {
	pattern := filepath.Join(dir, "*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("globbing posts: %w", err)
	}

	var posts []Post
	for _, f := range files {
		post, err := ParsePost(f)
		if err != nil {
			log.Printf("warning: skipping %s: %v", f, err)
			continue
		}
		posts = append(posts, post)
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}
