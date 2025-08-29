package service

import (
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/pkg/whisper/whisper_cpp_client"

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
	var (
		err                error
		translationService translation.ITranslation
	)

	GroqServiceInstance, err = NewGroqService(config.Groq.ApiKey, infra.DefaultCache, config.Server.HttpClientProxy)
	if err != nil {
		panic(err)
	}
	whisper.DefaultWhisperProvider.Register("groq", GroqServiceInstance)

	if config.WhisperCppClientConfig != nil {
		whisperProvider := whispercppclient.NewWhisperCppClient(config.WhisperCppClientConfig, infra.DefaultHttpClient)
		whisper.DefaultWhisperProvider.Register("whisper_cpp_client", whisperProvider)
	}

	M3u8dlServiceInstance = NewM3u8dlService()

	if config.Translation.OpenAiCompatible != nil {
		translationService = translation.NewOpenAiCompatibleTranslation(config.Translation.OpenAiCompatible)
	} else if config.Translation.DeepLX != nil {
		translationService = translation.NewDeepLXTranslation(config.Translation.DeepLX, infra.DefaultHttpClient)
	} else {
		panic("translation config is empty, please check your config file")
	}

	SubtitleServiceInstance = subtitle.NewSubtitleService(config.GetAbsCachePath(), infra.DefaultCache, translationService)

	SubtitleWorkerServiceInstance = NewSubtitleWorkerService(config.Subtitle)
}
