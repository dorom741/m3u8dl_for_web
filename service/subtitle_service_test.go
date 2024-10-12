package service

import (
	"testing"
)

func TestGenerateSubtitle(t *testing.T) {

	// ctx := context.Background()
	inputPath := "/sdcard/Download/meeting_01.mp4"
	outputName := "meeting_01.mp4"
	subtitleService := NewSubtitleService("../resource/download")

	fileList, err := subtitleService.extractAudioFromVideo(inputPath, subtitleService.tempAudioPath, outputName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fileList)

}
