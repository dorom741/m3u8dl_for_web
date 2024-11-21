package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"m3u8dl_for_web/model"
	"m3u8dl_for_web/model/aggregate"
	"m3u8dl_for_web/pkg/queue_worker"

	"github.com/orestonce/m3u8d"
)

var _ queue_worker.QueueWorkerConsumer[model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput]] = &M3u8dlService{}

type M3u8dlService struct {
	worker queue_worker.QueueWorker[model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput]]
}

func NewM3u8dlService() *M3u8dlService {
	option := queue_worker.QueueWorkerOption{
		MaxWorker:   4,
		RetryOnFail: 3,
	}
	service := &M3u8dlService{}
	go service.worker.Run()

	service.worker = queue_worker.NewQueueWorker[model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput]](option, service)

	return service
}

func (service *M3u8dlService) AddTask(taskRecord *model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput]) error {
	return service.worker.AddTask(taskRecord)
}

func (service *M3u8dlService) OnTaskRun(task *model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput]) error {
	req := task.Input.ToStartDownloadReq()
	req.ProgressBarShow = true
	downloadEnv := m3u8d.DownloadEnv{}
	logrus.Infof("m3u8dl req %+v", req)
	errMsg := downloadEnv.StartDownload(req)
	if errMsg != "" {
		return fmt.Errorf("%s", errMsg)
	}

	resp := downloadEnv.WaitDownloadFinish()
	if strings.Contains(resp.ErrMsg, io.ErrUnexpectedEOF.Error()) {
		logrus.Warnf("m3u8 '%s' merge error:%s,retry with ffmpeg", req.M3u8Url, resp.ErrMsg)
		if output, err := service.MergeWithFFMPEG(task); err != nil {
			return fmt.Errorf("merge with ffmpeg error:%s,output: %s", err, output)
		}
		tempDir, mergeTempFile, err := service.getTempDirAndFile(task)
		if err != nil {
			return err
		}

		err = os.RemoveAll(tempDir)
		if err != nil {
			return err
		}

		err = os.RemoveAll(mergeTempFile)
		if err != nil {
			return err
		}
		return nil

	} else if resp.ErrMsg != "" {
		return fmt.Errorf("%s", resp.ErrMsg)
	}

	if resp.IsSkipped {
		logrus.Warnf("m3u8 url:%s download is skiped,save to %s", task.Input.URL, resp.SaveFileTo)
	}

	return nil
}

func (service *M3u8dlService) OnTaskFinish(task *model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput], taskErr error) {
	errMsg := ""
	if taskErr != nil {
		logrus.Errorf("m3u8 url:%s download error:%s", task.Input.URL, taskErr.Error())
		errMsg = taskErr.Error()
	} else {
		logrus.Infof("m3u8 url:%s download success,save to %s", task.Input.URL, task.Input.GetSavePath())
		err := os.Chmod(task.Input.GetSavePath(), 0o777) // 注意前面的 0，表示八进制
		if err != nil {
			logrus.Warnf("set permissive permissions on '%s' error ", task.Input.GetSavePath())
		}
	}

	err := task.Finish(errMsg)
	if err != nil {
		logrus.Warnf("save download task record error:%s", err.Error())
	}
}

func (service *M3u8dlService) getVideoId(url string) string {
	tmp1 := sha256.Sum256([]byte(url))
	return hex.EncodeToString(tmp1[:])
}

func (service *M3u8dlService) getTempDirAndFile(task *model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput]) (tempDir string, mergeTempFile string, err error) {
	videoId := service.getVideoId(task.Input.URL)
	tempDir, err = filepath.Abs(path.Join(task.Input.SaveDir, "downloading", videoId))
	if err != nil {
		return
	}

	return tempDir, path.Join(tempDir, task.Input.Name+".mp4.temp"), nil
}

func (service *M3u8dlService) MergeWithFFMPEG(task *model.TaskRecord[aggregate.M3u8dlInput, aggregate.M3u8dlOutput]) (string, error) {
	var out bytes.Buffer
	workingDir, _, err := service.getTempDirAndFile(task)
	if err != nil {
		return "", err
	}

	fileListPath := path.Join(workingDir, "filelist.txt")

	savePath, err := filepath.Abs(task.Input.GetSavePath())
	if err != nil {
		return "", err
	}

	ffmpegArgs := []string{"-f", "concat", "-i", fileListPath, "-c", "copy", "-y", savePath}

	ffmpegExecPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", err
	}

	cmd := exec.Command(ffmpegExecPath, ffmpegArgs...)
	cmd.Stdout = &out
	cmd.Stderr = &out

	logrus.Infof("exec ffmpeg command line:%s", cmd.String())

	err = cmd.Run()

	return out.String(), err
}
