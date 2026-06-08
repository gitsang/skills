package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/internal/task"
)

func (h *Handler) CreateTaskHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SourceVideo string  `json:"source_video"`
		SourceLang  string  `json:"source_lang"`
		TargetLang  string  `json:"target_lang"`
		Voice       string  `json:"voice"`
		Speed       float64 `json:"speed"`
		CloneRef    string  `json:"clone_ref"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	t := &task.Task{
		ID:          generateID(),
		SourceVideo: req.SourceVideo,
		SourceLang:  req.SourceLang,
		TargetLang:  req.TargetLang,
		Voice:       req.Voice,
		Speed:       req.Speed,
		CloneRef:    req.CloneRef,
	}

	if t.SourceLang == "" {
		t.SourceLang = h.config.Defaults.SourceLang
	}
	if t.TargetLang == "" {
		t.TargetLang = h.config.Defaults.TargetLang
	}
	if t.Voice == "" {
		t.Voice = h.config.Defaults.Voice
	}
	if t.Speed == 0 {
		t.Speed = 1.0
	}

	if err := h.taskStore.Create(t); err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) GetTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

func (h *Handler) ListTasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.taskStore.List()
	if err != nil {
		log.Printf("Error listing tasks: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func (h *Handler) DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	if err := h.taskStore.Delete(id); err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetProgressHandler(w http.ResponseWriter, r *http.Request) {
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
	}{
		Task: t,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "progress", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
