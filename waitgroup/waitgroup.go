//go:build !solution

package waitgroup

// A WaitGroup waits for a collection of goroutines to finish.
// The main goroutine calls Add to set the number of
// goroutines to wait for. Then each of the goroutines
// runs and calls Done when finished. At the same time,
// Wait can be used to block until all goroutines have finished.
type WaitGroup struct {
	counter chan int
	free    chan struct{}
}

// New creates WaitGroup.
func New() *WaitGroup {
	answer := WaitGroup{counter: make(chan int, 1), free: make(chan struct{}, 1)}
	answer.counter <- 0
	answer.free <- struct{}{}
	return &answer
}

// Add adds delta, which may be negative, to the WaitGroup counter.
// If the counter becomes zero, all goroutines blocked on Wait are released.
// If the counter goes negative, Add panics.
//
// Note that calls with a positive delta that occur when the counter is zero
// must happen before a Wait. Calls with a negative delta, or calls with a
// positive delta that start when the counter is greater than zero, may happen
// at any time.
// Typically this means the calls to Add should execute before the statement
// creating the goroutine or other event to be waited for.
// If a WaitGroup is reused to wait for several independent sets of events,
// new Add calls must happen after all previous Wait calls have returned.
// See the WaitGroup example.
func (wg *WaitGroup) Add(delta int) {
	count := <-wg.counter
	if count == 0 {
		<-wg.free
	}
	wg.counter <- count + delta
	if count+delta < 0 {
		panic("negative WaitGroup counter")
	}
}

// Done decrements the WaitGroup counter by one.
func (wg *WaitGroup) Done() {
	count := <-wg.counter
	if count <= 0 {
		panic("negative WaitGroup counter")
	}
	wg.counter <- count - 1
	if count == 1 {
		wg.free <- struct{}{}
	}
}

// Wait blocks until the WaitGroup counter is zero.
func (wg *WaitGroup) Wait() {
	<-wg.free
	wg.free <- struct{}{}
}
