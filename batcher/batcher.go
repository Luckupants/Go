//go:build !solution

package batcher

import (
	"context"
	"gitlab.com/slon/shad-go/batcher/slow"
	"sync"
	"time"
)

type Batcher struct {
	entry         chan struct{}
	mx            sync.Mutex
	loadersAmount int
	timeout       context.Context
	cancel        context.CancelFunc
	enoughLoaders chan struct{}
	result        interface{}
	object        *slow.Value
	calculated    chan struct{}
}

const (
	queueSize = 100
	timeOut   = time.Millisecond * 100
)

func NewBatcher(v *slow.Value) *Batcher {
	answer := &Batcher{entry: make(chan struct{}), object: v}
	close(answer.entry)
	return answer
}

func (b *Batcher) Load() interface{} {
	<-b.entry
	b.mx.Lock()
	oldAmount := b.loadersAmount
	b.loadersAmount++
	switch b.loadersAmount {
	case 1:
		b.timeout, b.cancel = context.WithTimeout(context.Background(), timeOut)
		b.enoughLoaders = make(chan struct{})
		b.calculated = make(chan struct{})
	case queueSize:
		close(b.enoughLoaders)
	}
	b.mx.Unlock()
	select {
	case <-b.enoughLoaders:
	case <-b.timeout.Done():
	}
	if oldAmount == 0 {
		b.entry = make(chan struct{})
		res := b.object.Load()
		b.result = res
		close(b.calculated)
		b.mx.Lock()
		b.loadersAmount = 0
		b.mx.Unlock()
		close(b.entry)
		return res
	}
	<-b.calculated
	return b.result
}
