package sherpa

import (
	"context"
	"os"
	"path"
	"testing"

	"m3u8dl_for_web/pkg/whisper"
)

func TestSherpaWhisper(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}
	basePath := path.Join(dir, "../../../resource/download/sherpa")
	vadModelPath := path.Join(basePath, "silero_vad.onnx")
	whisperDecoderModelPath := path.Join(basePath, "sherpa-onnx-whisper-tiny.en", "tiny.en-decoder.onnx")
	whisperEncoderModelPath := path.Join(basePath, "sherpa-onnx-whisper-tiny.en", "tiny.en-encoder.onnx")
	whisperModelTokensPath := path.Join(basePath, "sherpa-onnx-whisper-tiny.en", "tiny.en-tokens.txt")
	embeddingModelPath := path.Join(basePath, "3dspeaker_speech_eres2net_base_sv_zh-cn_3dspeaker_16k.onnx")
	pyannoteModelPath := path.Join(basePath, "sherpa-onnx-pyannote-segmentation-3-0/model.onnx")

	sherpaWhisper := NewSherpaWhisper(vadModelPath, whisperDecoderModelPath, whisperEncoderModelPath, whisperModelTokensPath, embeddingModelPath, pyannoteModelPath)
	result, err := sherpaWhisper.HandleWhisper(context.Background(), whisper.WhisperInput{
		FilePath: "/workplace/project/demo/m3u8dl_for_web/resource/samples/jfk.wav",
		// FilePath: path.Join(basePath, "sherpa-onnx-whisper-tiny.en/test_wavs/1.wav"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("result %+v", result)
}
