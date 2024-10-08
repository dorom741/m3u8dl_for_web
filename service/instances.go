package service

import (
	"m3u8dl_for_web/conf"
)

var (
	GroqServiceInstance           *GroqService
	M3u8dlServiceInstance         *M3u8dlService
	SubtitleServiceInstance       *SubtitleService
	TranslationServiceInstance    *TranslationService
	SubtitleWorkerServiceInstance *SubtitleWorkerService
)

func InitService(config conf.Config) {
	var err error

	GroqServiceInstance, err = NewGroqService(config.Groq.ApiKey, config.Groq.CachePath, config.HttpClient.Proxy)
	if err != nil {
		panic(err)
	}

	M3u8dlServiceInstance = NewM3u8dlService()
	SubtitleServiceInstance = NewSubtitleService(config.GetTempPath())
	TranslationServiceInstance = NewTranslationService()
	SubtitleWorkerServiceInstance = NewSubtitleWorkerService()
}
