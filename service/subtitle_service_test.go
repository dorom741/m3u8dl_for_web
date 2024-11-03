package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/asticode/go-astisub"
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

func TestAstisub(t *testing.T) {
	s2, err := astisub.OpenFile("E:\\download'\\example-in.vtt")
	if err != nil {
		t.Fatal(err)
	}

	for i := range s2.Items {
		fmt.Printf("s2.Items[0] %+v\n", s2.Items[i])

	}
	fmt.Printf("s2 %+v\n", s2.Styles["astisub-webvtt-default-style-id"].InlineStyle)

	buffer := new(bytes.Buffer)
	encoder := json.NewEncoder(buffer)
	encoder.SetIndent(" ", " ")
	err = encoder.Encode(s2)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(buffer.String())

	s := astisub.NewSubtitles()
	s.Styles["zh"] = &astisub.Style{InlineStyle: &astisub.StyleAttributes{WebVTTStyles: []string{"size:35%"}}}
	s.Items = append(s.Items, &astisub.Item{
		Index:   1,
		StartAt: time.Second,
		EndAt:   time.Second * 2,
		Lines: []astisub.Line{
			{
				Items: []astisub.LineItem{
					{

						Text: "qqqq",
					},
				}},
		},
	})
	subtitleTempFile, err := os.OpenFile("./test.ssa", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		t.Logf("open file error %s", err)
	}
	defer subtitleTempFile.Close()

	err = s.WriteToSSA(subtitleTempFile)
	if err != nil {
		t.Logf("write file error %s", err)
	}

}
