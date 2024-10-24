package local

import (
	"m3u8dl_for_web/pkg/whisper"

	whispercpp "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

var _ whisper.WhisperOutput = &LocalWhisperOutput{}

type LocalWhisperOutput struct {
	SegmentList []whispercpp.Segment
}

func (output *LocalWhisperOutput) GetSegmentList() whisper.Segments {
	result := make(whisper.Segments, len(output.SegmentList))

	for i, item := range output.SegmentList {
		result[i] = whisper.Segment{
			Num:   item.Num,
			Start: item.Start.Seconds(),
			End:   item.End.Seconds(),
			Text:  item.Text,
		}
	}

	return result
}
