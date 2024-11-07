//go:build localWhisper
// +build localWhisper

package subtitle

import (
	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/pkg/whisper/local"
)

func RegisterWhisperProvider() {
	for name, modelPath := range conf.ConfigInstance.LocalWhisperModels {
		whisper.DefaultWhisperProvider.Register(name, local.NewLocalWhisper(modelPath))
		logrus.Infof("register local whisper provider of %s with modelPath:%s", name, modelPath)
	}
}
