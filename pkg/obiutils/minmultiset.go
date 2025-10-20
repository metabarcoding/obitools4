package obiutils

import (
	"container/heap"
)

// MinMultiset maintient un multiset de valeurs et expose le minimum courant.
// T doit être comparable pour servir de clé de map. L'ordre est défini par less.
type MinMultiset[T comparable] struct {
	pq      priorityQueue[T] // tas min
	less    func(a, b T) bool
	count   map[T]int // cardinalité logique par valeur
	pending map[T]int // suppressions en attente (lazy delete)
	size    int       // taille logique totale
}

// New crée un multiset. less doit imposer un ordre strict total.
func NewMinMultiset[T comparable](less func(a, b T) bool) *MinMultiset[T] {
	m := &MinMultiset[T]{
		pq:      priorityQueue[T]{less: less},
		less:    less,
		count:   make(map[T]int),
		pending: make(map[T]int),
	}
	heap.Init(&m.pq)
	return m
}

// Add ajoute une occurrence.
func (m *MinMultiset[T]) Add(v T) {
	heap.Push(&m.pq, v)
	m.count[v]++
	m.size++
}

// RemoveOne retire UNE occurrence de v. Retourne false si absente.
func (m *MinMultiset[T]) RemoveOne(v T) bool {
	if m.count[v] == 0 {
		return false
	}
	m.count[v]--
	m.pending[v]++
	m.size--
	m.shrink()
	return true
}

// Min retourne le minimum courant. ok=false si vide.
func (m *MinMultiset[T]) Min() (v T, ok bool) {
	if m.size == 0 {
		var zero T
		return zero, false
	}
	m.cleanTop()
	return m.pq.data[0], true
}

// Len retourne la taille logique.
func (m *MinMultiset[T]) Len() int { return m.size }

// --- interne ---

// retire du sommet toutes les valeurs marquées à supprimer.
func (m *MinMultiset[T]) cleanTop() {
	for m.pq.Len() > 0 {
		top := m.pq.data[0]
		if m.pending[top] > 0 {
			m.pending[top]--
			if m.pending[top] == 0 {
				delete(m.pending, top) // ← nettoyage de la map
			}
			heap.Pop(&m.pq)
			continue
		}
		break
	}
}

// rééquilibre le tas si trop de tombstones.
func (m *MinMultiset[T]) shrink() {
	// nettoyage léger au retrait pour borner la dérive
	if m.pq.Len() > 0 {
		m.cleanTop()
	}
}

// priorityQueue implémente heap.Interface pour T.
type priorityQueue[T any] struct {
	data []T
	less func(a, b T) bool
}

func (q priorityQueue[T]) Len() int           { return len(q.data) }
func (q priorityQueue[T]) Less(i, j int) bool { return q.less(q.data[i], q.data[j]) }
func (q priorityQueue[T]) Swap(i, j int)      { q.data[i], q.data[j] = q.data[j], q.data[i] }
func (q *priorityQueue[T]) Push(x any)        { q.data = append(q.data, x.(T)) }
func (q *priorityQueue[T]) Pop() any {
	n := len(q.data)
	x := q.data[n-1]
	q.data = q.data[:n-1]
	return x
}
func (q priorityQueue[T]) peek() (T, bool) {
	if len(q.data) == 0 {
		var z T
		return z, false
	}
	return q.data[0], true
}
func (q *priorityQueue[T]) Top() (T, bool) { return q.peek() }
func (q *priorityQueue[T]) PushValue(v T)  { heap.Push(q, v) }
func (q *priorityQueue[T]) PopValue() (T, bool) {
	if q.Len() == 0 {
		var z T
		return z, false
	}
	return heap.Pop(q).(T), true
}
