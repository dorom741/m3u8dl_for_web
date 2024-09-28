package controller

import (
	"io"
	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/service"
	"strings"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	worker service.M3u8dlWorker
}

func NewTaskController() TaskController {
	return TaskController{
		worker: service.M3u8dlWorkerInstance,
	}
}

func (controller *TaskController) AddTask(c *gin.Context) {
	req := new(BatchAddTaskReq)
	if err := c.BindJSON(req); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
	}

	for _, addTaskReq := range *req {
		if len(addTaskReq.SaveDir) == 0 || len(addTaskReq.Name) == 0 {
			addTaskReq.SaveDir = conf.ConfigInstance.Server.SavePath
		}

		url := strings.Trim(addTaskReq.URL, " ")
		fileName := strings.Trim(addTaskReq.Name, " ")
		fileName = strings.ReplaceAll(fileName, ".mp4", "")

		taskRecord := model.TaskRecord{
			URL:     url,
			Name:    fileName,
			SaveDir: addTaskReq.SaveDir,
		}

		err := taskRecord.Save()
		if err != nil {
			c.JSON(400, gin.H{"err": err.Error()})
			return
		}
		service.M3u8dlWorkerInstance.AddTask(taskRecord)
	}

	c.JSON(200, gin.H{"msg": "ok"})
}

func (controller *TaskController) AddTaskByAria2(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	infra.Logger.Infof("req %s %+v %+v", data, c.Request.Header, err)

	c.JSON(200, gin.H{"msg": "ok"})

}
