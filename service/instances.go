package service

import (
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/pkg/whisper/lateral"
	whispercppclient "m3u8dl_for_web/pkg/whisper/whisper_cpp_client"

	"m3u8dl_for_web/service/subtitle"
	"m3u8dl_for_web/service/translation"
)

var (
	GroqServiceInstance           *GroqService
	M3u8dlServiceInstance         *M3u8dlService
	SubtitleServiceInstance       *subtitle.SubtitleService
	SubtitleWorkerServiceInstance *SubtitleWorkerService
	TranslationServiceInstance    *translation.TranslationProviderHub
)

func InitService(config *conf.Config) {
	var (
		err error
	)

	if config.Groq.ApiKey != "" {
		GroqServiceInstance, err = NewGroqService(config.Groq.ApiKey, infra.DefaultCache, config.Server.HttpClientProxy)
		if err != nil {
			panic(err)
		}
		whisper.DefaultWhisperProvider.Register("groq", GroqServiceInstance)
	}
	
	if config.WhisperCppClientConfig != nil {
		whisperProvider := whispercppclient.NewWhisperCppClient(config.WhisperCppClientConfig, infra.DefaultHttpClient)
		whisper.DefaultWhisperProvider.Register("whisper_cpp_client", whisperProvider)
	}

	if config.LateralConfig != nil {
		whisperProvider := lateral.NewLateralProvider(config.LateralConfig, infra.DefaultHttpClient)
		whisper.DefaultWhisperProvider.Register("lateral", whisperProvider)
	}

	M3u8dlServiceInstance = NewM3u8dlService()

	TranslationServiceInstance, err = translation.NewTranslationProviderHub(config.Translation, infra.DefaultHttpClient)
	if err != nil {
		panic(err)
	}

	SubtitleServiceInstance = subtitle.NewSubtitleService(config.GetAbsCachePath(), infra.DefaultCache, TranslationServiceInstance)

	SubtitleWorkerServiceInstance = NewSubtitleWorkerService(config.Subtitle)
}
