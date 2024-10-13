package whisper

import (
	"fmt"
	"io"
	"os"
	"time"

	// Package imports
	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	// wav "github.com/go-audio/wav"
)

func Process(context whisper.Context, params Params) ([]whisper.Segment, error) {
	fmt.Printf("\n%s\n", context.SystemInfo())

	if len(params.data) == 0 {
		return nil, fmt.Errorf("input data is empty")
	}

	// Segment callback when -tokens is specified
	var cb whisper.SegmentCallback
	if params.IsTokenize {
		cb = func(segment whisper.Segment) {
			fmt.Printf("%02d [%6s->%6s] %s", segment.Num, segment.Start.Truncate(time.Millisecond), segment.End.Truncate(time.Millisecond), segment.Text)
			// for _, token := range segment.Tokens {
			// 	if  context.IsText(token) {
			// 		fmt.Printf(flags.Output(), token.Text, int(token.P*24.0)), " ")
			// 	} else {

			// 	}
			// }
		}
	}

	context.ResetTimings()
	if err := context.Process(params.data, cb, nil); err != nil {
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
