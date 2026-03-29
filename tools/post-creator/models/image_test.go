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
		{800, 400, "movies"},
		{600, 400, "mid-img"},
		{400, 600, "mid-img"},
		{1920, 1080, "movies"},
		{500, 500, "mid-img"},
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
