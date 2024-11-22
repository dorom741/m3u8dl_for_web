package aggregate

import (
	"io"
	"os"
	"path"
	"strings"
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

func (subtitleInput *SubtitleInput) GetSavePath() string {
	if subtitleInput.SavePath == "" {
		subtitleInput.SavePath = strings.ReplaceAll(subtitleInput.InputPath, path.Ext(subtitleInput.InputPath), ".ass")
	}

	return subtitleInput.SavePath
}

func (subtitleInput *SubtitleInput) HasSavePathExists() bool {
	stat, _ := os.Stat(subtitleInput.GetSavePath())
	return stat != nil && stat.Size() > 0
}

type SubtitleOutput struct {
	Skip            bool
	ProcessDuration float64
	MediaDuration   float64
	FinishTimestamp int64
	StartTimestamp  int64
}
