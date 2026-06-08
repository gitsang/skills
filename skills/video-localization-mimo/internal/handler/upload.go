package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gitsang/skills/video-localization-mimo/internal/task"
)

func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Invalid file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	taskID := fmt.Sprintf("%d", time.Now().UnixNano())
	uploadDir := filepath.Join("uploads", taskID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Printf("Error creating upload directory: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	uploadPath := filepath.Join(uploadDir, header.Filename)
	dst, err := os.Create(uploadPath)
	if err != nil {
		log.Printf("Error creating file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		log.Printf("Error saving file: %v", err)
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	sourceLang := r.FormValue("source_lang")
	targetLang := r.FormValue("target_lang")
	voice := r.FormValue("voice")
	speedStr := r.FormValue("speed")

	if sourceLang == "" {
		sourceLang = h.config.Defaults.SourceLang
	}
	if targetLang == "" {
		targetLang = h.config.Defaults.TargetLang
	}
	if voice == "" {
		voice = h.config.Defaults.Voice
	}

	speed := 1.0
	if speedStr != "" {
		fmt.Sscanf(speedStr, "%f", &speed)
	}

	outputDir := filepath.Join("outputs", taskID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	t := &task.Task{
		ID:          taskID,
		SourceVideo: uploadPath,
		SourceLang:  sourceLang,
		TargetLang:  targetLang,
		OutputDir:   outputDir,
		Voice:       voice,
		Speed:       speed,
	}

	if err := h.taskStore.Create(t); err != nil {
		log.Printf("Error creating task: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/task/%s", taskID))
	w.WriteHeader(http.StatusSeeOther)
}
