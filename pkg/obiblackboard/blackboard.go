package obiblackboard

import (
	"slices"
	"sync"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

type DoTask func(*Blackboard, *Task) *Task

type Runner struct {
	Run DoTask
	To  string
}

type Blackboard struct {
	Board      map[int]Queue
	BoardLock  *sync.Mutex
	Runners    map[string]Runner
	Running    *obiutils.Counter
	TargetSize int
	Size       int
}

func doFinal(bb *Blackboard, task *Task) *Task {
	if bb.Len() > bb.TargetSize {
		return nil
	}

	return NewInitialTask()
}

func NewBlackBoard(size int) *Blackboard {
	board := make(map[int]Queue, 0)
	runners := make(map[string]Runner, 0)

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

	bb.RegisterRunner("final", "initial", doFinal)

	return bb
}

func (bb *Blackboard) RegisterRunner(from, to string, runner DoTask) {
	bb.Runners[from] = Runner{
		Run: runner,
		To:  to,
	}
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
				task = runner.Run(bb, task)
				if task != nil {
					task.Role = runner.To
				}
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
				if bb.Running.Value() <= 0 {
					break
				}
			}

		}

		close(ctask)
	}()

	lock.Wait()
}

func (bb *Blackboard) Len() int {
	return bb.Size
}
