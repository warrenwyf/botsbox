package schedule

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

var s = New()

func Test_Start(t *testing.T) {
	s.Start()
}

func Test_CreateTask(t *testing.T) {
	id1 := s.CreateTask(func() {
		fmt.Println(time.Now(), "Task 1 is executed")
	}, 2*time.Second, 0)
	id2 := s.CreateTask(func() {
		fmt.Println(time.Now(), "Task 2 is executed")
	}, 4*time.Second, 0)

	time.Sleep(10 * time.Second)

	t.Log("Delete task 1 result", s.DeleteTask(id1))
	t.Log("Delete task 2 result", s.DeleteTask(id2))
}

func Test_Stop(t *testing.T) {
	s.Stop()
}

func Benchmark_CreateTask(b *testing.B) {
	s.Start()

	count := 10

	for i := 0; i < count; i++ {
		log := fmt.Sprintf("BenchTask %d is executed", i)
		id := s.CreateTask(func() {
			fmt.Println(time.Now(), log)
		}, time.Second, time.Duration(rand.Intn(10))*time.Second)

		b.Logf("BenchTask %d has id %d ", i, id)
	}

	time.Sleep(10 * time.Second)

	s.Stop()
}
