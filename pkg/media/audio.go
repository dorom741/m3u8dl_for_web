package media

import (
	"fmt"
	"io"

	"github.com/go-audio/aiff"
	"github.com/mattetti/audio/decoder"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

func ConvertToWav(input io.ReadSeeker, output io.WriteSeeker) error {
	var (
		dec decoder.Decoder
		buf *audio.IntBuffer
		err error
	)



	wd := wav.NewDecoder(input)

	if wd.IsValidFile() {
		dec = wd
	} else {
		input.Seek(0, 0)
		aiffd := aiff.NewDecoder(input)
		if !aiffd.IsValidFile() {
			return fmt.Errorf("input file isn't a valid wav or aiff file")
		}
		dec = aiffd
	}

	if !dec.WasPCMAccessed() {
		err := dec.FwdToPCM()
		if err != nil {
			panic(err)
		}
	}

	format := dec.Format()

	enc := wav.NewEncoder(output, format.SampleRate, int(dec.SampleBitDepth()), format.NumChannels, 1)
	defer enc.Close()

	buf = &audio.IntBuffer{Format: format, Data: make([]int, 100000)}
	var n int
	var doneReading bool

	for err == nil {
		n, err = dec.PCMBuffer(buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("failed to read the input file:%s", err)
		}
		if n != len(buf.Data) {
			buf.Data = buf.Data[:n]
			doneReading = true
		}
		if err = enc.Write(buf); err != nil {
			return fmt.Errorf("failed to write to the output file%:s", err)
		}
		if doneReading {
			break
		}
	}

	return nil
}
