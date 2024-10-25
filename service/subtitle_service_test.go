package service

import (
	"context"
	"strings"
	"testing"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/model"
)

func init() {
	conf.InitConf("../config.yaml")
	InitService(conf.ConfigInstance)
}

func TestGenerateSubtitle(t *testing.T) {
	ctx := context.Background()
	cachePath := "../resource/download"
	inputPath := "../m3u8dl_for_web/resource/samples/jfk.wav"
	outputPath := strings.ReplaceAll(inputPath, ".wav", ".srt")
	subtitleService := NewSubtitleService(cachePath)
	input := model.SubtitleInput{
		Provider:    "",
		InputPath:   inputPath,
		SavePath:    outputPath,
		Prompt:      "",
		Temperature: 0,
		Language:    "zh",
	}

	err := subtitleService.GenerateSubtitle(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
}
