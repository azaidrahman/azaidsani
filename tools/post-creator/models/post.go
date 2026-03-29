package models

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
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

// Slugify converts a title to a URL-friendly filename slug.
func Slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]")
	s = reg.ReplaceAllString(s, "")
	reg = regexp.MustCompile("-{2,}")
	s = reg.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// WriteFrontmatter updates only the YAML frontmatter of a post file,
// preserving the body byte-for-byte. Writes atomically (temp file + rename).
func WriteFrontmatter(path string, post Post) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	content := string(data)
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return fmt.Errorf("invalid frontmatter in %s", path)
	}

	// Build a frontmatter-only struct to avoid serializing Body/Filename
	fm := struct {
		Title string    `yaml:"title"`
		Date  time.Time `yaml:"date"`
		Draft bool      `yaml:"draft"`
		Tags  []string  `yaml:"tags"`
	}{
		Title: post.Title,
		Date:  post.Date,
		Draft: post.Draft,
		Tags:  post.Tags,
	}

	yamlBytes, err := yaml.Marshal(&fm)
	if err != nil {
		return fmt.Errorf("marshaling frontmatter: %w", err)
	}

	body := parts[2]
	newContent := "---\n" + string(yamlBytes) + "---" + body

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, "post-*.md")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.WriteString(newContent); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// CreatePost scaffolds a new post file with the given title.
// Returns the created filename (e.g., "my-cool-post.md").
func CreatePost(postsDir, title string) (string, error) {
	slug := Slugify(title)
	filename := slug + ".md"
	path := filepath.Join(postsDir, filename)

	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("file already exists: %s", path)
	}

	now := time.Now()
	fm := struct {
		Title string    `yaml:"title"`
		Date  time.Time `yaml:"date"`
		Draft bool      `yaml:"draft"`
		Tags  []string  `yaml:"tags"`
	}{
		Title: title,
		Date:  time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC),
		Draft: true,
		Tags:  []string{},
	}

	yamlBytes, err := yaml.Marshal(&fm)
	if err != nil {
		return "", fmt.Errorf("marshaling frontmatter: %w", err)
	}

	content := "---\n" + string(yamlBytes) + "---\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("writing post: %w", err)
	}

	return filename, nil
}
