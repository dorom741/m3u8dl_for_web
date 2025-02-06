package service

import (
	"context"
	"io/fs"
	"path"
	"path/filepath"
	"regexp"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/model"
	"m3u8dl_for_web/model/aggregate"
	"m3u8dl_for_web/pkg/queue_worker"
)

type SubtitleWorkerService struct {
	worker queue_worker.QueueWorker[model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]]

	subtitleConfig *conf.SubtitleConfig
}

func NewSubtitleWorkerService(subtitleConfig *conf.SubtitleConfig) *SubtitleWorkerService {
	option := queue_worker.QueueWorkerOption{
		MaxWorker:   1,
		RetryOnFail: 1,
	}
	service := &SubtitleWorkerService{subtitleConfig: subtitleConfig}
	service.worker = queue_worker.NewQueueWorker[model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]](option, service)

	go service.worker.Run()
	if subtitleConfig != nil {
		go func() {
			err := service.ScanDirToAddTask(subtitleConfig)
			if err != nil {
				logrus.Errorf("scanDir error:%+v", err)
			}
		}()
	}

	return service
}

func (service *SubtitleWorkerService) AddTask(taskRecord *model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]) error {
	return service.worker.AddTask(taskRecord)
}

func (service *SubtitleWorkerService) OnTaskRun(task *model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]) error {
	if err := service.subtitleConfig.WriteLockFile(path.Base(task.Input.InputPath)); err != nil {
		logrus.Errorf("write lock file error %s", err)
	}
	output, err := SubtitleServiceInstance.GenerateSubtitle(context.Background(), task.Input)
	if err != nil {
		return err
	}
	task.Output = *output

	return task.Save()
}

func (service *SubtitleWorkerService) OnTaskFinish(task *model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput], taskErr error) {
	errMsg := ""
	if taskErr != nil {
		logrus.Errorf("%s generate subtitle error:%s", task.Input.InputPath, taskErr.Error())
		errMsg = taskErr.Error()
	} else {
		logrus.Infof("%s generate subtitle success,save to %s", task.Input.InputPath, task.Input.GetSavePath())
	}

	if err := service.subtitleConfig.RemoveLastLockFile(); err != nil {
		logrus.Errorf("remove last lock file error:%s", err.Error())
	}

	err := task.Finish(errMsg)
	if err != nil {
		logrus.Warnf("save generate subtitle task record error:%s", err.Error())
	}
}

func (service *SubtitleWorkerService) ScanDirToAddTask(config *conf.SubtitleConfig) error {
	compiledRexp, err := regexp.Compile(config.Pattern)
	if err != nil {
		return err
	}

	fullDirPath, err := filepath.Abs(config.DirPath)
	if err != nil {
		return err
	}

	allFileList, err := service.scanDir(fullDirPath, compiledRexp)
	if err != nil {
		return err
	}
	totalFileList := len(allFileList)
	logrus.Infof("scanDir %d files to add task", totalFileList)

	var (
		fixMissTranslateChan = make(chan struct{}, 1)
		wg                   sync.WaitGroup
		fixMissTranslateFunc = service.fixMissTranslateFunc(fixMissTranslateChan, &wg, config)
	)

	blacklistJudgementFunc, err := config.GenerateBlacklistJudgement()
	if err != nil {
		return err
	}

	addTask := func(filePath string) {
		newTaskInput := config.SubtitleInput
		newTaskInput.InputPath = filePath

		if newTaskInput.HasSavePathExists() {
			go fixMissTranslateFunc(newTaskInput.GetSavePath())
			logrus.Debugf("skip subtitle task: %s", filePath)
			return
		}
		if config.JustFixMissTranslate {
			return
		}

		task := &model.TaskRecord[aggregate.SubtitleInput, aggregate.SubtitleOutput]{
			Type:   "generateSubtitle",
			State:  model.StateReady,
			Input:  newTaskInput,
			Output: aggregate.SubtitleOutput{},
		}
		// logrus.Infof("add generate subtitle task for path:%s", filePath)

		service.worker.AddTaskBlocking(task)
		if err := task.Save(); err != nil {
			logrus.Warnf("save task error %+v on '%s'", err, task.Input.InputPath)
		}
	}

	if config.Watch {
		logrus.Infof("start dir watch:%s", fullDirPath)
		err := service.doDirWatch(fullDirPath, func(filename string) {
			if compiledRexp.MatchString(filename) && !blacklistJudgementFunc(filename) {
				addTask(filename)
				logrus.Infof("add generate subtitle task for path:%s on dir watch", filename)

			}
		})
		if err != nil {
			return err
		}
	}

	for i := range allFileList {
		fileItem := allFileList[i]
		if !blacklistJudgementFunc(fileItem) {
			logrus.Infof("add generate subtitle task for path:%s on dir watch [%d/%d]", fileItem, i+1, totalFileList)
			addTask(fileItem)

		}

	}

	wg.Wait()
	return nil
}

func (service *SubtitleWorkerService) fixMissTranslateFunc(fixMissTranslateChan chan struct{}, wg *sync.WaitGroup, config *conf.SubtitleConfig) func(filePath string) {
	ctx := context.Background()

	if !config.FixMissTranslate {
		return func(filePath string) {}
	}

	return func(filePath string) {
		wg.Add(1)
		defer wg.Done()
		fixMissTranslateChan <- struct{}{}
		defer func() { <-fixMissTranslateChan }()

		if err := SubtitleServiceInstance.ReGenerateBilingualSubtitleFromSegmentList(
			ctx,
			filePath,
			config.SubtitleInput.Language,
			config.SubtitleInput.TranslateTo,
			filePath,
			true); err != nil {
			logrus.Warnf("re generate bilingual subtitle error %s", err)
		}
	}
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
					go onAddFile(event.Name)
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
