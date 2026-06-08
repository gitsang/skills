package srt

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []Segment
		wantErr bool
	}{
		{
			name: "valid SRT with single segment",
			input: `1
00:00:01,000 --> 00:00:04,000
Hello, this is the first subtitle.`,
			want: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "Hello, this is the first subtitle.",
				},
			},
		},
		{
			name: "valid SRT with multiple segments",
			input: `1
00:00:01,000 --> 00:00:04,000
Hello, this is the first subtitle.

2
00:00:05,000 --> 00:00:08,000
This is the second subtitle.
It can have multiple lines.`,
			want: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "Hello, this is the first subtitle.",
				},
				{
					Index: 2,
					Start: time.Date(0, 1, 1, 0, 0, 5, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 8, 0, time.UTC),
					Text:  "This is the second subtitle.\nIt can have multiple lines.",
				},
			},
		},
		{
			name: "SRT with dot separator",
			input: `1
00:00:01.000 --> 00:00:04.000
Hello, this is the first subtitle.`,
			want: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "Hello, this is the first subtitle.",
				},
			},
		},
		{
			name: "SRT with millisecond precision",
			input: `1
00:00:01,500 --> 00:00:04,750
Hello, this is the first subtitle.`,
			want: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 500000000, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 750000000, time.UTC),
					Text:  "Hello, this is the first subtitle.",
				},
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantErr: true,
		},
		{
			name: "missing text",
			input: `1
00:00:01,000 --> 00:00:04,000`,
			wantErr: true,
		},
		{
			name: "invalid index",
			input: `abc
00:00:01,000 --> 00:00:04,000
Hello`,
			wantErr: true,
		},
		{
			name: "invalid timestamp format",
			input: `1
00:00:01,000
Hello`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("Parse() got %d segments, want %d", len(got), len(tt.want))
					return
				}
				for i, seg := range got {
					if seg.Index != tt.want[i].Index {
						t.Errorf("Parse() segment %d index = %d, want %d", i, seg.Index, tt.want[i].Index)
					}
					if !seg.Start.Equal(tt.want[i].Start) {
						t.Errorf("Parse() segment %d start = %v, want %v", i, seg.Start, tt.want[i].Start)
					}
					if !seg.End.Equal(tt.want[i].End) {
						t.Errorf("Parse() segment %d end = %v, want %v", i, seg.End, tt.want[i].End)
					}
					if seg.Text != tt.want[i].Text {
						t.Errorf("Parse() segment %d text = %q, want %q", i, seg.Text, tt.want[i].Text)
					}
				}
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		segments []Segment
		want     string
		wantErr  bool
	}{
		{
			name: "single segment",
			segments: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "Hello, this is the first subtitle.",
				},
			},
			want: "1\n00:00:01,000 --> 00:00:04,000\nHello, this is the first subtitle.\n",
		},
		{
			name: "multiple segments",
			segments: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "Hello, this is the first subtitle.",
				},
				{
					Index: 2,
					Start: time.Date(0, 1, 1, 0, 0, 5, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 8, 0, time.UTC),
					Text:  "This is the second subtitle.\nIt can have multiple lines.",
				},
			},
			want: "1\n00:00:01,000 --> 00:00:04,000\nHello, this is the first subtitle.\n\n2\n00:00:05,000 --> 00:00:08,000\nThis is the second subtitle.\nIt can have multiple lines.\n",
		},
		{
			name:     "empty segments",
			segments: []Segment{},
			want:     "",
		},
		{
			name: "invalid segment - negative index",
			segments: []Segment{
				{
					Index: -1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "Hello",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid segment - empty text",
			segments: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Generate(&buf, tt.segments)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got := buf.String(); got != tt.want {
					t.Errorf("Generate() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		segments []Segment
		wantErr  bool
	}{
		{
			name: "valid segments",
			segments: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "Hello",
				},
				{
					Index: 2,
					Start: time.Date(0, 1, 1, 0, 0, 5, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 8, 0, time.UTC),
					Text:  "World",
				},
			},
		},
		{
			name:     "empty segments",
			segments: []Segment{},
			wantErr:  true,
		},
		{
			name: "non-sequential index",
			segments: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "Hello",
				},
				{
					Index: 3,
					Start: time.Date(0, 1, 1, 0, 0, 5, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 8, 0, time.UTC),
					Text:  "World",
				},
			},
			wantErr: true,
		},
		{
			name: "overlapping timestamps",
			segments: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 6, 0, time.UTC),
					Text:  "Hello",
				},
				{
					Index: 2,
					Start: time.Date(0, 1, 1, 0, 0, 5, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 8, 0, time.UTC),
					Text:  "World",
				},
			},
			wantErr: true,
		},
		{
			name: "end before start",
			segments: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					Text:  "Hello",
				},
			},
			wantErr: true,
		},
		{
			name: "empty text",
			segments: []Segment{
				{
					Index: 1,
					Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
					End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
					Text:  "",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.segments)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	input := `1
00:00:01,000 --> 00:00:04,000
Hello, this is the first subtitle.

2
00:00:05,000 --> 00:00:08,000
This is the second subtitle.
It can have multiple lines.
`

	segments, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	var buf bytes.Buffer
	err = Generate(&buf, segments)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	segments2, err := Parse(strings.NewReader(buf.String()))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(segments) != len(segments2) {
		t.Fatalf("Round trip failed: got %d segments, want %d", len(segments2), len(segments))
	}

	for i, seg := range segments {
		seg2 := segments2[i]
		if seg.Index != seg2.Index {
			t.Errorf("Round trip segment %d index = %d, want %d", i, seg2.Index, seg.Index)
		}
		if !seg.Start.Equal(seg2.Start) {
			t.Errorf("Round trip segment %d start = %v, want %v", i, seg2.Start, seg.Start)
		}
		if !seg.End.Equal(seg2.End) {
			t.Errorf("Round trip segment %d end = %v, want %v", i, seg2.End, seg.End)
		}
		if seg.Text != seg2.Text {
			t.Errorf("Round trip segment %d text = %q, want %q", i, seg2.Text, seg.Text)
		}
	}
}

func TestSegmentDuration(t *testing.T) {
	segment := Segment{
		Index: 1,
		Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
		End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
		Text:  "Hello",
	}

	want := 3 * time.Second
	got := segment.Duration()
	if got != want {
		t.Errorf("Duration() = %v, want %v", got, want)
	}
}

func TestSegmentString(t *testing.T) {
	segment := Segment{
		Index: 1,
		Start: time.Date(0, 1, 1, 0, 0, 1, 0, time.UTC),
		End:   time.Date(0, 1, 1, 0, 0, 4, 0, time.UTC),
		Text:  "Hello",
	}

	got := segment.String()
	want := "[1] 00:00:01,000 --> 00:00:04,000\nHello"
	if got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
