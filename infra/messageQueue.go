package infra


// MessageQueue 是一个简单的消息队列，使用通道实现
type MessageQueue[T any] struct {
	queue chan T
}

func NewMessageQueue[T any](max_worker int64 ) *MessageQueue[T] {
	return &MessageQueue[T]{
		queue: make(chan T, max_worker),
	}
}

func (q *MessageQueue[T]) Push(message T) {

	q.queue <- message
}

// Pop 从队列中取出消息（如果队列为空，会阻塞）
func (q *MessageQueue[T]) Pop() T {
	return <-q.queue
}

// Len 返回队列中消息的数量（注意：这不是并发安全的）
func (q *MessageQueue[T]) Len() int {
	return len(q.queue)
}