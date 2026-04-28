package task

import (
	"fmt"
	"time"
)

type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusCompleted TaskStatus = "completed"
	StatusFailed    TaskStatus = "failed"
	StatusCancelled TaskStatus = "cancelled"
)

type StepInfo struct {
	Name        string     `json:"name"`
	Status      TaskStatus `json:"status"`
	Progress    float64    `json:"progress"`
	Error       string     `json:"error,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type Task struct {
	ID          string     `json:"id"`
	Status      TaskStatus `json:"status"`
	SourceVideo string     `json:"source_video"`
	SourceLang  string     `json:"source_lang"`
	TargetLang  string     `json:"target_lang"`
	OutputDir   string     `json:"output_dir"`
	Voice       string     `json:"voice,omitempty"`
	Speed       float64    `json:"speed,omitempty"`
	CloneRef    string     `json:"clone_ref,omitempty"`
	Progress    float64    `json:"progress"`
	CurrentStep string     `json:"current_step,omitempty"`
	Steps       []StepInfo `json:"steps"`
	Error       string     `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

func (t *Task) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("task ID is required")
	}
	if t.SourceVideo == "" {
		return fmt.Errorf("source video is required")
	}
	if t.TargetLang == "" {
		return fmt.Errorf("target language is required")
	}
	return nil
}

func (t *Task) IsTerminal() bool {
	return t.Status == StatusCompleted || t.Status == StatusFailed || t.Status == StatusCancelled
}

func (t *Task) Duration() time.Duration {
	if t.CompletedAt != nil {
		return t.CompletedAt.Sub(t.CreatedAt)
	}
	if t.Status == StatusRunning {
		return time.Since(t.CreatedAt)
	}
	return 0
}
