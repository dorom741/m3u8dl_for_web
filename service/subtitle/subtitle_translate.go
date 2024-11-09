package subtitle

import (
	"context"
	"strings"
)

func (service *SubtitleService) BatchTranslate(ctx context.Context, textList []string, sourceLang string, targetLang string) ([]string, error) {
	translatedTextList := make([]string, 0, len(textList))

	support, separator := service.translation.SupportMultipleTextBySeparator()
	if !support {
		for _, text := range textList {
			translatedText, err := service.translation.Translate(ctx, text, sourceLang, targetLang)
			if err != nil {
				return nil, err
			}

			translatedTextList = append(translatedTextList, translatedText)

		}

		return translatedTextList, nil

	}

	textListLen := len(textList)
	step := 20
	for i := 0; i < textListLen; i += step {
		endIndex := i + step
		if endIndex > textListLen {
			endIndex = textListLen
		}
		fullText := strings.Join(textList[i:endIndex], separator)
		translatedText, err := service.translation.Translate(ctx, fullText, sourceLang, targetLang)
		if err != nil {
			return nil, err
		}
		translatedTextList = append(translatedTextList, strings.Split(translatedText, separator)...)

	}

	return translatedTextList, nil
}
