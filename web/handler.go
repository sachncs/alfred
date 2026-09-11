// Package web serves the htmx-based UI shell.
package web

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/sachncs/alfred/internal/contract"
)

//go:embed templates/layout.html templates/pages/*.html templates/partials/*.html
var templateFS embed.FS

// ThreadLister lists threads from the store.
type ThreadLister interface {
	List(limit int, token string) ([]contract.Thread, string, error)
}

// ThreadGetter gets a single thread.
type ThreadGetter interface {
	Get(id contract.ThreadID) (*contract.Thread, error)
}

// ThreadCreator creates a thread.
type ThreadCreator interface {
	Create(t *contract.Thread) error
}

// SessionReader reads turn items for a thread.
type SessionReader interface {
	Read(threadID contract.ThreadID, offset int) ([]contract.TurnItem, error)
}

// TurnStarter starts a turn (returns the queued turn).
type TurnStarter interface {
	StartTurn(thread *contract.Thread, input contract.UserInput) (*contract.Turn, error)
}

// Handlers holds parsed templates and store references.
type Handlers struct {
	chat     *template.Template
	design   *template.Template
	settings *template.Template
	parts    *template.Template
	Threads  ThreadLister
	Thread   ThreadGetter
	Create   ThreadCreator
	Session  SessionReader
	Starter  TurnStarter
}

var funcMap = template.FuncMap{
	"formatTime": func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.In(time.Local).Format("3:04 PM")
	},
	"formatDate": func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.In(time.Local).Format("Jan 2")
	},
	"lower": strings.ToLower,
}

// NewHandlers parses embedded templates.
func NewHandlers() *Handlers {
	chat, err := template.New("").Funcs(funcMap).ParseFS(templateFS,
		"templates/layout.html",
		"templates/pages/chat.html",
	)
	if err != nil {
		log.Printf("web: parse chat templates: %v", err)
	}

	design, err := template.New("").Funcs(funcMap).ParseFS(templateFS,
		"templates/layout.html",
		"templates/pages/design.html",
		"templates/partials/neuro_gauge.html",
		"templates/partials/neuro_trend_chart.html",
		"templates/partials/neuro_owner_card.html",
	)
	if err != nil {
		log.Printf("web: parse design templates: %v", err)
	}

	parts, err := template.New("").Funcs(funcMap).ParseFS(templateFS,
		"templates/partials/*.html",
	)
	if err != nil {
		log.Printf("web: parse partials: %v", err)
	}

	settings, err := template.New("").Funcs(funcMap).ParseFS(templateFS,
		"templates/layout.html",
		"templates/pages/settings.html",
	)
	if err != nil {
		log.Printf("web: parse settings templates: %v", err)
	}

	return &Handlers{chat: chat, design: design, settings: settings, parts: parts}
}

// Register mounts all web routes on mux. staticDir serves CSS/JS/images.
func Register(mux *http.ServeMux, h *Handlers, staticDir string) {
	mux.HandleFunc("GET /", h.handleChat)
	mux.HandleFunc("GET /design", h.handleDesign)
	mux.HandleFunc("GET /settings", h.handleSettings)
	mux.HandleFunc("GET /ui/sidebar", h.handleSidebar)
	mux.HandleFunc("GET /ui/threads/{id}/timeline", h.handleTimeline)
	mux.HandleFunc("GET /ui/threads/{id}/todos", h.handleTodos)
	mux.HandleFunc("POST /ui/threads", h.handleCreateThread)
	mux.HandleFunc("POST /ui/threads/{id}/turns", h.handleSubmitTurn)
	if staticDir != "" {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	}
}

// RegisterFS mounts web routes using an fs.FS for static files.
func RegisterFS(mux *http.ServeMux, h *Handlers, staticFS fs.FS) {
	mux.HandleFunc("GET /", h.handleChat)
	mux.HandleFunc("GET /design", h.handleDesign)
	mux.HandleFunc("GET /ui/sidebar", h.handleSidebar)
	mux.HandleFunc("GET /ui/threads/{id}/timeline", h.handleTimeline)
	mux.HandleFunc("GET /ui/threads/{id}/todos", h.handleTodos)
	mux.HandleFunc("POST /ui/threads", h.handleCreateThread)
	mux.HandleFunc("POST /ui/threads/{id}/turns", h.handleSubmitTurn)
	if staticFS != nil {
		mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))
	}
}

func render(w http.ResponseWriter, tmpl *template.Template, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("web: render %s: %v", name, err)
		http.Error(w, "template error", 500)
	}
}

func (h *Handlers) renderPartial(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmplName := name
	if i := strings.LastIndex(name, "/"); i >= 0 {
		tmplName = name[i+1:]
	}
	tmplName = strings.TrimSuffix(tmplName, ".html")
	if err := h.parts.ExecuteTemplate(w, tmplName, data); err != nil {
		log.Printf("web: render partial %s: %v", name, err)
		http.Error(w, "template error", 500)
	}
}
