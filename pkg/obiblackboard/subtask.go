package obiblackboard

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// RepeatTask creates a new DoTask function that repeats the given task n times.
//
// It takes an integer n as input, which specifies the number of times the task should be repeated.
// It returns a new DoTask function that can be used to execute the repeated task.
//
// The returned DoTask function maintains a map of tasks to their counts and tasks.
// When a task is executed, it checks if the task has been executed before.
// If it has, it increments the count and returns the previously executed task.
// If it has not been executed before, it executes the task using the provided runner function.
// If the runner function returns nil, the task is not added to the task memory and nil is returned.
// If the runner function returns a non-nil task, it is added to the task memory with a count of 0.
// After executing the task, the function checks if the count is less than (n-1).
// If it is, the task is added back to the blackboard to be executed again.
// If the count is equal to (n-1), the task is removed from the task memory.
// Finally, the function returns the executed task.
func (runner DoTask) RepeatTask(n int) DoTask {
	type memtask struct {
		count int
		task  *Task
	}
	taskMemory := make(map[*Task]*memtask)
	taskMemoryLock := sync.Mutex{}

	if n < 1 {
		log.Fatalf("Cannot repeat a task less than once (n=%d)", n)
	}

	st := func(bb *Blackboard, task *Task) *Task {
		taskMemoryLock.Lock()

		mem, ok := taskMemory[task]

		if !ok {
			nt := runner(bb, task)

			if nt == nil {
				taskMemoryLock.Unlock()
				return nt
			}

			mem = &memtask{
				count: 0,
				task:  nt,
			}

			taskMemory[task] = mem
		} else {
			mem.count++
		}

		taskMemoryLock.Unlock()

		if mem.count < (n - 1) {
			bb.PushTask(task)
		}

		if mem.count == (n - 1) {
			taskMemoryLock.Lock()
			delete(taskMemory, task)
			taskMemoryLock.Unlock()
		}

		return mem.task
	}

	return st
}

// CombineWith returns a new DoTask function that combines the given DoTask
// functions. The returned function applies the `other` function to the result
// of the `runner` function. The `bb` parameter is the Blackboard instance,
// and the `task` parameter is the Task instance.
//
// Parameters:
// - bb: The Blackboard instance.
// - task: The Task instance.
//
// Returns:
//   - *Task: The result of applying the `other` function to the result of the
//     `runner` function.
func (runner DoTask) CombineWith(other DoTask) DoTask {
	return func(bb *Blackboard, task *Task) *Task {
		return other(bb, runner(bb, task))
	}
}

// SetTarget sets the target role for the task.
//
// Parameters:
// - target: The target role to set.
//
// Returns:
// - DoTask: The modified DoTask function.
func (runner DoTask) SetTarget(target string) DoTask {
	return func(bb *Blackboard, task *Task) *Task {
		nt := runner(bb, task)
		nt.Role = target
		return nt
	}
}
