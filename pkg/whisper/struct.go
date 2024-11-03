package whisper

type WhisperInput struct {
	FilePath string `json:"filepath"`
	//Reader   io.ReadSeeker `json:"-"`

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
	Num   int     `json:"num"`   // Segment Number
	Start float64 `json:"start"` // Start is the start of the segment.
	End   float64 `json:"end"`   // End is the end of the segment.
	Text  string  `json:"text"`  // Text is the text of the segment.
}

type WhisperOutput struct {
	Segments []Segment

	// seconds
	Duration float64
}
