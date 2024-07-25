package obiblackboard

type Queue *[]*Task

func NewQueue() Queue {
	q := make([]*Task, 0)
	return &q
}
