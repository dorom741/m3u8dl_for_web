package service

import (
	"io"
	"os"
	"sync/atomic"

	"github.com/orestonce/m3u8d"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"
)

var M3u8dlWorkerInstance = NewM3u8dlWorker()

type M3u8dlWorker struct {
	queue       *infra.MessageQueue[model.TaskRecord]
	maxWorker   int64
	workerCount atomic.Int64
}

func NewM3u8dlWorker() M3u8dlWorker {

	return M3u8dlWorker{
		queue:     infra.NewMessageQueue[model.TaskRecord](conf.ConfigInstance.Server.MaxWorker),
		maxWorker: conf.ConfigInstance.Server.MaxWorker,
	}
}

func (worker *M3u8dlWorker) WorkerRun() {
	taskWrap := func(taskRecord model.TaskRecord) {
		worker.workerCount.Add(1)
		worker.doDownload(taskRecord)
		worker.workerCount.Add(-1)
	}

	for {
		if worker.workerCount.Load() < int64(worker.maxWorker) {
			task := worker.queue.Pop()
			go taskWrap(task)
		}

	}
}

func (worker *M3u8dlWorker) AddTask(taskRecord model.TaskRecord) error {
	// if worker.queue.Len() >= worker.maxWorker {
	// 	return errors.New("download queue exceeding limits")
	// }

	worker.queue.Push(taskRecord)
	return nil
}

func (worker *M3u8dlWorker) doDownload(taskRecord model.TaskRecord) {
	req := taskRecord.ToStartDownloadReq()
	req.ProgressBarShow = true
	downloadEnv := m3u8d.DownloadEnv{}
	infra.Logger.Infof("m3u8dl req %+v", req)
	errMsg := downloadEnv.StartDownload(req)
	if errMsg != "" {
		infra.Logger.Errorf("m3u8 '%s' download error %s", req.M3u8Url, errMsg)
		taskRecord.Finish(errMsg)

		return
	}

	resp := downloadEnv.WaitDownloadFinish()
	if io.ErrUnexpectedEOF.Error() == resp.ErrMsg {
		infra.Logger.Warnf("m3u8 '%s' merge error:%s,retry with ffmpeg", req.M3u8Url, resp.ErrMsg)

	} else if resp.ErrMsg != "" {
		infra.Logger.Errorf("m3u8 '%s' download error %s", req.M3u8Url, resp.ErrMsg)
		taskRecord.Finish(resp.ErrMsg)
		return
	}

	if resp.IsSkipped {
		infra.Logger.Warnf("m3u8 url:%s download is skiped", req.M3u8Url)
		taskRecord.Finish("skiped")
		return
	}

	infra.Logger.Infof("m3u8 url:%s download is success,save to %s", req.M3u8Url, resp.SaveFileTo)
	err := os.Chmod(resp.SaveFileTo, 0777) // 注意前面的 0，表示八进制
	if err != nil {
		infra.Logger.Warnf("set permissive permissions on '%s' error ", resp.SaveFileTo)
		return
	}

	taskRecord.Finish("")

}
