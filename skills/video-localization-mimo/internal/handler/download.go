package handler

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gitsang/skills/video-localization-mimo/internal/task"
)

func (h *Handler) DownloadHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	fileName := r.PathValue("file")

	if id == "" || fileName == "" {
		http.Error(w, "Task ID and file name required", http.StatusBadRequest)
		return
	}

	t, err := h.taskStore.Get(id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	filePath := filepath.Join(t.OutputDir, fileName)

	cleanPath := filepath.Clean(filePath)
	if filepath.Dir(cleanPath) != filepath.Clean(t.OutputDir) {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	ext := filepath.Ext(fileName)
	contentType := getContentType(ext)

	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, cleanPath)
}

func (h *Handler) DownloadPageHandler(w http.ResponseWriter, r *http.Request) {
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

	artifacts := getArtifacts(t)

	data := struct {
		Task      *task.Task
		Artifacts []Artifact
	}{
		Task:      t,
		Artifacts: artifacts,
	}

	w.Header().Set("Content-Type", "text/html")
	if err := h.templates.ExecuteTemplate(w, "artifacts", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

type Artifact struct {
	Name string
	Path string
	Size int64
}

func getArtifacts(t *task.Task) []Artifact {
	var artifacts []Artifact

	entries, err := os.ReadDir(t.OutputDir)
	if err != nil {
		return artifacts
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		artifacts = append(artifacts, Artifact{
			Name: entry.Name(),
			Path: entry.Name(),
			Size: info.Size(),
		})
	}

	return artifacts
}

func getContentType(ext string) string {
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".wav":
		return "audio/wav"
	case ".mp3":
		return "audio/mpeg"
	case ".srt":
		return "text/plain"
	case ".ass":
		return "text/plain"
	case ".txt":
		return "text/plain"
	case ".json":
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
