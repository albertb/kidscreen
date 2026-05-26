package internal

import "sync"

// newLazy returns a function that calls fetch exactly once and caches the result.
func newLazy[T any](fetch func() (T, error)) func() (T, error) {
	var once sync.Once
	var val T
	var err error
	return func() (T, error) {
		once.Do(func() {
			val, err = fetch()
		})
		return val, err
	}
}
