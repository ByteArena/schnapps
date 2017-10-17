package scheduler

import "time"

func pollUntil(poll func() bool) {
	for {
		if stop := poll(); stop {
			return
		}

		time.Sleep(2 * time.Millisecond)
	}
}
