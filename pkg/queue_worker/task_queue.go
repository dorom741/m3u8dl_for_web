package queue_worker

import (
	"errors"
	"sync/atomic"
)

type QueueWorkerOption struct {
	MaxWorker   int64
	MaxQueueLen int64
	RetryOnFail int
}

type QueueWorker[T any] struct {
	option      QueueWorkerOption
	queue       chan T
	workerCount atomic.Int64

	onTaskRun    func(task T) error
	onTaskFinish func(task T, err error)
}

func NewQueueWorker[T any](option QueueWorkerOption, consumer QueueWorkerConsumer[T]) QueueWorker[T] {

	if option.MaxQueueLen == 0 {
		option.MaxQueueLen = 10
	}

	return QueueWorker[T]{
		queue:        make(chan T, option.MaxQueueLen),
		option:       option,
		onTaskRun:    consumer.OnTaskRun,
		onTaskFinish: consumer.OnTaskFinish,
	}
}

func (worker *QueueWorker[T]) CurrentWorkingWorker() int64 {
	return worker.workerCount.Load()
}

func (worker *QueueWorker[T]) QueueLen() int64 {
	return int64(len(worker.queue))
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
	if worker.option.MaxQueueLen > 0 && worker.QueueLen() >= worker.option.MaxQueueLen {
		return errors.New("task queue exceeding limits")
	}

	worker.queue <- task
	return nil
}

func (worker *QueueWorker[T]) doTask(task T) {
	worker.workerCount.Add(1)
	var err error

	for i := 0; i < worker.option.RetryOnFail; i++ {
		err = worker.onTaskRun(task)
		if err == nil {
			break
		}
	}
	worker.onTaskFinish(task, err)

	worker.workerCount.Add(-1)

}
