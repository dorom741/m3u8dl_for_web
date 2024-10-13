package whisper

import (
	"fmt"
	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/go-audio/wav"
	"io"
	"os"
)

type Params struct {
	IsTokenize bool

	data []float32
}

func (params *Params) InputFromFile(inputFilePath string) error {
	file, err := os.Open(inputFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return params.readPCM(file)

}

func (params *Params) InputFromWavStream(reader io.ReadSeeker) error {
	return params.readPCM(reader)

}

func (params *Params) readPCM(reader io.ReadSeeker) error {
	dec := wav.NewDecoder(reader)
	if buf, err := dec.FullPCMBuffer(); err != nil {
		return err
	} else if dec.SampleRate != whisper.SampleRate {
		return fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	} else if dec.NumChans != 1 {
		return fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	} else {
		params.data = buf.AsFloat32Buffer().Data
	}

	return nil
}
