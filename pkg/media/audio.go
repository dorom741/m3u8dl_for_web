package media

import (
	"encoding/binary"
	"fmt"

	gomp3 "github.com/hajimehoshi/go-mp3"

	"io"
)

func ConvertMp3ToFloatArray(input io.Reader) ([]float32, error) {
	mp3Decoder, err := gomp3.NewDecoder(input)
	if err != nil {
		return nil, err
	}

	mp3Bytes, err := io.ReadAll(mp3Decoder)
	if err != nil {
		return nil, err
	}
	fmt.Printf("mp3Decoder.SampleRate(): %v\n", mp3Decoder.SampleRate())

	data := ToFloat32(mp3Bytes)

	return data, nil
	// return bpm.ReadFloatArray(data), nil

	// bpm.ScanForBpm(data, 40, 200, 1024, mp3Decoder.SampleRate())

	// formatter := audio.FormatMono225008bLE
	// formatter.SampleRate = mp3Decoder.SampleRate()
	// var out []float32
	// // bytesPerSample := int((formatter.BitDepth-1)/8 + 1)
	// buf := bytes.NewBuffer(mp3Bytes)
	// // out1 := make([]float32, len(mp3Bytes)/bytesPerSample)
	// binary.Read(buf, formatter.Endianness, &out)

	// pcmBuffer := audio.NewPCMByteBuffer(mp3Bytes, formatter)

	// return out, nil

}

func ToFloat32(bytes []byte) []float32 {
	var result []float32
	for i := 0; i < len(bytes)/2; i++ {
		s16 := int16(binary.LittleEndian.Uint16(bytes[2*i : 2*i+2]))
		float := float32(s16) / 32768
		result = append(result, float)
	}

	return result
}
