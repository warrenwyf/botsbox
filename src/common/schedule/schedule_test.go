package schedule

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

var s = NewSchedule()

type TestTask struct {
	title    string
	fn       func()
	interval time.Duration
	delay    time.Duration
}

func (t *TestTask) GetTitle() string {
	return t.title
}

func (t *TestTask) GetFn() func() {
	return t.fn
}

func (t *TestTask) GetInterval() time.Duration {
	return t.interval
}

func (t *TestTask) GetDelay() time.Duration {
	return t.delay
}

func Test_CreateAndDeleteTask(t *testing.T) {
	s.Clear()

	s.Start()

	id1 := s.CreateTask(&TestTask{
		title: "t1",
		fn: func() {
			t.Log(time.Now(), "Task 1 is executed")
		},
		interval: 2 * time.Second,
		delay:    500 * time.Millisecond,
	})

	id2 := s.CreateTask(&TestTask{
		title: "t2",
		fn: func() {
			t.Log(time.Now(), "Task 2 is executed")
		},
		interval: 4 * time.Second,
		delay:    500 * time.Millisecond,
	})

	time.Sleep(10 * time.Second)

	t.Log("Delete task 1 with id", id1, s.DeleteTask(id1))
	t.Log("Delete task 2 with id", id2, s.DeleteTask(id2))

	s.Stop()
}

func Test_TaskExecuted(t *testing.T) {
	s.Clear()

	s.Start()

	var executed int32 = 0

	t.Log("Start add tasks", time.Now())
	count := 1000
	for i := 0; i < count; i++ {

		s.CreateTask(&TestTask{
			title: "",
			fn: func() {
				atomic.AddInt32(&executed, 1)
			},
			interval: time.Hour,
			delay:    time.Duration(rand.Intn(10)) * time.Second,
		})

	}
	t.Log("Finish add tasks", time.Now())

	time.Sleep(12 * time.Second)

	allCount := len(s.AllTasks())

	if allCount != count {
		t.Errorf("%d tasks created, should be %d", allCount, count)
	}

	if int32(allCount) != executed {
		t.Errorf("%d tasks executed, total %d", executed, allCount)
	}

	s.Stop()
}

func Benchmark_CreateTask(b *testing.B) {
	s.Clear()

	s.Start()

	count := 10
	for i := 0; i < count; i++ {
		title := fmt.Sprintf("t%d", i)
		log := fmt.Sprintf("BenchTask %d is executed", i)

		id := s.CreateTask(&TestTask{
			title: title,
			fn: func() {
				b.Log(time.Now(), log)
			},
			interval: time.Second,
			delay:    time.Duration(rand.Intn(10)) * time.Second,
		})

		b.Logf("BenchTask %d has id %d ", i, id)
	}

	time.Sleep(10 * time.Second)

	s.Stop()
}
