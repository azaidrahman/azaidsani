package models

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"regexp"
	"strings"
)

var (
	bracketRe   = regexp.MustCompile(`[\[\]\(\)]`)
	imgCleanRe  = regexp.MustCompile(`[^a-z0-9.\-]`)
	imgHyphenRe = regexp.MustCompile(`-{2,}`)
)

// CleanFilename sanitizes an image filename for web use.
func CleanFilename(name string) string {
	s := strings.ToLower(name)
	s = bracketRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, " ", "-")
	// Split on dot to preserve extension
	parts := strings.SplitN(s, ".", 2)
	base := parts[0]
	ext := ""
	if len(parts) > 1 {
		ext = "." + parts[1]
	}
	base = imgCleanRe.ReplaceAllString(base, "")
	base = imgHyphenRe.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	return base + ext
}

// DetectDimensions reads an image file and returns its width and height.
func DetectDimensions(filepath string) (width, height int, err error) {
	f, err := os.Open(filepath)
	if err != nil {
		return 0, 0, fmt.Errorf("opening image: %w", err)
	}
	defer f.Close()

	config, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("decoding image config: %w", err)
	}

	return config.Width, config.Height, nil
}

// RecommendShortcode returns "movies" if width > 1.5*height, else "mid-img".
func RecommendShortcode(width, height int) string {
	if float64(width) > 1.5*float64(height) {
		return "movies"
	}
	return "mid-img"
}

// GenerateShortcode builds the Hugo shortcode string.
func GenerateShortcode(shortcodeType, filename, caption string) string {
	src := fmt.Sprintf("/images/%s", filename)
	if caption != "" {
		return fmt.Sprintf(`{{< %s src="%s" caption="%s" >}}`, shortcodeType, src, caption)
	}
	return fmt.Sprintf(`{{< %s src="%s" >}}`, shortcodeType, src)
}
