package requester

import "time"

func Elapsed(fn func()) time.Duration {
	t1 := time.Now()
	fn()
	return time.Since(t1)
}
