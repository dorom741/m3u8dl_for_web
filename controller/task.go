package controller

import (
	"path"
	"strings"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/service"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	m3u8dlService   *service.M3u8dlService
	subtitleService *service.SubtitleWorkerService
}

func NewTaskController(m3u8dlService *service.M3u8dlService, subtitleService *service.SubtitleWorkerService) TaskController {
	return TaskController{
		m3u8dlService:   m3u8dlService,
		subtitleService: subtitleService,
	}
}

func (controller *TaskController) AddM3u8dlTask(c *gin.Context) {
	req := new(BatchAddM3u8dlTaskReq)
	if err := c.BindJSON(req); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
	}

	for _, addTaskReq := range *req {
		if len(addTaskReq.SaveDir) == 0 || len(addTaskReq.Name) == 0 {
			addTaskReq.SaveDir = conf.ConfigInstance.GetAbsSavePath()
		}

		url := strings.Trim(addTaskReq.URL, " ")
		fileName := strings.Trim(addTaskReq.Name, " ")
		fileName = strings.ReplaceAll(fileName, ".mp4", "")

		taskRecord := model.TaskRecord[model.M3u8dlInput, model.M3u8dlOutput]{
			Input: model.M3u8dlInput{
				URL:     url,
				Name:    fileName,
				SaveDir: addTaskReq.SaveDir,
			},
			Type:  "m3u8dl",
			State: model.StateReady,
		}

		err := taskRecord.Save()
		if err != nil {
			c.JSON(400, gin.H{"err": err.Error()})
			return
		}
		err = controller.m3u8dlService.AddTask(taskRecord)
		if err != nil {
			c.JSON(400, gin.H{"err": err.Error()})
			return
		}
	}

	c.JSON(200, gin.H{"msg": "ok"})
}

// func (controller *TaskController) AddTaskByAria2(c *gin.Context) {
// 	data, err := io.ReadAll(c.Request.Body)
// 	logrus.Infof("req %s %+v %+v", data, c.Request.Header, err)

//		c.JSON(200, gin.H{"msg": "ok"})
//	}

func (controller *TaskController) AddGenerateSubtitleTask(c *gin.Context) {
	req := new(AddGenerateSubtitleTaskReq)
	if err := c.BindJSON(req); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
	}

	if req.SaveSubtitleFilePath == "" {
		req.SaveSubtitleFilePath = strings.ReplaceAll(req.Filepath, path.Ext(req.Filepath), ".ass")
	}

	taskRecord := model.TaskRecord[model.SubtitleInput, model.SubtitleOutput]{
		Type:  "generateSubtitle",
		State: model.StateReady,
		Input: model.SubtitleInput{
			Provider:  req.Provider,
			InputPath: req.Filepath,
			SavePath:  req.SaveSubtitleFilePath,

			Prompt:      req.Prompt,
			Temperature: req.Temperature,
			Language:    req.Language,
			TranslateTo: req.TranslateTo,
		},
	}

	if err := taskRecord.Save(); err != nil {
		c.JSON(400, gin.H{"err": err.Error()})
		return
	}

	if err := controller.subtitleService.AddTask(taskRecord); err != nil {
		c.JSON(400, gin.H{"err": err.Error()})
	}
}
