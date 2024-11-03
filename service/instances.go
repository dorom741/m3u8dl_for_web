package service

import (
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/service/subtitle"
	"m3u8dl_for_web/service/translation"
)

var (
	GroqServiceInstance           *GroqService
	M3u8dlServiceInstance         *M3u8dlService
	SubtitleServiceInstance       *subtitle.SubtitleService
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

	translationService := translation.NewDeepLXTranslation(config.Translation.DeeplX.Url, infra.DefaultHttpClient)

	SubtitleServiceInstance = subtitle.NewSubtitleService(config.GetAbsCachePath(), infra.DefaultCache, translationService)

	SubtitleWorkerServiceInstance = NewSubtitleWorkerService()
}
