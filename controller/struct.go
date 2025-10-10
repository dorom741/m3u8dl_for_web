package controller

import (
	"io"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/model/aggregate"
)

type Response struct {
	OK      bool        `json:"ok"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type BatchAddM3u8dlTaskReq []AddM3u8dlTaskReq

type AddM3u8dlTaskReq struct {
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	SaveDir        string            `json:"saveDir"`
	RequestHeaders map[string]string `json:"requestHeaders"`
}

type AddGenerateSubtitleTaskReq struct {
	Provider             string `json:"provider"`
	Filepath             string `json:"filepath"`
	SaveSubtitleFilePath string `json:"saveSubtitleFilePath"`

	Prompt         string  `json:"prompt"`
	Temperature    float32 `json:"temperature"`
	Language       string  `json:"language"`
	TranslateTo    string  `json:"translateTo"`
	ReplaceOnExist bool    `json:"replaceOnExist"`
}

func (req *AddGenerateSubtitleTaskReq) ToTaskRecord() *model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput] {
	return &model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]{
		Type:  "generateSubtitle",
		State: model.StateReady,
		Input: aggregate.SubtitleInput{
			Provider:  req.Provider,
			InputPath: req.Filepath,
			SavePath:  req.SaveSubtitleFilePath,

			Prompt:         req.Prompt,
			Temperature:    req.Temperature,
			Language:       req.Language,
			TranslateTo:    req.TranslateTo,
			ReplaceOnExist: req.ReplaceOnExist,
		},
	}

}

type AddGenerateSubtitleAsyncTaskReq struct {
	*AddGenerateSubtitleTaskReq

	io.Reader
}
