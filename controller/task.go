package controller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/model/aggregate"

	// "m3u8dl_for_web/pkg/whisper" // unused in async flow
	"m3u8dl_for_web/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	req := new(AddGenerateSubtitleTaskReq)

	// accept form fields
	req.Provider = c.PostForm("provider")
	if temperature, err := strconv.ParseFloat(c.PostForm("temperature"), 32); err == nil {
		req.Temperature = float32(temperature)
	}
	req.Prompt = c.PostForm("prompt")
	req.Language = c.PostForm("language")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "获取文件失败: " + err.Error()})
		return
	}

	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, file.Filename)

	if err := c.SaveUploadedFile(file, tempFilePath); err != nil {
		c.JSON(500, gin.H{"error": "保存文件失败: " + err.Error()})
		return
	}

	taskRecord := req.ToTaskRecord()
	taskRecord.Input.InputPath = tempFilePath
	taskRecord.Input.JustTranscribe = true

	taskRecord.Input.OnFinishCallback = func(input aggregate.SubtitleInput, output aggregate.SubtitleOutput) {
		if err := os.Remove(tempFilePath); err != nil {
			logrus.Warnf("remove temp file '%s' error:%s", tempFilePath, err)
		}
	}

	logrus.Infof("creating async subtitle task: %+v", taskRecord)

	if err := taskRecord.Save(); err != nil {
		c.JSON(400, gin.H{"err": err.Error()})
		return
	}

	if err := controller.subtitleService.AddTask(taskRecord); err != nil {
		c.JSON(400, gin.H{"err": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"msg":     "task accepted, processing",
		"task_id": strconv.FormatUint(uint64(taskRecord.ID), 10),
	})
}

func (controller *TaskController) GetTaskResult(c *gin.Context) {
	idStr := c.Query("id")
	if len(idStr) == 0 {
		c.JSON(400, gin.H{"err": "missing id"})
		return
	}

	var id uint64
	var err error
	if id, err = strconv.ParseUint(idStr, 10, 64); err != nil {
		c.JSON(400, gin.H{"err": fmt.Sprintf("parse id '%s' error: %s", idStr, err)})
		return
	}

	var task model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]
	db := infra.DataDB.First(&task, id)
	if db.Error != nil {
		c.JSON(404, gin.H{"err": db.Error.Error()})
		return
	}

	c.JSON(200, gin.H{
		"err": "",
		"task": map[string]any{
			"id":     task.ID,
			"state":  task.State,
			"result": task.Result,
			"input":  task.Input,
			"output": task.Output,
		},
	})
}
