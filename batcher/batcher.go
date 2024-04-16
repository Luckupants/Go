//go:build !solution

package batcher

import (
	"context"
	"gitlab.com/slon/shad-go/batcher/slow"
	"sync"
	"time"
)

type BatchInfo struct {
	result       interface{}
	isCalculated chan struct{}
	timeout      context.Context
}

type Batcher struct {
	mx            sync.Mutex
	entry         sync.Cond
	entryAllowed  bool
	loadersAmount int
	enoughLoaders chan struct{}
	object        *slow.Value
	curBatchInfo  *BatchInfo
}

const (
	queueSize = 100
	timeOut   = time.Millisecond * 100
)

func NewBatcher(v *slow.Value) *Batcher {
	answer := &Batcher{entryAllowed: true, object: v}
	answer.entry = *sync.NewCond(&answer.mx)
	return answer
}

func (b *Batcher) Load() interface{} {
	b.mx.Lock()
	for !b.entryAllowed {
		b.entry.Wait()
	}
	oldAmount := b.loadersAmount
	b.loadersAmount++
	switch b.loadersAmount {
	case 1:
		b.curBatchInfo = &BatchInfo{isCalculated: make(chan struct{})}
		var cancel context.CancelFunc
		b.curBatchInfo.timeout, cancel = context.WithTimeout(context.Background(), timeOut)
		defer cancel()
		b.enoughLoaders = make(chan struct{})
	case queueSize:
		close(b.enoughLoaders)
	}
	curInfo := b.curBatchInfo
	b.mx.Unlock()
	if oldAmount == 0 {
		select {
		case <-b.enoughLoaders:
		case <-curInfo.timeout.Done():
		}
		b.mx.Lock()
		b.entryAllowed = false
		b.loadersAmount = 0
		res := b.object.Load()
		curInfo.result = res
		close(curInfo.isCalculated)
		b.entryAllowed = true
		b.entry.Broadcast()
		b.mx.Unlock()
		return res
	}
	<-curInfo.isCalculated
	return curInfo.result
}
