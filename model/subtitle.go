package model

import (
	"io"
)

type SubtitleInput struct {
	Provider    string    `json:"provider"`
	InputPath   string    `json:"inputPath"`
	SavePath    string    `json:"savePath"`
	Reader      io.Reader `json:"-"`
	Prompt      string    `json:"prompt"`
	Temperature float32   `json:"temperature"`
	Language    string    `json:"language"`
}

type SubtitleOutput struct{}
