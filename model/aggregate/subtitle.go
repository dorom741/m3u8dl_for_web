package aggregate

import (
	"io"
)

type SubtitleInput struct {
	Provider       string    `yaml:"provider" json:"provider"`
	InputPath      string    `yaml:"inputPath" json:"inputPath"`
	SavePath       string    `yaml:"savePath" json:"savePath"`
	Reader         io.Reader `yaml:"-" json:"-"`
	Prompt         string    `yaml:"prompt" json:"prompt"`
	Temperature    float32   `yaml:"temperature" json:"temperature"`
	Language       string    `yaml:"language" json:"language"`
	TranslateTo    string    `yaml:"translateTo" json:"translateTo"`
	ReplaceOnExist bool      `yaml:"replaceOnExist" json:"replaceOnExist"`
}

type SubtitleOutput struct {
	Skip            bool
	ProcessDuration float64
	MediaDuration   float64
	FinishTimestamp int64
	StartTimestamp  int64
}
