package Collection

import (
	"sync"
)

type ConcurrentMapOfQueues struct {
	queues   sync.Map
	capacity int
}

func NewConcurrentMapOfQueues(queueCapacity int) *ConcurrentMapOfQueues {
	return &ConcurrentMapOfQueues{
		capacity: queueCapacity,
	}
}

func (cm *ConcurrentMapOfQueues) GetOrCreateQueue(key string) *ConcurrentQueue {
	queue, ok := cm.queues.Load(key)
	if !ok {
		newQueue := NewConcurrentQueue(cm.capacity)
		actual, _ := cm.queues.LoadOrStore(key, newQueue)
		queue = actual
	}
	return queue.(*ConcurrentQueue)
}

func (cm *ConcurrentMapOfQueues) Enqueue(key string, item string) {
	queue := cm.GetOrCreateQueue(key)
	queue.Enqueue(item)
}

func (cm *ConcurrentMapOfQueues) Dequeue(key string) (string, bool) {
	if queue, ok := cm.queues.Load(key); ok {
		return queue.(*ConcurrentQueue).Dequeue()
	}
	return "", false
}
