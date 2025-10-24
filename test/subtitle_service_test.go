package test

import (
	"context"
	"os"
	"testing"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/model/aggregate"
	"m3u8dl_for_web/service"
	"m3u8dl_for_web/service/subtitle"
)

func TestGenerateSubtitle(t *testing.T) {
	ctx := context.Background()
	inputPath := "../resource/samples/jfk.wav"

	input := aggregate.SubtitleInput{
		Provider:       "whisper_cpp_client",
		InputPath:      inputPath,
		ReplaceOnExist: true,
		SavePath:       "",
		Prompt:         "",
		Temperature:    0.2,
		Language:       "",
	}

	_, err := service.SubtitleServiceInstance.GenerateSubtitle(ctx, input)
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

func TestScanDirToAddTask(t *testing.T) {
	config := &conf.SubtitleConfig{
		DirPath: "./",
		Pattern: `\.go$`,
		Watch:   true,
		SubtitleInput: aggregate.SubtitleInput{
			Prompt:         "",
			Temperature:    0.0,
			Language:       "",
			TranslateTo:    "",
			ReplaceOnExist: false,
		},
		FixMissTranslate: false,
	}
	err := service.SubtitleWorkerServiceInstance.ScanDirToAddTask(config)
	if err != nil {
		t.Fatal(err)
	}
}

func TestReGenerateSubtitle(t *testing.T) {
	ctx := context.Background()
	inputPath := "../resource/samples/jfk.ass"
	outputPath := "../resource/samples/jfk.zh.ass"

	err := service.SubtitleServiceInstance.ReGenerateBilingualSubtitleFromSegmentList(ctx, inputPath, "en", "zh", outputPath, true)
	if err != nil {
		t.Error(err)
	}
}
