package whisper

import (
	"fmt"
	"io"
)

type WhisperInput struct {
	FilePath string        `json:"filepath"`
	Reader   io.ReadSeeker `json:"-"`

	Prompt      string  `json:"prompt"`
	Temperature float32 `json:"temperature"`
	Language    string  `json:"language"`
}

// func (input *WhisperInput) GetInputReader(inputFilePath string) (io.ReadSeekCloser, error) {
// 	if input.Reader != nil {
// 		return input.Reader, nil
// 	}

// 	file, err := os.Open(input.FilePath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// defer file.Close()

// 	return file, nil
// }

type Segment struct {
	Num   int     // Segment Number
	Start float64 // Start is the start of the segment.
	End   float64 // End is the end of the segment.
	Text  string  // Text is the text of the segment.
}

type Segments []Segment

// Return srtTimestamp
func formatTimestamp(seconds float64) string {
	h := int(seconds) / 3600
	m := (int(seconds) % 3600) / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

type WhisperOutput interface {
	GetSegmentList() Segments
	// return second 
	GetDuration() float64
}
