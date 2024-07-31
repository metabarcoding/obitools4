package obiblackboard

type Task struct {
	Role      string
	SavedTask *Task
	Priority  int
	Body      interface{}
}

func NewInitialTask() *Task {
	return &Task{
		Role:      "initial",
		SavedTask: nil,
		Priority:  0,
		Body:      nil,
	}
}

func (task *Task) GetNext(target string, copy bool, save bool) *Task {
	t := NewInitialTask()
	t.Priority = task.Priority + 1
	t.Role = target
	if copy {
		t.Body = task.Body
	}

	if save {
		t.SavedTask = task
	} else {
		t.SavedTask = task.SavedTask
	}

	return t
}
