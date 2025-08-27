//go:build localWhisper
// +build localWhisper

package conf

import "m3u8dl_for_web/pkg/whisper/sherpa"

// SherpaConfigType is an alias to the real sherpa.SherpaConfig when built with -tags localWhisper
type SherpaConfigType = sherpa.SherpaConfig
