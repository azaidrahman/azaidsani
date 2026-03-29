package models

import (
	"slices"
	"sort"
	"strings"
)

type TagInfo struct {
	Name  string
	Count int
}

// CollectAllTags returns unique tags with counts, sorted alphabetically.
func CollectAllTags(posts []Post) []TagInfo {
	counts := make(map[string]int)
	for _, p := range posts {
		for _, tag := range p.Tags {
			counts[tag]++
		}
	}

	tags := make([]TagInfo, 0, len(counts))
	for name, count := range counts {
		tags = append(tags, TagInfo{Name: name, Count: count})
	}

	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Name < tags[j].Name
	})

	return tags
}

// SearchTags filters tags by prefix (case-insensitive).
func SearchTags(allTags []TagInfo, query string) []TagInfo {
	q := strings.ToLower(query)
	var results []TagInfo
	for _, tag := range allTags {
		if strings.HasPrefix(strings.ToLower(tag.Name), q) {
			results = append(results, tag)
		}
	}
	return results
}

// SuggestTags recommends tags for a post based on what similar posts use.
func SuggestTags(targetPost Post, allPosts []Post, maxResults int) []string {
	targetTags := make(map[string]bool)
	for _, t := range targetPost.Tags {
		targetTags[t] = true
	}

	// Find posts sharing at least one tag
	freq := make(map[string]int)
	for _, post := range allPosts {
		if post.Filename == targetPost.Filename {
			continue
		}
		shared := false
		for _, t := range post.Tags {
			if targetTags[t] {
				shared = true
				break
			}
		}
		if !shared {
			continue
		}
		for _, t := range post.Tags {
			if !targetTags[t] {
				freq[t]++
			}
		}
	}

	// Sort by frequency descending
	type tagFreq struct {
		name string
		freq int
	}
	var sorted []tagFreq
	for name, f := range freq {
		sorted = append(sorted, tagFreq{name, f})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].freq > sorted[j].freq
	})

	var results []string
	for i, tf := range sorted {
		if i >= maxResults {
			break
		}
		results = append(results, tf.name)
	}
	return results
}

type SimilarGroup struct {
	Tags []string
}

// FindSimilarTags detects tag names that are likely duplicates.
func FindSimilarTags(tags []TagInfo) []SimilarGroup {
	normalize := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "-", "")
		s = strings.ReplaceAll(s, "_", "")
		return s
	}

	groups := make(map[string][]string)
	for _, tag := range tags {
		norm := normalize(tag.Name)
		groups[norm] = append(groups[norm], tag.Name)
	}

	var result []SimilarGroup
	for _, names := range groups {
		if len(names) > 1 {
			sort.Strings(names)
			result = append(result, SimilarGroup{Tags: names})
		}
	}
	return result
}

// RenameTag replaces oldTag with newTag in all posts' frontmatter.
func RenameTag(postsDir, oldTag, newTag string) ([]string, error) {
	posts, err := ParseAllPosts(postsDir)
	if err != nil {
		return nil, err
	}

	var modified []string
	for _, post := range posts {
		if !slices.Contains(post.Tags, oldTag) {
			continue
		}

		newTags := make([]string, 0, len(post.Tags))
		for _, t := range post.Tags {
			if t == oldTag {
				newTags = append(newTags, newTag)
			} else {
				newTags = append(newTags, t)
			}
		}
		post.Tags = newTags

		path := postsDir + "/" + post.Filename
		if err := WriteFrontmatter(path, post); err != nil {
			return modified, err
		}
		modified = append(modified, post.Filename)
	}

	return modified, nil
}

// MergeTags replaces all sourceTags with targetTag in all posts.
func MergeTags(postsDir string, sourceTags []string, targetTag string) ([]string, error) {
	posts, err := ParseAllPosts(postsDir)
	if err != nil {
		return nil, err
	}

	sourceSet := make(map[string]bool)
	for _, s := range sourceTags {
		sourceSet[s] = true
	}

	var modified []string
	for _, post := range posts {
		hasSource := false
		for _, t := range post.Tags {
			if sourceSet[t] {
				hasSource = true
				break
			}
		}
		if !hasSource {
			continue
		}

		hasTarget := false
		newTags := make([]string, 0, len(post.Tags))
		for _, t := range post.Tags {
			if sourceSet[t] {
				if !hasTarget {
					newTags = append(newTags, targetTag)
					hasTarget = true
				}
				continue
			}
			if t == targetTag {
				hasTarget = true
			}
			newTags = append(newTags, t)
		}
		post.Tags = newTags

		path := postsDir + "/" + post.Filename
		if err := WriteFrontmatter(path, post); err != nil {
			return modified, err
		}
		modified = append(modified, post.Filename)
	}

	return modified, nil
}
