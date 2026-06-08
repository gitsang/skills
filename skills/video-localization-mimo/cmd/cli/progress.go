package cli

import (
	"fmt"
	"strings"
	"time"
)

// ProgressBar displays a text-based progress bar in the terminal.
type ProgressBar struct {
	total     int
	current   int
	width     int
	prefix    string
	startTime time.Time
	done      chan struct{}
}

// NewProgressBar creates a new progress bar with the given total and prefix.
func NewProgressBar(total int, prefix string) *ProgressBar {
	return &ProgressBar{
		total:  total,
		width:  40,
		prefix: prefix,
		done:   make(chan struct{}),
	}
}

// Start begins the progress bar display loop.
func (p *ProgressBar) Start() {
	p.startTime = time.Now()
	p.render()
}

// Update sets the current progress value and re-renders.
func (p *ProgressBar) Update(current int) {
	p.current = current
	p.render()
}

// Finish marks the progress bar as complete and prints a newline.
func (p *ProgressBar) Finish() {
	p.current = p.total
	p.render()
	fmt.Println()
}

func (p *ProgressBar) render() {
	if p.total <= 0 {
		fmt.Printf("\r%s ...", p.prefix)
		return
	}

	percent := float64(p.current) / float64(p.total)
	if percent > 1.0 {
		percent = 1.0
	}

	filled := int(float64(p.width) * percent)
	empty := p.width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	elapsed := time.Since(p.startTime)
	elapsedStr := formatDuration(elapsed)

	fmt.Printf("\r%s [%s] %d/%d (%s)", p.prefix, bar, p.current, p.total, elapsedStr)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}
