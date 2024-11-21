package service

import (
	"context"
	"io/fs"
	"path/filepath"
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/model/aggregate"
	"m3u8dl_for_web/pkg/queue_worker"
)

type SubtitleWorkerService struct {
	worker queue_worker.QueueWorker[model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]]
}

func NewSubtitleWorkerService(subtitleConfig *conf.SubtitleConfig) *SubtitleWorkerService {
	option := queue_worker.QueueWorkerOption{
		MaxWorker:   1,
		RetryOnFail: 1,
	}
	service := &SubtitleWorkerService{}
	service.worker = queue_worker.NewQueueWorker[model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]](option, service)

	go service.worker.Run()
	if subtitleConfig != nil {
		go func() {
			err := service.ScanDirToAddTask(subtitleConfig.DirPath, subtitleConfig.Pattern, subtitleConfig.Watch, subtitleConfig.SubtitleInput)
			if err != nil {
				logrus.Errorf("scanDir error:%+v", err)
			}
		}()
	}

	return service
}

func (service *SubtitleWorkerService) AddTask(taskRecord model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]) error {
	return service.worker.AddTask(taskRecord)
}

func (service *SubtitleWorkerService) OnTaskRun(task model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]) error {
	output, err := SubtitleServiceInstance.GenerateSubtitle(context.Background(), task.Input)
	if err != nil {
		return err
	}
	task.Output = *output

	return task.Save()
}

func (service *SubtitleWorkerService) OnTaskFinish(task model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput], taskErr error) {
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

func (service *SubtitleWorkerService) ScanDirToAddTask(dirPath string, matchPattern string, watch bool, input aggregate.SubtitleInput) error {
	compiledRexp, err := regexp.Compile(matchPattern)
	if err != nil {
		return err
	}

	fullDirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return err
	}

	allFileList, err := service.scanDir(fullDirPath, compiledRexp)
	if err != nil {
		return err
	}

	addTask := func(filePath string) {
		newTaskInput := input
		newTaskInput.InputPath = filePath
		task := model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]{
			Type:   "generateSubtitle",
			State:  model.StateReady,
			Input:  newTaskInput,
			Output: aggregate.SubtitleOutput{},
		}

		service.worker.AddTaskBlocking(task)
		if err := task.Save(); err != nil {
			logrus.Warnf("save task error %+v on '%s'", err, task.Input.InputPath)
		}
	}

	if watch {
		logrus.Infof("start dir watch:%s", fullDirPath)
		service.doDirWatch(fullDirPath, func(filename string) {
			if compiledRexp.MatchString(filename) {
				addTask(filename)
				logrus.Infof("add generate subtitle task for path:%s on dir watch", filename)

			}
		})
	}

	for i := range allFileList {
		fileItem := allFileList[i]
		addTask(fileItem)

		logrus.Infof("add generate subtitle task for path:%s", fileItem)
	}

	return nil
}

func (service *SubtitleWorkerService) doDirWatch(dirPath string, onAddFile func(filename string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	err = watcher.Add(dirPath)
	if err != nil {
		return err
	}

	defer func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logrus.Infof("fsnotify event: %+v", event)
				if event.Has(fsnotify.Create) {
					onAddFile(event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Errorf("watcher dir error:%+v", err)
			}
		}
	}()

	return nil
}

func (service *SubtitleWorkerService) scanDir(dirPath string, compiledRexp *regexp.Regexp) ([]string, error) {
	fileList := make([]string, 0)
	err := filepath.Walk(dirPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !compiledRexp.MatchString(path) {
			return nil
		}

		fileList = append(fileList, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return fileList, nil
}
