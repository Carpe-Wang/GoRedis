package wait

import (
	"sync"
	"time"
)

// Wait 与sync.WaitGroup类似，可以超时等待
type Wait struct {
	wg sync.WaitGroup
}

// Add 将增量（可能为负数）添加到WaitGroup计数器。
func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

// Done 将WaitGroup计数器减少一
func (w *Wait) Done() {
	w.wg.Done()
}

// Wait 直到WaitGroup计数器为零。
func (w *Wait) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout 阻塞，直到WaitGroup计数器为零或超时
// returns 如果超时return ture
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan bool, 1)
	go func() {
		defer close(c)
		w.wg.Wait()
		c <- true
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
