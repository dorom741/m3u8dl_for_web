package service

import (
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/pkg/whisper"
)

var (
	GroqServiceInstance           *GroqService
	M3u8dlServiceInstance         *M3u8dlService
	SubtitleServiceInstance       *SubtitleService
	TranslationServiceInstance    *TranslationService
	SubtitleWorkerServiceInstance *SubtitleWorkerService
)

func InitService(config *conf.Config) {
	var err error

	GroqServiceInstance, err = NewGroqService(config.Groq.ApiKey, infra.DefaultCache, config.Server.HttpClientProxy)
	if err != nil {
		panic(err)
	}

	whisper.DefaultWhisperProvider.Register("groq", GroqServiceInstance)

	M3u8dlServiceInstance = NewM3u8dlService()
	SubtitleServiceInstance = NewSubtitleService(config.GetTempPath())
	TranslationServiceInstance = NewTranslationService()
	SubtitleWorkerServiceInstance = NewSubtitleWorkerService()
}
