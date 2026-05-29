package internal

func withRetry[T any](n int, fetch func() (T, error)) func() (T, error) {
	return func() (T, error) {
		var (
			val T
			err error
		)
		for range n {
			val, err = fetch()
			if err == nil {
				return val, nil
			}
		}
		return val, err
	}
}
