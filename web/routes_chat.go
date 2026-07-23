package web

import (
	"fmt"
	"html"
	"net/http"
	"time"

	"github.com/alfred/alfred/internal/contract"
)

// ChatPageData is passed to the chat page layout template.
type ChatPageData struct {
	ThreadID    string
	ThreadTitle string
}

// SidebarData is passed to the sidebar partial.
type SidebarData struct {
	Threads []SidebarThread
}

// SidebarThread is a sidebar row.
type SidebarThread struct {
	ID     string
	Title  string
	Active bool
}

// TimelineData is passed to the timeline partial.
type TimelineData struct {
	ThreadID string
	Messages []TimelineMessage
}

// TimelineMessage is a single message in the timeline.
type TimelineMessage struct {
	Kind      string
	Text      string
	ToolName  string
	ToolInput string
}

// handleChat renders the full chat page.
func (h *Handlers) handleChat(w http.ResponseWriter, r *http.Request) {
	threadID := r.URL.Query().Get("thread")
	data := ChatPageData{ThreadID: threadID}

	if threadID == "" && h.Threads != nil {
		threads, _, err := h.Threads.List(1, "")
		if err == nil && len(threads) > 0 {
			data.ThreadID = string(threads[0].ID)
			data.ThreadTitle = threads[0].Title
		}
	}

	render(w, h.chat, "layout", data)
}

// handleSidebar renders the sidebar partial.
func (h *Handlers) handleSidebar(w http.ResponseWriter, r *http.Request) {
	data := SidebarData{}

	if h.Threads != nil {
		threads, _, err := h.Threads.List(50, "")
		if err == nil {
			for _, t := range threads {
				data.Threads = append(data.Threads, SidebarThread{
					ID:    string(t.ID),
					Title: t.Title,
				})
			}
		}
	}

	h.renderPartial(w, "sidebar.html", data)
}

// handleTimeline renders the timeline partial for a thread.
func (h *Handlers) handleTimeline(w http.ResponseWriter, r *http.Request) {
	id := contract.ThreadID(r.PathValue("id"))
	data := TimelineData{ThreadID: string(id)}

	if h.Session != nil {
		items, err := h.Session.Read(id, 0)
		if err == nil {
			for _, item := range items {
				msg := TimelineMessage{Kind: string(item.Kind)}
				if item.Text != nil {
					msg.Text = *item.Text
				}
				if item.ToolCall != nil {
					msg.ToolName = item.ToolCall.Name
					msg.ToolInput = string(item.ToolCall.Input)
				}
				data.Messages = append(data.Messages, msg)
			}
		}
	}

	h.renderPartial(w, "timeline.html", data)
}

// handleTodos renders the todo panel partial (placeholder — todos not yet in store).
func (h *Handlers) handleTodos(w http.ResponseWriter, r *http.Request) {
	// ponytail: todos are in-memory GoalStore, not wired yet. return empty.
	h.renderPartial(w, "todo_panel.html", nil)
}

// handleCreateThread creates a thread and returns the refreshed sidebar.
func (h *Handlers) handleCreateThread(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		title = fmt.Sprintf("Thread %s", time.Now().Format("3:04 PM"))
	}

	th := &contract.Thread{
		ID:        contract.ThreadID(fmt.Sprintf("thr-%d", time.Now().UnixNano())),
		Title:     title,
		Status:    contract.ThreadStatusIdle,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if h.Create != nil {
		if err := h.Create.Create(th); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	h.handleSidebar(w, r)
}

// handleSubmitTurn starts a turn and returns a thinking bubble HTML fragment.
func (h *Handlers) handleSubmitTurn(w http.ResponseWriter, r *http.Request) {
	id := contract.ThreadID(r.PathValue("id"))
	text := r.FormValue("text")
	if text == "" {
		http.Error(w, "empty message", 400)
		return
	}

	var th *contract.Thread
	if h.Thread != nil {
		var err error
		th, err = h.Thread.Get(id)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}
	}

	if h.Starter != nil && th != nil {
		_, _ = h.Starter.StartTurn(th, contract.UserInput{Text: text})
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	safe := html.EscapeString(text)
	_, _ = fmt.Fprintf(w, `<div class="ds-bubble-user">%s</div>`, safe)
	_, _ = fmt.Fprintf(w, `<div class="ds-bubble-assistant"><div class="ds-thinking"><span class="ds-thinking-dot"></span><span class="ds-thinking-dot"></span><span class="ds-thinking-dot"></span> thinking...</div></div>`)
}
