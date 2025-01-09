package queue_worker

import "context"

// type QueueWorkerConsumer[T any] interface {
// 	OnTaskRun(task *T) error
// 	OnTaskFinish(task *T, err error)
// 	OnTaskCancel(task *T) 
// }


var (

	

)


type Task interface {
	OnTaskRun(ctx context.Context)
	// OnTaskFinish(ctx context.Context, err error)
	OnTaskInterrupt(cause error) 

}


