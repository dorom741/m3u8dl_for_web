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
	Filepath string
	SaveSubtitleFilePath string
}


