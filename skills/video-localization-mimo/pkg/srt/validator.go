package srt

import (
	"fmt"
)

// Validate checks the provided segments for correctness and returns an error if any issues are found.
func Validate(segments []Segment) error {
	if len(segments) == 0 {
		return fmt.Errorf("no segments provided")
	}

	for i, segment := range segments {
		if err := validateSegmentIndex(segment, i); err != nil {
			return err
		}

		if err := validateSegmentTimestamps(segment); err != nil {
			return err
		}

		if err := validateSegmentText(segment); err != nil {
			return err
		}

		if i > 0 {
			if err := validateSegmentOrder(segments[i-1], segment); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateSegmentIndex(segment Segment, expectedIndex int) error {
	if segment.Index != expectedIndex+1 {
		return fmt.Errorf("segment %d: index should be %d, got %d", expectedIndex+1, expectedIndex+1, segment.Index)
	}
	return nil
}

func validateSegmentTimestamps(segment Segment) error {
	if segment.Start.IsZero() {
		return fmt.Errorf("segment %d: start time is missing", segment.Index)
	}

	if segment.End.IsZero() {
		return fmt.Errorf("segment %d: end time is missing", segment.Index)
	}

	if segment.End.Before(segment.Start) || segment.End.Equal(segment.Start) {
		return fmt.Errorf("segment %d: end time (%v) must be after start time (%v)",
			segment.Index, segment.End, segment.Start)
	}

	return nil
}

func validateSegmentText(segment Segment) error {
	if segment.Text == "" {
		return fmt.Errorf("segment %d: text cannot be empty", segment.Index)
	}
	return nil
}

func validateSegmentOrder(prev, current Segment) error {
	if current.Start.Before(prev.End) {
		return fmt.Errorf("segment %d: start time (%v) overlaps with segment %d end time (%v)",
			current.Index, current.Start, prev.Index, prev.End)
	}
	return nil
}
