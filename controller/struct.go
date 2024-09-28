package controller

type BatchAddTaskReq []AddTaskReq


type AddTaskReq struct {
	Name           string            `json:"name"`
	URL            string            `json:"url"`
	SaveDir        string            `json:"saveDir"`
	RequestHeaders map[string]string `json:"requestHeaders"`
}


