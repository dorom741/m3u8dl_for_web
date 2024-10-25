package service

import (
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/service/subtitle"
)

var (
	GroqServiceInstance           *GroqService
	M3u8dlServiceInstance         *M3u8dlService
	SubtitleServiceInstance       *subtitle.SubtitleService
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
	SubtitleServiceInstance = subtitle.NewSubtitleService(config.GetAbsCachePath())
	TranslationServiceInstance = NewTranslationService()
	SubtitleWorkerServiceInstance = NewSubtitleWorkerService()
}
