package queue_worker

import (
	"errors"
)

type QueueWorkerOption struct {
	MaxWorker   int64
	MaxQueueLen int64
	RetryOnFail int
}

type QueueWorker[T any] struct {
	option QueueWorkerOption
	queue  chan T
	//workerCount atomic.Int64
	workingWorkerChan chan struct{}

	onTaskRun    func(task T) error
	onTaskFinish func(task T, err error)
}

func NewQueueWorker[T any](option QueueWorkerOption, consumer QueueWorkerConsumer[T]) QueueWorker[T] {

	if option.MaxQueueLen == 0 {
		option.MaxQueueLen = 10
	}

	return QueueWorker[T]{
		workingWorkerChan: make(chan struct{}, option.MaxWorker),
		queue:             make(chan T, option.MaxQueueLen),
		option:            option,
		onTaskRun:         consumer.OnTaskRun,
		onTaskFinish:      consumer.OnTaskFinish,
	}
}

func (worker *QueueWorker[T]) CurrentWorkingWorker() int64 {
	return int64(len(worker.workingWorkerChan))
}

func (worker *QueueWorker[T]) QueueLen() int64 {
	return int64(len(worker.queue))
}

func (worker *QueueWorker[T]) Run() {

	for {
		worker.workingWorkerChan <- struct{}{}
		task := <-worker.queue
		go func() {
			defer func() {
				<-worker.workingWorkerChan
			}()
			worker.doTask(task)

		}()

	}

}

func (worker *QueueWorker[T]) AddTask(task T) error {
	if worker.option.MaxQueueLen > 0 && worker.QueueLen() >= worker.option.MaxQueueLen {
		return errors.New("task queue exceeding limits")
	}

	worker.queue <- task
	return nil
}

func (worker *QueueWorker[T]) AddTaskBlocking(task T)  {
	worker.queue <- task
}


func (worker *QueueWorker[T]) doTask(task T) {
	var err error

	for i := 0; i < worker.option.RetryOnFail; i++ {
		err = worker.onTaskRun(task)
		if err == nil {
			break
		}
	}
	worker.onTaskFinish(task, err)

}
