package Collection

type ConcurrentQueue struct {
	ch chan string
}

func NewConcurrentQueue(capacity int) *ConcurrentQueue {
	return &ConcurrentQueue{
		ch: make(chan string, capacity),
	}
}

func (q *ConcurrentQueue) Enqueue(item string) {
	q.ch <- item
}

func (q *ConcurrentQueue) Dequeue() (string, bool) {
	select {
	case item := <-q.ch:
		return item, true
	default:
		return "", false
	}
}
