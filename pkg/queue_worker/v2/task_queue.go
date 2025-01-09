package queue_worker

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type QueueWorkerOption struct {
	MaxWorker   int64
	MaxQueueLen int64
	RetryOnFail int
}

type WorkingTask struct {
	PriorityQueueNode[Task]
	Ctx        context.Context
	CancelFunc context.CancelFunc
}

type QueueWorker struct {
	option QueueWorkerOption
	queue  *PriorityQueue[Task]
	// workerCount atomic.Int64
	workingWorkerChan chan *WorkingTask
	wg                sync.WaitGroup
}

func NewQueueWorker(option QueueWorkerOption) QueueWorker {
	if option.MaxQueueLen == 0 {
		option.MaxQueueLen = 10
	}

	return QueueWorker{
		workingWorkerChan: make(chan *WorkingTask, option.MaxWorker),
		queue:             NewPriorityQueue[Task](),
		option:            option,
	}
}

func (worker *QueueWorker) CurrentWorkingWorker() int64 {
	return int64(len(worker.workingWorkerChan))
}

func (worker *QueueWorker) QueueLen() int64 {
	return int64(worker.queue.Length())
}

func (worker *QueueWorker) CancelPendingTask(id string) {
	node, exist := worker.queue.PopNode(id)
	if !exist {
		return
	}

	node.Value.OnTaskInterrupt(fmt.Errorf("TASK_CANCELED"))
}

func (worker *QueueWorker) CancelRunningTask(id string) {
	for workingTask := range worker.workingWorkerChan {
		if workingTask.ID == id {
			workingTask.CancelFunc()
			break
		}
	}
}

func (worker *QueueWorker) Run() {
	for {
		if worker.queue.Length() == 0 {
			continue
		}

		ctx, canelFunc := context.WithCancel(context.Background())
		workingTask := &WorkingTask{
			Ctx:        ctx,
			CancelFunc: canelFunc,
		}
		worker.workingWorkerChan <- workingTask

		node, exist := worker.queue.PopHead()
		if !exist {
			<-worker.workingWorkerChan
			continue
		}

		workingTask.PriorityQueueNode = node
		go func(ctx context.Context, currentNode PriorityQueueNode[Task]) {
			defer func() {
				<-worker.workingWorkerChan
			}()

			select {
			case <-ctx.Done():
				node.Value.OnTaskInterrupt(ctx.Err())
			default:
				node.Value.OnTaskRun(ctx)
			}
		}(ctx, node)

	}
}

func (worker *QueueWorker) AddTask(task Task, priority int) error {
	if worker.option.MaxQueueLen > 0 && worker.QueueLen() >= worker.option.MaxQueueLen {
		return errors.New("task queue exceeding limits")
	}

	worker.queue.Insert(task, priority)
	return nil
}
