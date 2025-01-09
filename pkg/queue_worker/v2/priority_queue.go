package queue_worker

import (
	"sync"

	"github.com/google/uuid"
)

// PriorityQueueNode 表示链表中的一个节点
type PriorityQueueNode[T any] struct {
	ID       string
	Value    T
	Priority int
	Next     *PriorityQueueNode[T]
}

// PriorityQueue 表示优先级链表
type PriorityQueue[T any] struct {
	head *PriorityQueueNode[T]
	mu   sync.RWMutex // 用于保护链表的互斥锁
	len  int
}

// NewPriorityQueue 创建一个新的优先级链表
func NewPriorityQueue[T any]() *PriorityQueue[T] {
	return &PriorityQueue[T]{}
}

// Insert 插入一个新节点到优先级链表中
func (pq *PriorityQueue[T]) Insert(value T, priority int) string {
	pq.mu.Lock()         // 加锁
	defer pq.mu.Unlock() // 解锁

	newPriorityQueueNode := &PriorityQueueNode[T]{ID: uuid.New().String(), Value: value, Priority: priority}
	if pq.head == nil || pq.head.Priority < priority {
		newPriorityQueueNode.Next = pq.head
		pq.head = newPriorityQueueNode
		return newPriorityQueueNode.ID
	}

	current := pq.head
	for current.Next != nil && current.Next.Priority >= priority {
		current = current.Next
	}
	newPriorityQueueNode.Next = current.Next
	current.Next = newPriorityQueueNode
	pq.len++
	return newPriorityQueueNode.ID
}

// Length 返回链表的长度
func (pq *PriorityQueue[T]) Length() int {
	pq.mu.RLock()         // 读锁
	defer pq.mu.RUnlock() // 解锁


	return pq.len
}

// PopHead 移除并返回链表头部的节点
func (pq *PriorityQueue[T]) PopHead() (PriorityQueueNode[T], bool) {
	pq.mu.Lock()         // 加锁
	defer pq.mu.Unlock() // 解锁

	if pq.head == nil {
		return PriorityQueueNode[T]{}, false
	}
	node := *pq.head
	pq.head = pq.head.Next

	pq.len--
	return node, true
}

// Traverse 遍历链表并执行给定的操作
func (pq *PriorityQueue[T]) Traverse(action func(node PriorityQueueNode[T])) {
	pq.mu.RLock()         // 读锁
	defer pq.mu.RUnlock() // 解锁

	current := pq.head
	for current != nil {
		action(*current)
		current = current.Next
	}
}

// RemoveNode 根据 ID 移除任意节点
func (pq *PriorityQueue[T]) PopNode(id string) (PriorityQueueNode[T], bool) {
	pq.mu.Lock()         // 加锁
	defer pq.mu.Unlock() // 解锁

	if pq.head == nil {
		return PriorityQueueNode[T]{}, false
	}
	pq.len--

	
	// 特殊情况：移除头部节点
	if pq.head.ID == id {
		node := *pq.head
		pq.head = pq.head.Next
		return node, true
	}

	current := pq.head
	for current.Next != nil {
		if current.Next.ID == id {
			// 找到目标节点，移除它
			node := *current.Next
			current.Next = current.Next.Next
			return node, true
		}
		current = current.Next
	}

	return PriorityQueueNode[T]{}, false
}
