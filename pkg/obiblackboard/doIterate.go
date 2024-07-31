package obiblackboard

import "git.metabarcoding.org/obitools/obitools4/obitools4/pkg/obiutils"

type Iteration[T any] struct {
	Index int
	Value T
}

// DoIterateSlice generates a DoTask function that iterates over a given slice and
// creates a new InitialTask for each element. The function takes in a slice of type
// T and a target string. It returns a DoTask function that can be used to execute
// the iteration. The DoTask function takes a Blackboard and a Task as input and
// returns a new Task. The Task's Role is set to the target string and its Body is
// set to an Iteration struct containing the index i and the element s[i] from the
// input slice. The iteration stops when the index i is equal to or greater than
// the length of the input slice.
//
// Parameters:
// - s: The slice of type T to iterate over.
// - target: The target string to set as the Task's Role.
//
// Return type:
// - DoTask: The DoTask function that can be used to execute the iteration.
func DoIterateSlice[T any](s []T, target string) DoTask {
	n := len(s)
	idx := obiutils.AtomicCounter()

	dt := func(bb *Blackboard, t *Task) *Task {
		i := idx()
		if i < n {
			nt := t.GetNext(target, false, false)
			nt.Body = Iteration[T]{i, s[i]}
			return nt
		}
		return nil
	}

	return dt
}

// DoCount generates a DoTask function that iterates over a given integer n and
// creates a new InitialTask for each iteration. The function takes in an integer n
// and a target string. It returns a DoTask function that can be used to execute
// the iteration. The DoTask function takes a Blackboard and a Task as input and
// returns a new Task. The Task's Role is set to the target string and its Body is
// set to the current iteration index i. The iteration stops when the index i is
// equal to or greater than the input integer n.
//
// Parameters:
// - n: The integer to iterate over.
// - target: The target string to set as the Task's Role.
//
// Return type:
// - DoTask: The DoTask function that can be used to execute the iteration.
func DoCount(n int, target string) DoTask {
	idx := obiutils.AtomicCounter()

	dt := func(bb *Blackboard, t *Task) *Task {
		i := idx()
		if i < n {
			nt := t.GetNext(target, false, false)
			nt.Body = i
			return nt
		}
		return nil
	}

	return dt
}
