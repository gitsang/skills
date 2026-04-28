// Package srt provides functionality for parsing, generating, and validating SRT subtitle files.
// SRT (SubRip Text) is the most common subtitle format, consisting of indexed segments
// with timestamps and text content.
package srt

import (
	"fmt"
	"time"
)

// TimeFormat represents the SRT timestamp format: HH:MM:SS,mmm
const TimeFormat = "15:04:05.000"

// TimeFormatComma represents the SRT timestamp format with comma as millisecond separator
const TimeFormatComma = "15:04:05,000"

// Segment represents a single subtitle segment in an SRT file.
// Each segment contains an index, start/end timestamps, and text content.
type Segment struct {
	// Index is the sequential number of this segment in the SRT file (1-based)
	Index int

	// Start is the timestamp when this subtitle should appear
	Start time.Time

	// End is the timestamp when this subtitle should disappear
	End time.Time

	// Text contains the subtitle content, which may span multiple lines
	Text string
}

// Duration returns the duration of this subtitle segment.
func (s Segment) Duration() time.Duration {
	return s.End.Sub(s.Start)
}

// Validate checks if this segment has valid data.
func (s Segment) Validate() error {
	if s.Index <= 0 {
		return fmt.Errorf("segment index must be positive, got %d", s.Index)
	}

	if s.End.Before(s.Start) || s.End.Equal(s.Start) {
		return fmt.Errorf("segment %d: end time (%v) must be after start time (%v)",
			s.Index, s.End, s.Start)
	}

	if s.Text == "" {
		return fmt.Errorf("segment %d: text cannot be empty", s.Index)
	}

	return nil
}

// String returns a human-readable representation of the segment.
func (s Segment) String() string {
	return fmt.Sprintf("[%d] %s --> %s\n%s",
		s.Index,
		s.Start.Format(TimeFormatComma),
		s.End.Format(TimeFormatComma),
		s.Text)
}
