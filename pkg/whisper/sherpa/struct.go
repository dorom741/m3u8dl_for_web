package sherpa

import (
	sherpa "github.com/k2-fsa/sherpa-onnx-go/sherpa_onnx"
)

type SherpaConfig struct {
	SplitDuration int64 `yaml:"splitDuration"`

	EmbeddingModelPath string `yaml:"embeddingModelPath"`
	PyannoteModelPath  string `yaml:"pyannoteModelPath"`

	OfflineModelConfig sherpa.OfflineModelConfig `yaml:"offlineModelConfig"`

	OfflineStreamBatchSize int `yaml:"offlineStreamBatchSize"`
}
