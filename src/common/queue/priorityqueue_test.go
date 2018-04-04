package queue

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type Item struct {
	id        string
	priority  int
	createdAt time.Time
}

func (i *Item) Higher(other interface{}) bool {
	j := other.(*Item)

	if i.priority == j.priority {
		return i.createdAt.Before(j.createdAt)
	}

	return i.priority > j.priority
}

func Test_PriorityQueue(t *testing.T) {
	q := NewPriorityQueue()

	count := 10

	for i := 0; i < count; i++ {
		q.Push(&Item{
			id:        fmt.Sprintf("item-%d", i),
			priority:  rand.Intn(count),
			createdAt: time.Now(),
		})
	}

	for i := 0; i < count; i++ {
		item := q.Pop().(*Item)
		t.Log(item.id, item.priority, item.createdAt)
	}

	left := q.Len()
	if left > 0 {
		t.Error("No item should be left")
	}
}
