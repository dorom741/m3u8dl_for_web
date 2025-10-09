package subtitle

import (
	"context"
	"github.com/sirupsen/logrus"
	"strings"
	"sync"
)

func (service *SubtitleService) BatchTranslate(ctx context.Context, textList []string, sourceLang string, targetLang string) ([]string, error) {
	batchTranslateForSingle := func(textListForTranslate []string) ([]string, error) {
		var (
			wg                     sync.WaitGroup
			sem                    = make(chan struct{}, 5)
			translatedTextListTemp = make([]string, len(textListForTranslate))
		)
		for i, text := range textListForTranslate {
			wg.Add(1)

			go func(index int, t string) {
				sem <- struct{}{}
				defer func() { <-sem }()
				defer wg.Done()

				var (
					lastErr        error
					maxAttempts    = 3
					attempts       = 0
					translatedText = ""
				)
				for attempts < maxAttempts {
					attempts++
					translatedText, lastErr = service.translation.Translate(ctx, t, sourceLang, targetLang)
					if lastErr == nil  {
						translatedTextListTemp[index] = translatedText
						return 
					}
				}
				logrus.Warnf("translate failed after %d attempts,last error: %v", attempts, lastErr)

			}(i, text)
		}

		wg.Wait()
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
