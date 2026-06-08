package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/internal/task"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ProgressMessage struct {
	Type      string  `json:"type"`
	TaskID    string  `json:"task_id"`
	Step      string  `json:"step"`
	Progress  float64 `json:"progress"`
	Message   string  `json:"message"`
	Timestamp string  `json:"timestamp"`
}

func (h *Handler) WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	t, err := h.taskStore.Get(id)
	if err != nil {
		conn.WriteJSON(map[string]string{"error": "Task not found"})
		return
	}

	initialMsg := ProgressMessage{
		Type:      "status",
		TaskID:    t.ID,
		Step:      t.CurrentStep,
		Progress:  t.Progress,
		Message:   string(t.Status),
		Timestamp: time.Now().Format(time.RFC3339),
	}
	if err := conn.WriteJSON(initialMsg); err != nil {
		log.Printf("WebSocket write error: %v", err)
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		t, err := h.taskStore.Get(id)
		if err != nil {
			break
		}

		msg := ProgressMessage{
			Type:      "progress",
			TaskID:    t.ID,
			Step:      t.CurrentStep,
			Progress:  t.Progress,
			Timestamp: time.Now().Format(time.RFC3339),
		}

		if t.IsTerminal() {
			msg.Type = "complete"
			msg.Message = string(t.Status)
			conn.WriteJSON(msg)
			break
		}

		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}

func (h *Handler) TaskListPartialHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.taskStore.List()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "task-list", tasks); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) TaskStepsPartialHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	t, err := h.taskStore.Get(id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "step-list", t); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) TaskLogsPartialHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	t, err := h.taskStore.Get(id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	data := struct {
		Task *task.Task
		Logs []string
	}{
		Task: t,
		Logs: []string{},
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "logs", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
