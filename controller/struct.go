package controller

type Response struct {
	OK      bool        `json:"ok"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type BatchAddTaskReq []AddTaskReq

type AddTaskReq struct {
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	SaveDir        string            `json:"saveDir"`
	RequestHeaders map[string]string `json:"requestHeaders"`
}
