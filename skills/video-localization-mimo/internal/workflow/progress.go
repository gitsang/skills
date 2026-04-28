package workflow

import (
	"fmt"
	"strings"
	"time"
)

type ProgressReporter interface {
	OnStart(totalSteps int)
	OnStepStart(index int, name string)
	OnStepComplete(index int, result StepResult)
	OnFinish(results []StepResult, success bool)
}

type NoOpProgressReporter struct{}

func (r *NoOpProgressReporter) OnStart(totalSteps int)                              {}
func (r *NoOpProgressReporter) OnStepStart(index int, name string)                  {}
func (r *NoOpProgressReporter) OnStepComplete(index int, result StepResult)         {}
func (r *NoOpProgressReporter) OnFinish(results []StepResult, success bool)         {}

type CLIProgressReporter struct {
	total     int
	current   int
	width     int
	startTime time.Time
}

func NewCLIProgressReporter() *CLIProgressReporter {
	return &CLIProgressReporter{width: 40}
}

func (r *CLIProgressReporter) OnStart(totalSteps int) {
	r.total = totalSteps
	r.startTime = time.Now()
	r.render("准备中...")
}

func (r *CLIProgressReporter) OnStepStart(index int, name string) {
	r.current = index + 1
	r.render(fmt.Sprintf("%s...", name))
}

func (r *CLIProgressReporter) OnStepComplete(index int, result StepResult) {
	if result.Status == StepCompleted {
		r.render(fmt.Sprintf("✓ %s (%s)", result.Name, formatDur(result.Duration)))
		fmt.Println()
	} else {
		r.render(fmt.Sprintf("✗ %s: %s", result.Name, result.Error))
		fmt.Println()
	}
}

func (r *CLIProgressReporter) OnFinish(results []StepResult, success bool) {
	totalDur := time.Since(r.startTime)

	if success {
		fmt.Printf("\n✅ 本地化完成! 总耗时: %s\n", formatDur(totalDur))
	} else {
		fmt.Printf("\n❌ 本地化失败 (耗时: %s)\n", formatDur(totalDur))
	}

	fmt.Println(strings.Repeat("─", 60))
	for _, result := range results {
		status := "✓"
		if result.Status == StepFailed {
			status = "✗"
		}
		fmt.Printf("  %s %-12s %s\n", status, result.Name, formatDur(result.Duration))
		if result.Error != "" {
			fmt.Printf("    └─ %s\n", result.Error)
		}
	}
	fmt.Println(strings.Repeat("─", 60))
}

func (r *CLIProgressReporter) render(msg string) {
	percent := float64(r.current) / float64(r.total)
	if percent > 1.0 {
		percent = 1.0
	}
	if r.total == 0 {
		percent = 0
	}

	filled := int(float64(r.width) * percent)
	empty := r.width - filled
	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	elapsed := time.Since(r.startTime)
	fmt.Printf("\r  [%s] %d/%d %s (%s)", bar, r.current, r.total, msg, formatDur(elapsed))
}

func formatDur(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
