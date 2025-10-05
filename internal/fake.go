package internal

import (
	"math"
	"math/rand"
)

// Returns n random integers between min and max, biased towards min, with smooth transitions.
func getBiasedSmoothRandomValues(n, min, max int) []int {
	values := make([]int, n)

	for i := range values {
		// Prefer lower values.
		bias := rand.Float64()
		target := min + int(math.Pow(bias, 4)*float64(max-min))

		next := target
		if i > 0 {
			// Smooth the change from one value to the next.
			step := rand.Intn(10) - 5
			next = values[i-1] + step
			next = (next*3 + target) / 4
		}

		if next < min {
			next = min
		}
		if next > max {
			next = max
		}
		values[i] = next
	}

	return values
}
