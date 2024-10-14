package media

import (
	"testing"
)

func TestConvertToWavWithFFmpeg(t *testing.T) {
	inputPath := "/workplace/project/demo/whisper.cpp/samples/jfk.mp3"
	outputPath := "../resource/download/output_%03.wav"

	// inputFile, err := os.Open(inputPath)
	// if err != nil {
	// 	panic(err)
	// }
	// defer inputFile.Close()

	// outputFile, err := os.Create(outputPath)
	// if err != nil {
	// 	panic(err)
	// }
	// defer outputFile.Close()

	if err := ConvertToWavWithFFmpeg(inputPath, outputPath,ConvertToWavOption{}); err != nil {
		t.Errorf("Mp3ToWavWithFFmpeg error: %v", err)
	}
}
