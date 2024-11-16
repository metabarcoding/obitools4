package obitax

import "sync"

// InnerString is a struct that holds a map of strings and a read-write lock for concurrent access.
// The index map is used to store key-value pairs of strings.
type InnerString struct {
	index map[string]*string // Map to store string values
	lock  sync.RWMutex       // Read-write lock for concurrent access
}

// NewInnerString creates a new instance of InnerString.
// It initializes the index map and prepares the lock for use.
func NewInnerString() *InnerString {
	return &InnerString{
		index: make(map[string]*string),
	}
}

// Innerize stores the given value in the index map if it is not already present.
// It returns the pointer to the string associated with the key, which is either the newly stored value
// or the existing value if it was already present in the map.
//
// Parameters:
//   - value: The string value to be stored in the index map.
//
// Returns:
//   - A pointer to the string value associated with the key.
func (i *InnerString) Innerize(value string) *string {
	i.lock.Lock()
	defer i.lock.Unlock()
	s, ok := i.index[value]
	if !ok {
		s = &value
		i.index[value] = s
	}

	return s
}

// Slice returns a slice of strings containing all the values stored in the index map.
func (i *InnerString) Slice() []string {
	rep := make([]string, len(i.index))
	j := 0
	for _, v := range i.index {
		rep[j] = *v
		j++
	}
	return rep
}
