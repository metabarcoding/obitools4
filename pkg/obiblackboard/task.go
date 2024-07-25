package obiblackboard

type Task struct {
	Role     string
	Priority int
	Body     interface{}
}

func NewInitialTask() *Task {
	return &Task{
		Role:     "initial",
		Priority: 0,
		Body:     nil,
	}
}

func (task *Task) GetNext() *Task {
	t := NewInitialTask()
	t.Priority = task.Priority + 1
	return t
}
