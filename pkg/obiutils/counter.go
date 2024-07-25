package obiutils

import "sync"

type Counter struct {
	Inc   func() int
	Dec   func() int
	Value func() int
}

func NewCounter(initial ...int) *Counter {
	value := 0
	lock := sync.Mutex{}

	if len(initial) > 0 {
		value = initial[0]
	}

	c := Counter{
		Inc: func() int {
			lock.Lock()
			defer lock.Unlock()
			value++
			return value
		},

		Dec: func() int {
			lock.Lock()
			defer lock.Unlock()
			value--
			return value
		},

		Value: func() int {
			lock.Lock()
			defer lock.Unlock()
			return value
		},
	}

	return &c
}
