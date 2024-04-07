//go:build !solution

package tparallel

import (
	"sync"
	"sync/atomic"
)

type T struct {
	parent           *T
	parallelSubtests []*T
	parallelBlocker  chan struct{}
	parallelSignal   chan struct{}
	done             chan struct{}
	isParallel       atomic.Bool
	waitGroup        sync.WaitGroup
	task             func(t *T)
}

// Parallel is not allowed to be called twice(while already parallel)
func (t *T) Parallel() {
	t.isParallel.Store(true)
	t.parent.parallelSubtests = append(t.parent.parallelSubtests, t)
	t.parent.parallelSignal <- struct{}{}
	<-t.parallelBlocker
}

func (t *T) Run(subtest func(t *T)) {
	newNode := &T{parent: t, parallelBlocker: make(chan struct{}), parallelSignal: make(chan struct{}), done: make(chan struct{}, 1), task: subtest}
	newNode.isParallel.Store(false)
	DoTask(newNode)
}

func DoTask(t *T) {
	go func() {
		t.task(t)
		t.waitGroup.Add(len(t.parallelSubtests))
		for _, parallelSubtest := range t.parallelSubtests {
			parallelSubtest.parallelBlocker <- struct{}{}
		}
		t.waitGroup.Wait()
		if t.isParallel.Load() {
			t.parent.waitGroup.Done()
		}
		t.done <- struct{}{}
	}()
	select {
	case <-t.done:
	case <-t.parent.parallelSignal:
	}
}

func Run(topTests []func(t *T)) {
	root := &T{parallelSignal: make(chan struct{}), parallelBlocker: make(chan struct{}), done: make(chan struct{}, 1)}
	root.isParallel.Store(false)
	root.Run(func(t *T) {
		for _, task := range topTests {
			t.Run(task)
		}
	})
}
