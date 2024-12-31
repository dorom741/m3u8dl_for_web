package subtitle

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"strings"

	"m3u8dl_for_web/pkg/whisper"

	"github.com/asticode/go-astisub"
	"github.com/pemistahl/lingua-go"
)

func detectLanguage(text string) string {
	detector := lingua.NewLanguageDetectorBuilder().
		FromIsoCodes639_3(lingua.ENG, lingua.ZHO, lingua.JPN, lingua.KOR).
		Build()

	result := detector.DetectMultipleLanguagesOf(text)

	if len(result) == 0 {
		return "Unknown"
	}

	return strings.ToLower(result[0].Language().IsoCode639_1().String())
}

func SplitBilingualSubtitle(subtitlePath string) (bool, []whisper.Segment, error) {
	subtitles, err := astisub.OpenFile(subtitlePath)
	if err != nil {
		return false, nil, err
	}

	result := make([]whisper.Segment, 0)
	lastTimestamp := ""
	currentTimestamp := ""
	isBilingualSubtitle := false

	for _, item := range subtitles.Items {
		currentTimestamp = fmt.Sprintf("%.2f-%.2f", item.StartAt.Seconds(), item.EndAt.Seconds())
		if currentTimestamp == lastTimestamp && item.StartAt.Seconds() != 0 {
			isBilingualSubtitle = true
			continue
		}
		lastTimestamp = currentTimestamp
		for _, line := range item.Lines {
			for _, lineItem := range line.Items {
				segment := whisper.Segment{
					Num:   len(result),
					Start: item.StartAt.Seconds(),
					End:   item.EndAt.Seconds(),
					Text:  lineItem.Text,
				}
				result = append(result, segment)
			}
		}
	}

	return isBilingualSubtitle, result, nil
}

func (service *SubtitleService) ReGenerateBilingualSubtitleFromSegmentList(ctx context.Context, subtitlePath string, sourceLang string, targetLang string, savePath string, skipOnExists bool) error {
	sub := NewSubtitleSub()
	isBilingualSubtitle, subtitleList, err := SplitBilingualSubtitle(subtitlePath)
	if err != nil {
		return err
	}
	logrus.Infof("subtitle file '%s' isBilingualSubtitle: %v ,segment len: %d", subtitlePath, isBilingualSubtitle, len(subtitleList))

	if skipOnExists && isBilingualSubtitle {
		return nil
	}

	for _, segment := range subtitleList {
		translatedText, err := service.translation.Translate(ctx, segment.Text, "", targetLang)
		if err != nil {
			continue
		}

		sub.AddLine(segment.Num, segment.Start, segment.End, segment.Text, translatedText)
	}

	subFile, err := os.Create(savePath)
	defer subFile.Close()
	return sub.WriteToFile(subFile)
}
