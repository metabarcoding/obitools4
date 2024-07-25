package main

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiblackboard"
	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"
)

func MakeCounter(n int) func(*obiblackboard.Task) *obiblackboard.Task {
	count := obiutils.AtomicCounter()

	r1 := func(task *obiblackboard.Task) *obiblackboard.Task {
		val := count()
		if val < n {
			nt := task.GetNext()
			nt.Body = val
			return nt
		}
		return nil
	}

	return r1
}

func r2(task *obiblackboard.Task) *obiblackboard.Task {
	fmt.Printf("value : %v\n", task.Body)
	return obiblackboard.NewInitialTask()
}

func rmul(task *obiblackboard.Task) *obiblackboard.Task {
	nt := task.GetNext()
	nt.Body = task.Body.(int) * 2
	return nt
}

func main() {

	black := obiblackboard.NewBlackBoard(20)

	black.RegisterRunner("todisplay", "initial", r2)
	black.RegisterRunner("multiply", "todisplay", rmul)
	black.RegisterRunner("initial", "multiply", MakeCounter(1000))

	black.Run()
}
