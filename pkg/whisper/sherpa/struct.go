package sherpa

import (
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

type SherpaConfig struct {
	EmbeddingModelPath string `yaml:"embeddingModelPath"`
	PyannoteModelPath  string `yaml:"pyannoteModelPath"`

	OfflineModelConfig sherpa.OfflineModelConfig `yaml:"offlineModelConfig"`
}
