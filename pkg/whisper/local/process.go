package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"m3u8dl_for_web/pkg/whisper"

	whispercpp "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	wav "github.com/go-audio/wav"
	"github.com/sirupsen/logrus"
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

	modelContext, err := model.NewContext()
	if err != nil {
		return nil, err
	}

	if err = modelContext.SetLanguage(input.Language); err != nil {
		return nil, err
	}

	logrus.Debugln(modelContext.SystemInfo())

	reader := input.Reader
	if reader == nil {
		file, err := os.Open(input.FilePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		reader = file
	}

	data, duration, err := localWhisper.readPCMInfo(reader)
	if err != nil {
		return nil, err
	}

	var (
		segments = make([]whispercpp.Segment, 0)
		cb       = func(segment whispercpp.Segment) {
			segments = append(segments, segment)
		}
	)

	modelContext.ResetTimings()
	if err := modelContext.Process(data, cb, nil); err != nil {
		return nil, err
	}

	modelContext.PrintTimings()

	return &LocalWhisperOutput{SegmentList: segments, Duration: duration}, nil
}

func (localWhisper *LocalWhisper) readPCMInfo(reader io.ReadSeeker) ([]float32, time.Duration, error) {
	dec := wav.NewDecoder(reader)

	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, 0, err
	}
	if dec.SampleRate != whispercpp.SampleRate {
		return nil, 0, fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	}
	if dec.NumChans != 1 {
		return nil, 0, fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	}

	duration, err := dec.Duration()
	if err != nil {
		return nil, 0, err
	}

	return buf.AsFloat32Buffer().Data, duration, nil
}
