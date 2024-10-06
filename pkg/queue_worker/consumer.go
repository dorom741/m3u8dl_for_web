package queue_worker

type QueueWorkerConsumer[T any] interface {
	OnTaskRun(task T) error
	OnTaskFinish(task T, err error)
}
