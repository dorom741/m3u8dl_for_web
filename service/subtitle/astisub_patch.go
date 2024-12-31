package subtitle

import (
	"context"
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

func SplitBilingualSubtitle(subtitlePath string) (map[string][]whisper.Segment, error) {
	subtitles, err := astisub.OpenFile(subtitlePath)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]whisper.Segment)

	for _, item := range subtitles.Items {
		for _, line := range item.Lines {
			for _, lineItem := range line.Items {
				lang := detectLanguage(lineItem.Text)
				if result[lang] == nil {
					result[lang] = make([]whisper.Segment, 0)
				}

				segment := whisper.Segment{
					Num:   len(result[lang]),
					Start: item.StartAt.Seconds(),
					End:   item.EndAt.Seconds(),
					Text:  lineItem.Text,
				}

				result[lang] = append(result[lang], segment)
			}
		}
	}

	return result, nil
}

func (service *SubtitleService) ReGenerateBilingualSubtitleFromSegmentList(ctx context.Context, subtitlePath string, sourceLang string, targetLang string, savePath string, skipOnExists bool) error {
	sub := NewSubtitleSub()
	subtitleMap, err := SplitBilingualSubtitle(subtitlePath)
	if err != nil {
		return err
	}
	if skipOnExists && subtitleMap[targetLang] != nil {
		return nil
	}
	for _, segment := range subtitleMap[sourceLang] {
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
