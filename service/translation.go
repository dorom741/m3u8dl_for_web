package service

import (
	"context"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/pkg/translation"
)

type TranslationFunc func(ctx context.Context, text string, sourceLang string, targetLang string) (string, error)

type TranslationService struct {
	translationFunc TranslationFunc
}

func NewTranslationService() *TranslationService {
	deeplConf := conf.ConfigInstance.Translation.DeeplX
	return &TranslationService{
		translationFunc: translation.NewDeepLTranslation(deeplConf.Url).Translation,
	}
}

func (service *TranslationService) Translation(ctx context.Context, text string, sourceLang string, targetLang string) (string, error) {
	return service.translationFunc(ctx, text, sourceLang, targetLang)
}
