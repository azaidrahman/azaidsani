package models

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
)

var shortcodeRe = regexp.MustCompile(`\{\{<\s*(movies|mid-img)\s+src="([^"]+)"(?:\s+caption="([^"]*)")?\s*>\}\}`)

// ReplaceShortcodes converts Hugo shortcodes to HTML for preview rendering.
func ReplaceShortcodes(markdown string) string {
	return shortcodeRe.ReplaceAllStringFunc(markdown, func(match string) string {
		parts := shortcodeRe.FindStringSubmatch(match)
		if len(parts) < 3 {
			return match
		}
		scType := parts[1]
		src := parts[2]
		caption := ""
		if len(parts) > 3 {
			caption = parts[3]
		}

		alt := caption
		if alt == "" {
			alt = src
		}

		result := `<figure class="` + scType + `"><img src="` + src + `" alt="` + alt + `">`
		if caption != "" {
			result += `<figcaption>` + caption + `</figcaption>`
		}
		result += `</figure>`
		return result
	})
}

// RenderPreview converts markdown (with Hugo shortcodes) to HTML.
func RenderPreview(markdown string) (string, error) {
	processed := ReplaceShortcodes(markdown)
	md := goldmark.New(
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(processed), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
