//go:build localWhisper
// +build localWhisper

package main

import (
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/pkg/whisper"
	"m3u8dl_for_web/pkg/whisper/local"
)

func RegisterWhisperProvider() {
	for name, modelPath := range conf.ConfigInstance.LocalWhisperModels {
		whisper.DefaultWhisperProvider.Register(name, local.NewLocalWhisper(modelPath))
		infra.Logger.Infof("register local whisper provider of %s with modelPath:%s", name, modelPath)
	}
}
