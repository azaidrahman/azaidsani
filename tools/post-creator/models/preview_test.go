package models

import (
	"strings"
	"testing"
)

func TestReplaceShortcodes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
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
