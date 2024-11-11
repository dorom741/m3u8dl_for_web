package controller

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
