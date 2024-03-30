//go:build !solution

package cond

// A Locker represents an object that can be locked and unlocked.
type Locker interface {
	Lock()
	Unlock()
}

// Cond implements a condition variable, a rendezvous point
// for goroutines waiting for or announcing the occurrence
// of an event.
//
// Each Cond has an associated Locker L (often a *sync.Mutex or *sync.RWMutex),
// which must be held when changing the condition and
// when calling the Wait method.
type Cond struct {
	L       Locker
	free    chan struct{}
	counter chan int
}

// New returns a new Cond with Locker l.
func New(l Locker) *Cond {
	answer := Cond{L: l, free: make(chan struct{}, 1), counter: make(chan int, 1)}
	answer.counter <- 0
	return &answer
}

// Wait atomically unlocks c.L and suspends execution
// of the calling goroutine. After later resuming execution,
// Wait locks c.L before returning. Unlike in other systems,
// Wait cannot return unless awoken by Broadcast or Signal.
//
// Because c.L is not locked when Wait first resumes, the caller
// typically cannot assume that the condition is true when
// Wait returns. Instead, the caller should Wait in a loop:
//
//	c.L.Lock()
//	for !condition() {
//	    c.Wait()
//	}
//	... make use of condition ...
//	c.L.Unlock()
func (c *Cond) Wait() {
	c.L.Unlock()
	count := <-c.counter
	c.counter <- count + 1
	<-c.free
	c.L.Lock()
}

// Signal wakes one goroutine waiting on c, if there is any.
//
// It is allowed but not required for the caller to hold c.L
// during the call.
func (c *Cond) Signal() {
	count := <-c.counter
	c.counter <- count - 1
	c.free <- struct{}{}
}

// Broadcast wakes all goroutines waiting on c.
//
// It is allowed but not required for the caller to hold c.L
// during the call.
func (c *Cond) Broadcast() {
	for count := <-c.counter; count > 0; count = <-c.counter {
		c.counter <- count - 1
		c.free <- struct{}{}
	}
	c.counter <- 0
}
