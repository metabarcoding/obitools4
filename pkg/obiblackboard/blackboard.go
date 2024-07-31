package obiblackboard

import (
	"slices"
	"sync"
	"time"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type DoTask func(*Blackboard, *Task) *Task

type Blackboard struct {
	Board      map[int]Queue
	BoardLock  *sync.Mutex
	Runners    map[string]DoTask
	Running    *obiutils.Counter
	TargetSize int
	Size       int
}

func doFinal(bb *Blackboard, task *Task) *Task {
	if task.SavedTask != nil {
		return task.SavedTask
	}

	return NewInitialTask()
}

func NewBlackBoard(size int) *Blackboard {
	board := make(map[int]Queue, 0)
	runners := make(map[string]DoTask, 0)

	if size < 2 {
		size = 2
	}

	bb := &Blackboard{
		Board:      board,
		BoardLock:  &sync.Mutex{},
		Runners:    runners,
		Running:    obiutils.NewCounter(),
		TargetSize: size,
		Size:       0,
	}

	for i := 0; i < size; i++ {
		bb.PushTask(NewInitialTask())
	}

	bb.RegisterRunner("final", doFinal)

	return bb
}

func (bb *Blackboard) RegisterRunner(target string, runner DoTask) {
	bb.Runners[target] = runner
}

func (bb *Blackboard) MaxQueue() Queue {
	max_priority := -1
	max_queue := Queue(nil)
	for priority, queue := range bb.Board {
		if priority > max_priority {
			max_queue = queue
		}
	}

	return max_queue
}

func (bb *Blackboard) PopTask() *Task {
	bb.BoardLock.Lock()
	defer bb.BoardLock.Unlock()

	q := bb.MaxQueue()

	if q != nil {
		next_task := (*q)[0]
		(*q) = (*q)[1:]
		if len(*q) == 0 {
			delete(bb.Board, next_task.Priority)
		}
		bb.Size--
		return next_task
	}

	return (*Task)(nil)
}

func (bb *Blackboard) PushTask(task *Task) {
	bb.BoardLock.Lock()
	defer bb.BoardLock.Unlock()

	if task != nil {
		priority := task.Priority
		queue, ok := bb.Board[priority]

		if !ok {
			queue = NewQueue()
			bb.Board[priority] = queue
		}

		*queue = slices.Grow(*queue, 1)
		*queue = append((*queue), task)

		bb.Size++
	}
}

func (bb *Blackboard) Run() {

	ctask := make(chan *Task)
	lock := &sync.WaitGroup{}

	launcher := func() {
		for task := range ctask {
			runner, ok := bb.Runners[task.Role]

			if ok {
				task = runner(bb, task)
			}

			bb.PushTask(task)
			bb.Running.Dec()
		}

		lock.Done()
	}

	parallel := bb.TargetSize - 1
	lock.Add(parallel)

	for i := 0; i < parallel; i++ {
		go launcher()
	}

	go func() {

		for {
			bb.Running.Inc()
			task := bb.PopTask()

			if task != nil {
				ctask <- task
			} else {
				bb.Running.Dec()
				if bb.Running.Value()+bb.Len() <= 0 {
					break
				}
				time.Sleep(time.Millisecond)
			}

		}

		close(ctask)
	}()

	lock.Wait()
}

// func (bb *Blackboard) Run() {
// 	lock := &sync.WaitGroup{}

// 	launcher := func(runner DoTask, task *Task) {
// 		task = runner(bb, task)

// 		if task != nil {
// 			for bb.Len() > bb.TargetSize {
// 				time.Sleep(time.Millisecond)
// 			}
// 			bb.PushTask(task)
// 		}

// 		bb.Running.Dec()
// 		lock.Done()
// 	}

// 	lock.Add(1)

// 	func() {
// 		for bb.Len()+bb.Running.Value() > 0 {
// 			bb.Running.Inc()
// 			task := bb.PopTask()

// 			if task != nil {
// 				lock.Add(1)
// 				go launcher(bb.Runners[task.Role], task)
// 			} else {
// 				bb.Running.Dec()
// 			}
// 		}

// 		lock.Done()
// 	}()

// 	lock.Wait()
// }

func (bb *Blackboard) Len() int {
	return bb.Size
}

// 151431044 151431044 15083822152
