package local

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/pkg/whisper"

	whispercpp "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	wav "github.com/go-audio/wav"
)

var _ whisper.WhisperHandler = &LocalWhisper{}

type LocalWhisper struct {
	modelPath string

	whisperModel whispercpp.Model
	mu           sync.Mutex
	cancelTimer  *time.Timer
}

func NewLocalWhisper(modelPath string) *LocalWhisper {
	return &LocalWhisper{
		modelPath: modelPath,
	}
}

func (localWhisper *LocalWhisper) MaximumFileSize() int64 {
	return 0
}

func (localWhisper *LocalWhisper) newContext() (whispercpp.Context, error) {
	var err error
	localWhisper.mu.Lock()
	defer localWhisper.mu.Unlock()

	if localWhisper.cancelTimer != nil {
		localWhisper.cancelTimer.Stop()
		localWhisper.cancelTimer = nil
	}

	if localWhisper.whisperModel == nil {
		localWhisper.whisperModel, err = whispercpp.New(localWhisper.modelPath)
		if err != nil {
			return nil, err
		}
	}

	modelContext, err := localWhisper.whisperModel.NewContext()
	if err != nil {
		return nil, err
	}

	return modelContext, nil
}

func (localWhisper *LocalWhisper) closeModel() {
	localWhisper.mu.Lock()
	defer localWhisper.mu.Unlock()
	if localWhisper.cancelTimer != nil {
		return
	}

	localWhisper.cancelTimer = time.AfterFunc(time.Minute*10, func() {
		localWhisper.mu.Lock()
		defer localWhisper.mu.Unlock()
		if localWhisper.whisperModel == nil {
			return
		}

		if err := localWhisper.whisperModel.Close(); err != nil {
			panic(err)
		}
		localWhisper.whisperModel = nil
		logrus.Infoln("close local whisper model")
	})
}

func (localWhisper *LocalWhisper) HandleWhisper(ctx context.Context, input whisper.WhisperInput) (*whisper.WhisperOutput, error) {
	modelContext, err := localWhisper.newContext()
	if err != nil {
		return nil, err
	}
	defer localWhisper.closeModel()

	if len(input.Language) != 0 {
		if err = modelContext.SetLanguage(input.Language); err != nil {
			return nil, err
		}
	}

	file, err := os.Open(input.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, duration, err := localWhisper.readPCMInfo(file)
	if err != nil {
		return nil, err
	}

	var (
		segments = make([]whisper.Segment, 0)
		cb       = func(segment whispercpp.Segment) {
			segments = append(segments, whisper.Segment{
				Num:   segment.Num,
				Start: segment.Start.Seconds(),
				End:   segment.End.Seconds(),
				Text:  segment.Text,
			})
		}
	)

	modelContext.ResetTimings()
	if err := modelContext.Process(data, cb, nil); err != nil {
		return nil, err
	}
	modelContext.PrintTimings()

	return &whisper.WhisperOutput{Segments: segments, Duration: duration.Seconds()}, nil
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
