package controller

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/model/aggregate"
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

		taskRecord := &model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput]{
			Input: aggregate.M3u8dlInput{
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
		return
	}

	taskRecord := req.ToTaskRecord()

	stat, err := os.Stat(req.Filepath)
	if errors.Is(err, os.ErrNotExist) || stat.IsDir() {
		c.JSON(400, gin.H{"err": "路径不存在或为目录"})
		return
	}

	if err := taskRecord.Save(); err != nil {
		c.JSON(400, gin.H{"err": err.Error()})
		return
	}

	if err := controller.subtitleService.AddTask(taskRecord); err != nil {
		c.JSON(400, gin.H{"err": err.Error()})
	}
}

func (controller *TaskController) AddGenerateSubtitleTaskAsync(c *gin.Context) {
	req := new(AddGenerateSubtitleAsyncTaskReq)
	if err := c.BindJSON(req); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{
			"error": "获取文件失败: " + err.Error(),
		})
		return
	}

	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, file.Filename)

	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(500, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	taskRecord := req.ToTaskRecord()

	taskRecord.Input.OnFunishCallback = func(input aggregate.SubtitleInput) {
		if err := os.Remove(tempFilePath); err != nil {

		}

	}

	if err := taskRecord.Save(); err != nil {
		c.JSON(400, gin.H{"err": err.Error()})
		return
	}

	if err := controller.subtitleService.AddTask(taskRecord); err != nil {
		c.JSON(400, gin.H{"err": err.Error()})
	}
}
