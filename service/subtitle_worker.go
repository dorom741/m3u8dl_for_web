package service

import (
	"context"
	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/model"
	"m3u8dl_for_web/pkg/queue_worker"
)

type SubtitleWorkerService struct {
	worker queue_worker.QueueWorker[model.TaskRecord[model.SubtitleInput, model.SubtitleOutput]]
}

func NewSubtitleWorkerService() *SubtitleWorkerService {
	option := queue_worker.QueueWorkerOption{
		MaxWorker:   1,
		RetryOnFail: 1,
	}
	service := &SubtitleWorkerService{}
	go service.worker.Run()

	service.worker = queue_worker.NewQueueWorker(option, service)

	return service
}

func (service *SubtitleWorkerService) AddTask(taskRecord model.TaskRecord[model.SubtitleInput, model.SubtitleOutput]) error {
	return service.worker.AddTask(taskRecord)
}

func (service *SubtitleWorkerService) OnTaskRun(task model.TaskRecord[model.SubtitleInput, model.SubtitleOutput]) error {
	if err := SubtitleServiceInstance.GenerateSubtitle(context.Background(), task.Input); err != nil {
		return err
	}
	return nil
}

func (service *SubtitleWorkerService) OnTaskFinish(task model.TaskRecord[model.SubtitleInput, model.SubtitleOutput], taskErr error) {
	errMsg := ""
	if taskErr != nil {
		logrus.Errorf("%s generate subtitle error:%s", task.Input.InputPath, taskErr.Error())
		errMsg = taskErr.Error()
	} else {
		logrus.Infof("%s generate subtitle success,save to %s", task.Input.InputPath, task.Input.SavePath)
	}

	err := task.Finish(errMsg)
	if err != nil {
		logrus.Warnf("save generate subtitle task record error:%s", err.Error())
	}
}
