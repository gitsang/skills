package srt

import (
	"fmt"
	"io"
	"os"
)

// Generate writes the provided segments to the writer in SRT format.
func Generate(writer io.Writer, segments []Segment) error {
	for i, segment := range segments {
		if err := validateSegment(segment); err != nil {
			return fmt.Errorf("segment %d: %w", i+1, err)
		}

		if _, err := fmt.Fprintf(writer, "%d\n", segment.Index); err != nil {
			return fmt.Errorf("failed to write segment index: %w", err)
		}

		timestampLine := formatTimestamps(segment.Start, segment.End)
		if _, err := fmt.Fprintf(writer, "%s\n", timestampLine); err != nil {
			return fmt.Errorf("failed to write timestamps: %w", err)
		}

		if _, err := fmt.Fprintf(writer, "%s\n", segment.Text); err != nil {
			return fmt.Errorf("failed to write text: %w", err)
		}

		if i < len(segments)-1 {
			if _, err := fmt.Fprintln(writer); err != nil {
				return fmt.Errorf("failed to write separator: %w", err)
			}
		}
	}

	return nil
}

// GenerateFile writes the provided segments to a file in SRT format.
func GenerateFile(path string, segments []Segment) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create SRT file %s: %w", path, err)
	}
	defer file.Close()

	return Generate(file, segments)
}

func formatTimestamps(start, end interface{ Format(string) string }) string {
	startStr := start.Format(TimeFormatComma)
	endStr := end.Format(TimeFormatComma)
	return fmt.Sprintf("%s --> %s", startStr, endStr)
}

func validateSegment(segment Segment) error {
	if segment.Index <= 0 {
		return fmt.Errorf("index must be positive, got %d", segment.Index)
	}
	if segment.Text == "" {
		return fmt.Errorf("text cannot be empty")
	}
	return nil
}
