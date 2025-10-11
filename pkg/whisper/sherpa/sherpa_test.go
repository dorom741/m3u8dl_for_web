package sherpa

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"testing"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"

	"m3u8dl_for_web/pkg/whisper"
)

func TestSherpaWhisper(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	basePath := path.Join(dir, "../../../download/model")
	basePath, _ = filepath.Abs(basePath)
	// vadModelPath := path.Join(basePath, "silero_vad.onnx")
	senseVoicePath := "sherpa-onnx-sense-voice-zh-en-ja-ko-yue-2024-07-17"

	modelTokensPath := path.Join(basePath, senseVoicePath, "tokens.txt")
	modelPath := path.Join(basePath, senseVoicePath, "model.onnx")

	// modelTokensPath := path.Join(basePath, "sherpa-onnx-whisper-tiny.en", "tiny.en-tokens.txt")

	sherpaConfig := SherpaConfig{
		EmbeddingModelPath: path.Join(basePath, "3dspeaker_speech_eres2net_sv_zh-cn_16k-common.onnx"),
		PyannoteModelPath:  path.Join(basePath, "sherpa-onnx-pyannote-segmentation-3-0/model.onnx"),
		OfflineModelConfig: sherpa.OfflineModelConfig{
			// Whisper: sherpa.OfflineWhisperModelConfig{
			// 	Decoder:      path.Join(basePath, "sherpa-onnx-paraformer-zh-2023-09-14", "model.int8.onnx"),
			// 	Encoder:      path.Join(basePath, "sherpa-onnx-whisper-tiny.en", "tiny.en-encoder.onnx"),
			// 	Language:     "",
			// 	Task:         "",
			// 	TailPaddings: 0,
			// },
			SenseVoice: sherpa.OfflineSenseVoiceModelConfig{
				Model: modelPath,
			},
			Tokens: modelTokensPath,
		},
	}

	sherpaWhisper := NewSherpaWhisper(sherpaConfig)
	result, err := sherpaWhisper.HandleWhisper(context.Background(), whisper.WhisperInput{
		FilePath: "../../../resource/samples/jfk.wav",
		// FilePath: path.Join(basePath, "sherpa-onnx-whisper-tiny.en/test_wavs/1.wav"),
	})
	if err != nil {
		t.Fatal(err)
	}

	resultBytes, err := json.Marshal(result)
	t.Logf("result %s %+v", resultBytes, err)
}
