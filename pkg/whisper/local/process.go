package local

import (
	"context"
	"fmt"
	"io"
	"os"

	"m3u8dl_for_web/pkg/whisper"

	whispercpp "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	wav "github.com/go-audio/wav"
)

var _ whisper.WhisperHandler = &LocalWhisper{}

type LocalWhisper struct {
	modelPath string
}

func NewLocalWhisper(modelPath string) *LocalWhisper {
	return &LocalWhisper{
		modelPath: modelPath,
	}
}

func (localWhisper *LocalWhisper) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (whisper.WhisperOutput, error) {
	model, err := whispercpp.New(localWhisper.modelPath)
	if err != nil {
		return nil, err
	}
	defer model.Close()

	context, err := model.NewContext()
	if err != nil {
		return nil, err
	}

	reader := input.Reader
	if reader == nil {
		file, err := os.Open(input.FilePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	data, err := localWhisper.readPCM(reader)
	if err != nil {
		return nil, err
	}

	var (
		segments = make([]whispercpp.Segment, 0)
		cb       = func(segment whispercpp.Segment) {
			segments = append(segments, segment)
		}
	)

	context.ResetTimings()
	if err := context.Process(data, cb, nil); err != nil {
		return nil, err
	}

	context.PrintTimings()

	return &LocalWhisperOutput{SegmentList: segments}, nil
}

func (localWhisper *LocalWhisper) readPCM(reader io.ReadSeeker) ([]float32, error) {
	dec := wav.NewDecoder(reader)
	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, err
	}
	if dec.SampleRate != whispercpp.SampleRate {
		return nil, fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	}
	if dec.NumChans != 1 {
		return nil, fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	}

	return buf.AsFloat32Buffer().Data, nil
}
