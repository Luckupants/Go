//go:build !solution

package dupcall

import (
	"context"
	"sync"
	"sync/atomic"
)

type Call struct {
	mx            sync.Mutex
	result        interface{}
	err           error
	waitingAmount atomic.Int32
	cbCtx         context.Context
	cancel        context.CancelFunc
	done          chan struct{}
}

func (o *Call) Do(
	ctx context.Context,
	cb func(context.Context) (interface{}, error),
) (result interface{}, err error) {
	o.mx.Lock()
	var done chan struct{}
	if o.waitingAmount.Add(1) == 1 {
		done = make(chan struct{})
		o.done = done
		o.cbCtx, o.cancel = context.WithCancel(context.Background())
		go func() {
			result, err := cb(o.cbCtx)
			o.mx.Lock()
			o.result, o.err = result, err
			o.mx.Unlock()
			close(done)
		}()
	} else {
		done = o.done
	}
	o.mx.Unlock()
	select {
	case <-done:
		o.mx.Lock()
		result, err = o.result, o.err
		o.waitingAmount.Add(-1)
		o.mx.Unlock()
	case <-ctx.Done():
		o.mx.Lock()
		result, err = o.result, o.err
		if o.waitingAmount.Add(-1) == 0 {
			o.cancel()
		}
		o.mx.Unlock()
		return nil, ctx.Err()
	}
	return result, err
}
