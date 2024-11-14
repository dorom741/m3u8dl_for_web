//go:build localWhisper
// +build localWhisper

package subtitle

import (
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/pkg/whisper/local"
	"m3u8dl_for_web/pkg/whisper/sherpa"

	"github.com/sirupsen/logrus"
)

func RegisterWhisperProvider() {
	for name, modelPath := range conf.ConfigInstance.LocalWhisperModels {
		whisper.DefaultWhisperProvider.Register(name, local.NewLocalWhisper(modelPath))
		logrus.Infof("register local whisper provider of %s with modelPath:%s", name, modelPath)
	}
	logrus.Infof("SherpaConfig %+v", conf.ConfigInstance.SherpaConfig)
	if conf.ConfigInstance.SherpaConfig != nil {
		whisper.DefaultWhisperProvider.Register("sherpa", sherpa.NewSherpaWhisper(*conf.ConfigInstance.SherpaConfig))
	}
}
