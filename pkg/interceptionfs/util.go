package interceptionfs

import (
	"time"
)

func Uint64Max(x uint64, y uint64) uint64 {
	if x > y {
		return x
	}
	return y
}

func StopAndDrain(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func DebouncedNotifier(finally func(Node), d time.Duration) func(Node) {
	var p Node
	pp := &p

	t := time.AfterFunc(d, func() {
		finally(*pp)
	})
	StopAndDrain(t)

	return func(n Node) {
		StopAndDrain(t)
		*pp = n
		t.Reset(d)
	}
}
