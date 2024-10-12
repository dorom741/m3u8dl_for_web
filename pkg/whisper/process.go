package whisper

import (
	"bytes"
	"fmt"
	"io"
	"m3u8dl_for_web/pkg/media"
	"os"
	"time"

	// Package imports
	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	// wav "github.com/go-audio/wav"
)

func Process(model whisper.Model, context whisper.Context, params Params) ([]whisper.Segment, error) {
	var data []float32

	fmt.Printf("\n%s\n", context.SystemInfo())
	fh, err := os.Open(params.InputPath)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	tempBuffer := new(bytes.Buffer)

	media.SplitAudio(fh, tempBuffer)

	data = media.ToFloat32(tempBuffer.Bytes())

	// data, err = media.ConvertMp3ToFloatArray(tempBuffer)
	// if err != nil {
	// 	return nil, err
	// }
	fmt.Printf("ConvertMp3ToFloatArray %+v\n", len(data))

	// Decode the WAV file - load the full buffer
	// dec := wav.NewDecoder(fh)
	// if buf, err := dec.FullPCMBuffer(); err != nil {
	// 	return nil, err
	// } else if dec.SampleRate != whisper.SampleRate {
	// 	return nil, fmt.Errorf("unsupported sample rate: %d", dec.SampleRate)
	// } else if dec.NumChans != 1 {
	// 	return nil, fmt.Errorf("unsupported number of channels: %d", dec.NumChans)
	// } else {
	// 	data = buf.AsFloat32Buffer().Data
	// }

	// Segment callback when -tokens is specified
	var cb whisper.SegmentCallback
	if params.IsTokenize {
		cb = func(segment whisper.Segment) {
			fmt.Printf("%02d [%6s->%6s] ", segment.Num, segment.Start.Truncate(time.Millisecond), segment.End.Truncate(time.Millisecond))
			// for _, token := range segment.Tokens {
			// 	if  context.IsText(token) {
			// 		fmt.Printf(flags.Output(), token.Text, int(token.P*24.0)), " ")
			// 	} else {

			// 	}
			// }
		}
	}

	// Process the data
	fmt.Printf("...processing %q\n", params.InputPath)
	context.ResetTimings()
	if err := context.Process(data, cb, nil); err != nil {
		return nil, err
	}

	context.PrintTimings()

	return Output(os.Stdout, context)
}

// Output text as SRT file
func OutputSRT(w io.Writer, context whisper.Context) error {
	n := 1
	for {
		segment, err := context.NextSegment()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		fmt.Fprintln(w, n)
		fmt.Fprintln(w, srtTimestamp(segment.Start), " --> ", srtTimestamp(segment.End))
		fmt.Fprintln(w, segment.Text)
		fmt.Fprintln(w, "")
		n++
	}
}

func Output(w io.Writer, context whisper.Context) ([]whisper.Segment, error) {
	var segmentList []whisper.Segment
	for {
		segment, err := context.NextSegment()
		if err == io.EOF {
			return segmentList, nil
		} else if err != nil {
			return nil, err
		}
		fmt.Fprintf(w, "[%6s->%6s]", segment.Start.Truncate(time.Millisecond), segment.End.Truncate(time.Millisecond))
		segmentList = append(segmentList, segment)
		// for _, token := range segment.Tokens {
		// 	if !context.IsText(token) {
		// 		continue
		// 	}
		// 	fmt.Fprint(w, " ", Colorize(token.Text, int(token.P*24.0)))
		// }
		// fmt.Fprint(w, "\n")

	}
}

// Return srtTimestamp
func srtTimestamp(t time.Duration) string {
	return fmt.Sprintf("%02d:%02d:%02d,%03d", t/time.Hour, (t%time.Hour)/time.Minute, (t%time.Minute)/time.Second, (t%time.Second)/time.Millisecond)
}
