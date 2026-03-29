package main

import (
	"crypto/md5"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"post-creator/models"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	ProjectRoot string
	pages       map[string]*template.Template
}

func NewServer(projectRoot string) (*Server, error) {
	// Parse partials as the base template set
	base, err := template.ParseFS(templateFS, "templates/layout.html", "templates/partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("parsing base templates: %w", err)
	}

	// Each page gets its own clone so {{define "content"}} blocks don't conflict
	pages := make(map[string]*template.Template)
	for _, page := range []string{"post-list.html", "post-detail.html", "tag-dashboard.html"} {
		clone, err := base.Clone()
		if err != nil {
			return nil, fmt.Errorf("cloning base for %s: %w", page, err)
		}
		_, err = clone.ParseFS(templateFS, "templates/"+page)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", page, err)
		}
		pages[page] = clone
	}

	return &Server{
		ProjectRoot: projectRoot,
		pages:       pages,
	}, nil
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	imagesDir := filepath.Join(s.ProjectRoot, "static", "images")
	mux.Handle("GET /images/", http.StripPrefix("/images/", http.FileServer(http.Dir(imagesDir))))

	mux.HandleFunc("GET /", s.PostList)
	mux.HandleFunc("GET /posts/{filename}", s.PostDetail)
	mux.HandleFunc("GET /tags", s.TagDashboard)

	mux.HandleFunc("GET /api/posts/{filename}/preview", s.Preview)
	mux.HandleFunc("POST /api/posts/{filename}/frontmatter", s.UpdateFrontmatter)
	mux.HandleFunc("POST /api/posts/create", s.CreatePost)

	mux.HandleFunc("GET /api/tags/search", s.TagSearch)
	mux.HandleFunc("GET /api/posts/{filename}/tag-suggestions", s.TagSuggestions)
	mux.HandleFunc("POST /api/tags/rename", s.RenameTag)
	mux.HandleFunc("POST /api/tags/merge", s.MergeTags)

	mux.HandleFunc("POST /api/images/upload", s.ImageUpload)

	return mux
}

func (s *Server) PostList(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	postsDir := filepath.Join(s.ProjectRoot, "content", "posts")
	posts, err := models.ParseAllPosts(postsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allTags := models.CollectAllTags(posts)
	activeTag := r.URL.Query().Get("tag")

	if activeTag != "" {
		var filtered []models.Post
		for _, p := range posts {
			for _, t := range p.Tags {
				if t == activeTag {
					filtered = append(filtered, p)
					break
				}
			}
		}
		posts = filtered
	}

	data := struct {
		Posts     []models.Post
		AllTags   []models.TagInfo
		ActiveTag string
	}{posts, allTags, activeTag}

	if r.Header.Get("HX-Request") == "true" {
		s.pages["post-list.html"].ExecuteTemplate(w, "post-filter", data)
	} else {
		s.pages["post-list.html"].ExecuteTemplate(w, "layout.html", data)
	}
}
func (s *Server) PostDetail(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	postPath := filepath.Join(s.ProjectRoot, "content", "posts", filename)

	post, err := models.ParsePost(postPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	previewHTML, err := models.RenderPreview(post.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Post        models.Post
		PreviewHTML template.HTML
	}{post, template.HTML(previewHTML)}

	s.pages["post-detail.html"].ExecuteTemplate(w, "layout.html", data)
}
func (s *Server) Preview(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	postPath := filepath.Join(s.ProjectRoot, "content", "posts", filename)

	data, err := os.ReadFile(postPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum(data))
	if r.Header.Get("If-None-Match") == hash {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	post, err := models.ParsePost(postPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	previewHTML, err := models.RenderPreview(post.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("ETag", hash)
	s.pages["post-detail.html"].ExecuteTemplate(w, "preview", template.HTML(previewHTML))
}
func (s *Server) UpdateFrontmatter(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	postPath := filepath.Join(s.ProjectRoot, "content", "posts", filename)

	post, err := models.ParsePost(postPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post.Title = r.FormValue("title")
	dateStr := r.FormValue("date")
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			post.Date = parsed
		}
	}
	post.Draft = r.FormValue("draft") == "on"
	post.Tags = r.Form["tags"]
	if post.Tags == nil {
		post.Tags = []string{}
	}

	if err := models.WriteFrontmatter(postPath, post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
func (s *Server) CreatePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "title is required", http.StatusBadRequest)
		return
	}

	postsDir := filepath.Join(s.ProjectRoot, "content", "posts")
	filename, err := models.CreatePost(postsDir, title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/posts/"+filename, http.StatusSeeOther)
}
func (s *Server) TagSearch(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	postsDir := filepath.Join(s.ProjectRoot, "content", "posts")
	posts, err := models.ParseAllPosts(postsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allTags := models.CollectAllTags(posts)
	results := models.SearchTags(allTags, query)

	s.pages["post-detail.html"].ExecuteTemplate(w, "tag-search", results)
}
func (s *Server) TagSuggestions(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	postPath := filepath.Join(s.ProjectRoot, "content", "posts", filename)

	post, err := models.ParsePost(postPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	postsDir := filepath.Join(s.ProjectRoot, "content", "posts")
	allPosts, err := models.ParseAllPosts(postsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	suggestions := models.SuggestTags(post, allPosts, 5)
	s.pages["post-detail.html"].ExecuteTemplate(w, "tag-suggest", suggestions)
}
func (s *Server) TagDashboard(w http.ResponseWriter, r *http.Request) {
	postsDir := filepath.Join(s.ProjectRoot, "content", "posts")
	posts, err := models.ParseAllPosts(postsDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	allTags := models.CollectAllTags(posts)
	similarGroups := models.FindSimilarTags(allTags)

	data := struct {
		Tags          []models.TagInfo
		SimilarGroups []models.SimilarGroup
	}{allTags, similarGroups}

	s.pages["tag-dashboard.html"].ExecuteTemplate(w, "layout.html", data)
}
func (s *Server) RenameTag(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	oldTag := r.FormValue("old")
	newTag := r.FormValue("new")
	if oldTag == "" || newTag == "" {
		http.Error(w, "old and new tag names required", http.StatusBadRequest)
		return
	}

	postsDir := filepath.Join(s.ProjectRoot, "content", "posts")
	if _, err := models.RenameTag(postsDir, oldTag, newTag); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tags", http.StatusSeeOther)
}
func (s *Server) MergeTags(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sources := r.Form["sources"]
	target := r.FormValue("target")
	if len(sources) == 0 || target == "" {
		http.Error(w, "sources and target required", http.StatusBadRequest)
		return
	}

	postsDir := filepath.Join(s.ProjectRoot, "content", "posts")
	if _, err := models.MergeTags(postsDir, sources, target); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tags", http.StatusSeeOther)
}
func (s *Server) ImageUpload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "no image provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	cleanedName := models.CleanFilename(header.Filename)
	destPath := filepath.Join(s.ProjectRoot, "static", "images", cleanedName)

	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dst.Close()

	width, height, err := models.DetectDimensions(destPath)
	if err != nil {
		// Image saved but dimensions couldn't be read — still respond
		width, height = 0, 0
	}

	recommended := models.RecommendShortcode(width, height)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"filename":              cleanedName,
		"width":                 width,
		"height":                height,
		"recommended_shortcode": recommended,
		"shortcode_text":        models.GenerateShortcode(recommended, cleanedName, ""),
	})
}
