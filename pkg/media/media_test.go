package media

import (
	"os"
	"testing"
)

func TestConvertToWav(t *testing.T) {
	inputPath := "/workplace/project/demo/whisper.cpp/samples/jfk.mp3"
	outputPath := "./output.wav"

	inputFile, err := os.Open(inputPath)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	outputFile, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	// if err := ConvertToWav(inputFile, outputFile); err != nil {
	// 	t.Errorf("ConvertToWav error: %v", err)
	// }
}
