package aggregate

import (
	"io"
	"m3u8dl_for_web/pkg/whisper"
	"os"
	"path"
	"strings"
)

type SubtitleInput struct {
	Provider         string                              `yaml:"provider" json:"provider"`
	InputPath        string                              `yaml:"inputPath" json:"inputPath"`
	SavePath         string                              `yaml:"savePath" json:"savePath"`
	Reader           io.Reader                           `yaml:"-" json:"-"`
	Prompt           string                              `yaml:"prompt" json:"prompt"`
	Temperature      float32                             `yaml:"temperature" json:"temperature"`
	Language         string                              `yaml:"language" json:"language"`
	TranslateTo      string                              `yaml:"translateTo" json:"translateTo"`
	ReplaceOnExist   bool                                `yaml:"replaceOnExist" json:"replaceOnExist"`
	JustTranscribe   bool                                `yaml:"justTranscribe" json:"justTranscribe"`
	OnFinishCallback func(SubtitleInput, SubtitleOutput) `yaml:"-" json:"-"`
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
	Skip            bool              `json:"skip"`
	ProcessDuration float64           `json:"processDuration"`
	MediaDuration   float64           `json:"mediaDuration"`
	FinishTimestamp int64             `json:"finishTimestamp"`
	StartTimestamp  int64             `json:"startTimestampip"`
	SegmentList     []whisper.Segment `json:"segmentList"`
}
