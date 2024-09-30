package infra

import "sync/atomic"

type QueueWorkerOption struct {
	MaxWorker   int64
	MaxQueueLen int64
	RetryOnFail int
}

type QueueWorker[T any] struct {
	option      QueueWorkerOption
	queue       chan T
	workerCount atomic.Int64
}

func NewQueueWorker[T any](option QueueWorkerOption) QueueWorker[T] {

	return QueueWorker[T]{
		queue:  make(chan T),
		option: option,
	}
}

func (worker *QueueWorker[T]) CurrentWorkingWorker() int64 {
	return worker.workerCount.Load()
}

func (worker *QueueWorker[T]) QueueLen() int {
	return len(worker.queue)
}

func (worker *QueueWorker[T]) Run() {

	for {
		if worker.workerCount.Load() < worker.option.MaxWorker {
			task := <-worker.queue
			go worker.doTask(task)
		}

	}
}

func (worker *QueueWorker[T]) AddTask(task T) error {
	// if worker.queue.Len() >= worker.maxWorker {
	// 	return errors.New("download queue exceeding limits")
	// }

	worker.queue <- task
	return nil
}

func (worker *QueueWorker[T]) doTask(task T) {
	worker.workerCount.Add(1)

	for i := 0 ;i < worker.option.RetryOnFail; i++ {
		err := worker.DoTask(task)
		if err == nil {
			break
		}

		Logger.Errorf("doing task error: %v,retrying of %d", err, i)

	}

	worker.workerCount.Add(-1)

}

func (worker *QueueWorker[T]) DoTask(task T) error {
	return nil

}
