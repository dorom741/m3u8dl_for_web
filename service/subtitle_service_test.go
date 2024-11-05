package service

import (
	"context"
	"os"
	"strings"
	"testing"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/service/subtitle"
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

func TestAstisub(t *testing.T) {
	s := subtitle.NewSubtitleSub()
	s.Metadata().Comments = []string{"test"}
	s.Metadata().Language = "zh"
	s.Metadata().Title = "Title"

	s.AddLine(1, 1.1, 2, "main test", "secondText")
	s.AddLine(2, 3, 4, "main test", "")

	subtitleTempFile, err := os.OpenFile("./test.ass", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		t.Logf("open file error %s", err)
	}
	defer subtitleTempFile.Close()

	err = s.WriteToFile(subtitleTempFile)
	if err != nil {
		t.Logf("write file error %s", err)
	}
}
