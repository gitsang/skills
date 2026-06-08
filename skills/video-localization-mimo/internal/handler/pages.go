package handler

import (
	"html/template"
	"log"
	"net/http"
	"sort"

	"github.com/gitsang/skills/video-localization-mimo/internal/config"
	"github.com/gitsang/skills/video-localization-mimo/internal/task"
)

type Handler struct {
	taskStore task.Store
	config    *config.Config
	templates *template.Template
}

func NewHandler(store task.Store, cfg *config.Config, tmpl *template.Template) *Handler {
	return &Handler{
		taskStore: store,
		config:    cfg,
		templates: tmpl,
	}
}

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.taskStore.List()
	if err != nil {
		log.Printf("Error listing tasks: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.After(tasks[j].CreatedAt)
	})

	data := struct {
		Tasks  []*task.Task
		Config *config.Config
	}{
		Tasks:  tasks,
		Config: h.config,
	}

	if err := h.templates.ExecuteTemplate(w, "home.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) TaskHandler(w http.ResponseWriter, r *http.Request) {
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
		Task   *task.Task
		Config *config.Config
	}{
		Task:   t,
		Config: h.config,
	}

	if err := h.templates.ExecuteTemplate(w, "task.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) SettingsHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Config *config.Config
	}{
		Config: h.config,
	}

	if err := h.templates.ExecuteTemplate(w, "settings.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
