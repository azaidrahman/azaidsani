package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"post-creator/models"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	ProjectRoot string
	Templates   *template.Template
}

func NewServer(projectRoot string) (*Server, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{
		ProjectRoot: projectRoot,
		Templates:   tmpl,
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
		s.Templates.ExecuteTemplate(w, "post-filter", data)
	} else {
		s.Templates.ExecuteTemplate(w, "layout.html", data)
	}
}
func (s *Server) PostDetail(w http.ResponseWriter, r *http.Request)        {}
func (s *Server) Preview(w http.ResponseWriter, r *http.Request)           {}
func (s *Server) UpdateFrontmatter(w http.ResponseWriter, r *http.Request) {}
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
func (s *Server) TagSearch(w http.ResponseWriter, r *http.Request)         {}
func (s *Server) TagSuggestions(w http.ResponseWriter, r *http.Request)    {}
func (s *Server) TagDashboard(w http.ResponseWriter, r *http.Request)      {}
func (s *Server) RenameTag(w http.ResponseWriter, r *http.Request)         {}
func (s *Server) MergeTags(w http.ResponseWriter, r *http.Request)         {}
func (s *Server) ImageUpload(w http.ResponseWriter, r *http.Request)       {}
