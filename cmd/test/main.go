package main

import (
	"fmt"

	"git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiblackboard"
)

func r2(bb *obiblackboard.Blackboard, task *obiblackboard.Task) *obiblackboard.Task {
	fmt.Printf("value : %v\n", task.Body)
	return obiblackboard.NewInitialTask()
}

func rmul(bb *obiblackboard.Blackboard, task *obiblackboard.Task) *obiblackboard.Task {
	nt := task.GetNext()
	nt.Body = task.Body.(int) * 2
	return nt
}

func main() {

	black := obiblackboard.NewBlackBoard(20)

	black.RegisterRunner("todisplay", "final", r2)
	black.RegisterRunner("multiply", "todisplay", rmul)
	black.RegisterRunner("initial", "multiply", obiblackboard.DoCount(1000).RepeatTask(4))

	black.Run()
}
