package queue

import (
	"container/heap"
)

type PriorityItem interface {
	Higher(interface{}) bool
}

type priorityHeap []PriorityItem

func (h *priorityHeap) Len() int {
	return len(*h)
}

func (h *priorityHeap) Less(i, j int) bool {
	return (*h)[i].Higher((*h)[j])
}

func (h *priorityHeap) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
}

func (h *priorityHeap) Push(x interface{}) {
	*h = append(*h, x.(PriorityItem))
}

func (h *priorityHeap) Pop() interface{} {
	old := *h
	n := len(old)

	if n > 0 {
		item := old[n-1]
		*h = old[0 : n-1]
		return item
	}

	return nil
}

type PriorityQueue struct {
	sorter *priorityHeap
}

func NewPriorityQueue() *PriorityQueue {
	q := &PriorityQueue{
		sorter: new(priorityHeap),
	}

	heap.Init(q.sorter)

	return q
}

func (q *PriorityQueue) Push(x PriorityItem) {
	heap.Push(q.sorter, x)
}

func (q *PriorityQueue) Pop() PriorityItem {
	return heap.Pop(q.sorter).(PriorityItem)
}

func (q *PriorityQueue) Top() PriorityItem {
	if len(*q.sorter) > 0 {
		return (*q.sorter)[0].(PriorityItem)
	}

	return nil
}

func (q *PriorityQueue) Fix(x PriorityItem, i int) {
	(*q.sorter)[i] = x
	heap.Fix(q.sorter, i)
}

func (q *PriorityQueue) Remove(i int) PriorityItem {
	return heap.Remove(q.sorter, i).(PriorityItem)
}

func (q *PriorityQueue) Len() int {
	return q.sorter.Len()
}
