package obiblackboard

import "fmt"

func DisplayTask(bb *Blackboard, task *Task) *Task {
	if task == nil {
		return nil
	}

	fmt.Printf("Task: %s:\n%v\n\n", task.Role, task.Body)

	return task
}

func (runner DoTask) Display() DoTask {
	return runner.CombineWith(DisplayTask)
}
