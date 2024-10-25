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
	inputPath := "../resource/samples/jfk.wav"
	outputPath := strings.ReplaceAll(inputPath, ".wav", ".srt")

	input := model.SubtitleInput{
		Provider:    "ggml-base",
		InputPath:   inputPath,
		SavePath:    outputPath,
		Prompt:      "",
		Temperature: 0,
		Language:    "",
	}

	err := SubtitleServiceInstance.GenerateSubtitle(ctx, input)
	if err != nil {
		t.Error(err)
	}
}
