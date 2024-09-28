package service

import (
	"github.com/orestonce/m3u8d"

	"m3u8dl_for_web/conf"
	"m3u8dl_for_web/infra"
	"m3u8dl_for_web/model"
)

var M3u8dlWorkerInstance = NewM3u8dlWorker()

type M3u8dlWorker struct {
	queue     *infra.MessageQueue[model.TaskRecord]
	maxWorker int
}

func NewM3u8dlWorker() M3u8dlWorker {

	return M3u8dlWorker{
		queue:     infra.NewMessageQueue[model.TaskRecord](conf.ConfigInstance.Server.MaxWorker),
		maxWorker: conf.ConfigInstance.Server.MaxWorker,
	}
}

func (worker *M3u8dlWorker) WorkerRun() {
	for {
		task := worker.queue.Pop()
		worker.doDownload(task)
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
	if resp.ErrMsg != "" {
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

	taskRecord.Finish("")

}
