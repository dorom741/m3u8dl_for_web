package service

import (
	"context"
	"m3u8dl_for_web/conf"
	"path"
	"testing"
)

func init() {

	conf.InitConf("../config.yaml")
	InitService(*conf.ConfigInstance)
}

func TestExtractAudioFromVideo(t *testing.T) {

	// ctx := context.Background()
	inputPath := "/sdcard/Download/meeting_01.mp4"
	outputName := "meeting_01.mp4"
	subtitleService := NewSubtitleService("../resource/download")

	fileList, err := subtitleService.getAudioFromMediaWithFFmpeg(inputPath, subtitleService.tempAudioPath, outputName)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(fileList)

}

func TestGenerateSubtitle(t *testing.T) {
	ctx := context.Background()
	cachePath := "../resource/download"
	inputPath := "/sdcard/Download/meeting_01.mp4"
	outputPath := path.Join(cachePath, "meeting_01.srt")

	subtitleService := NewSubtitleService(cachePath)

	err := subtitleService.GenerateSubtitle(ctx, inputPath, outputPath,"zh")
	if err != nil {
		t.Fatal(err)
	}

}
