package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Report struct {
	Steps         []StepResult  `json:"steps"`
	TotalDuration time.Duration `json:"total_duration"`
	Success       bool          `json:"success"`
}

func GenerateReport(results []StepResult, success bool) Report {
	var total time.Duration
	for _, r := range results {
		total += r.Duration
	}
	return Report{
		Steps:         results,
		TotalDuration: total,
		Success:       success,
	}
}

func SaveReport(path string, report Report) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
