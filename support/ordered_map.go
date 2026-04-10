package support

type OrderedMap[T any, K comparable] struct {
	sl []T
	m  map[K]int
}

func NewOrderedMap[T any, K comparable]() *OrderedMap[T, K] {
	return &OrderedMap[T, K]{
		sl: make([]T, 0),
		m:  make(map[K]int),
	}
}

// func (m *OrderedMap[T, K]) Set(key K, val T) {
// 	if idx, exists := m.m[key]; exists {
// 		m.sl[idx] = val
// 		return
// 	}
// 	m.m[key] = len(m.sl)
// 	m.sl = append(m.sl, val)
// }

func (m *OrderedMap[T, K]) Add(key K, val T) bool {
	if _, exists := m.m[key]; exists {
		return false
	}
	m.m[key] = len(m.sl)
	m.sl = append(m.sl, val)
	return true
}

func (m *OrderedMap[T, K]) Get(key K) (T, bool) {
	var zero T
	idx, ok := m.m[key]
	if !ok {
		return zero, false
	}
	return m.sl[idx], true
}

// func (m *OrderedMap[T, K]) Delete(key K) {
// 	idx, exists := m.m[key]
// 	if !exists {
// 		return
// 	}
//
// 	m.sl = append(m.sl[:idx], m.sl[idx+1:]...)
// 	delete(m.m, key)
//
// 	for k, i := range m.m {
// 		if i > idx {
// 			m.m[k] = i - 1
// 		}
// 	}
// }

func (m *OrderedMap[T, K]) Slice() []T {
	return m.sl
}

func (m *OrderedMap[T, K]) Len() int {
	return len(m.sl)
}
