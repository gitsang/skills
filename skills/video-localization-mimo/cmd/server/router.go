package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gitsang/skills/video-localization-mimo/internal/config"
	"github.com/gitsang/skills/video-localization-mimo/internal/handler"
	"github.com/gitsang/skills/video-localization-mimo/internal/middleware"
	"github.com/gitsang/skills/video-localization-mimo/internal/task"
)

func SetupRouter(store task.Store, cfg *config.Config) http.Handler {
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 {
			return a * b
		},
		"formatSize": func(bytes int64) string {
			const unit = 1024
			if bytes < unit {
				return fmt.Sprintf("%d B", bytes)
			}
			div, exp := int64(unit), 0
			for n := bytes / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
		},
	}

	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("web/templates/**/*.html"))

	h := handler.NewHandler(store, cfg, tmpl)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", h.HomeHandler)
	mux.HandleFunc("GET /task/{id}", h.TaskHandler)
	mux.HandleFunc("GET /settings", h.SettingsHandler)

	mux.HandleFunc("POST /api/tasks", h.CreateTaskHandler)
	mux.HandleFunc("GET /api/tasks", h.ListTasksHandler)
	mux.HandleFunc("GET /api/tasks/{id}", h.GetTaskHandler)
	mux.HandleFunc("DELETE /api/tasks/{id}", h.DeleteTaskHandler)
	mux.HandleFunc("GET /api/tasks/{id}/progress", h.GetProgressHandler)
	mux.HandleFunc("GET /api/tasks/{id}/steps", h.TaskStepsPartialHandler)
	mux.HandleFunc("GET /api/tasks/{id}/logs", h.TaskLogsPartialHandler)
	mux.HandleFunc("GET /api/tasks/{id}/artifacts", h.DownloadPageHandler)

	mux.HandleFunc("POST /upload", h.UploadHandler)

	mux.HandleFunc("GET /ws/{id}", h.WebSocketHandler)

	mux.HandleFunc("GET /download/{id}/{file}", h.DownloadHandler)

	staticDir := "web/static"
	if _, err := os.Stat(staticDir); err == nil {
		fileServer := http.FileServer(http.Dir(staticDir))
		mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	}

	var handler http.Handler = mux
	handler = middleware.Logger(handler)
	handler = middleware.CORS(handler)

	return handler
}

func mustGlob(pattern string) []string {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		panic(fmt.Sprintf("glob %s: %v", pattern, err))
	}
	return matches
}

func parseTemplates(funcMap template.FuncMap, patterns ...string) *template.Template {
	tmpl := template.New("").Funcs(funcMap)
	for _, pattern := range patterns {
		matches := mustGlob(pattern)
		for _, match := range matches {
			content, err := os.ReadFile(match)
			if err != nil {
				panic(fmt.Sprintf("read %s: %v", match, err))
			}
			name := strings.TrimPrefix(match, "web/templates/")
			tmpl = template.Must(tmpl.New(name).Parse(string(content)))
		}
	}
	return tmpl
}
