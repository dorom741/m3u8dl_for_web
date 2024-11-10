package subtitle

import (
	"context"
	"strings"
)

func (service *SubtitleService) BatchTranslate(ctx context.Context, textList []string, sourceLang string, targetLang string) ([]string, error) {
	batchTranslateForSingle := func(textListForTranslate []string) ([]string, error) {
		translatedTextListTemp := make([]string, 0)
		for _, text := range textList {
			translatedText, err := service.translation.Translate(ctx, text, sourceLang, targetLang)
			if err != nil {
				return nil, err
			}

			translatedTextListTemp = append(translatedTextListTemp, translatedText)

		}

		return translatedTextListTemp, nil

	}

	support, separator := service.translation.SupportMultipleTextBySeparator()
	if !support {
		return batchTranslateForSingle(textList)
	}

	translatedTextList := make([]string, 0)
	textListLen := len(textList)
	step := 10
	for i := 0; i < textListLen; i += step {
		endIndex := i + step
		if endIndex > textListLen {
			endIndex = textListLen
		}

		textSegment := textList[i:endIndex]
		fullText := strings.Join(textSegment, separator)
		translatedText, err := service.translation.Translate(ctx, fullText, sourceLang, targetLang)
		if err != nil {
			return nil, err
		}
		translatedSegmentTextList := strings.Split(translatedText, separator)

		if len(textSegment) != len(translatedSegmentTextList) {
			translatedSegmentTextList, err = batchTranslateForSingle(textSegment)
			if err != nil {
				return nil, err
			}
		}

		translatedTextList = append(translatedTextList, translatedSegmentTextList...)

	}

	return translatedTextList, nil
}
