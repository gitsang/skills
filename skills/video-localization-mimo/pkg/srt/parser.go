package srt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// Parse reads SRT content from the provided reader and returns a slice of segments.
func Parse(reader io.Reader) ([]Segment, error) {
	scanner := bufio.NewScanner(reader)
	var segments []Segment
	var currentSegment *Segment
	var textLines []string
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			if currentSegment != nil {
				currentSegment.Text = strings.Join(textLines, "\n")
				segments = append(segments, *currentSegment)
				currentSegment = nil
				textLines = nil
			}
			continue
		}

		if currentSegment == nil {
			index, err := parseIndex(line, lineNum)
			if err != nil {
				return nil, err
			}
			currentSegment = &Segment{Index: index}
			continue
		}

		if currentSegment.Start.IsZero() {
			start, end, err := parseTimestamps(line, lineNum)
			if err != nil {
				return nil, err
			}
			currentSegment.Start = start
			currentSegment.End = end
			continue
		}

		textLines = append(textLines, line)
	}

	if currentSegment != nil {
		if len(textLines) == 0 {
			return nil, fmt.Errorf("segment %d: missing text content", currentSegment.Index)
		}
		currentSegment.Text = strings.Join(textLines, "\n")
		segments = append(segments, *currentSegment)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading SRT content: %w", err)
	}

	if len(segments) == 0 {
		return nil, fmt.Errorf("no segments found in SRT content")
	}

	return segments, nil
}

// ParseFile reads an SRT file from the given path and returns a slice of segments.
func ParseFile(path string) ([]Segment, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SRT file %s: %w", path, err)
	}
	defer file.Close()

	return Parse(file)
}

func parseIndex(line string, lineNum int) (int, error) {
	var index int
	if _, err := fmt.Sscanf(line, "%d", &index); err != nil {
		return 0, fmt.Errorf("line %d: invalid segment index %q: %w", lineNum, line, err)
	}
	if index <= 0 {
		return 0, fmt.Errorf("line %d: segment index must be positive, got %d", lineNum, index)
	}
	return index, nil
}

func parseTimestamps(line string, lineNum int) (start, end time.Time, err error) {
	parts := strings.Split(line, "-->")
	if len(parts) != 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("line %d: invalid timestamp format %q: expected 'HH:MM:SS,mmm --> HH:MM:SS,mmm'", lineNum, line)
	}

	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	start, err = parseTimestamp(startStr, lineNum)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	end, err = parseTimestamp(endStr, lineNum)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return start, end, nil
}

func parseTimestamp(s string, lineNum int) (time.Time, error) {
	s = strings.Replace(s, ",", ".", 1)

	t, err := time.Parse(TimeFormat, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("line %d: invalid timestamp %q: %w", lineNum, s, err)
	}

	return t, nil
}
