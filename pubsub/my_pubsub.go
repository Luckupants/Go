//go:build !solution

package pubsub

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var _ Subscription = (*MySubscription)(nil)

type MySubscription struct {
	mx      *sync.Mutex
	list    *list.List
	element *list.Element
}

func (s *MySubscription) Unsubscribe() {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.list.Remove(s.element)
}

var _ PubSub = (*MyPubSub)(nil)

type MyPubSub struct {
	subscriptions map[string]*list.List
	mx            sync.Mutex
	ctx           context.Context
	isClosed      atomic.Bool
}

func NewPubSub() PubSub {
	return &MyPubSub{subscriptions: make(map[string]*list.List)}
}

func (p *MyPubSub) Subscribe(subj string, cb MsgHandler) (Subscription, error) {
	if p.isClosed.Load() {
		return nil, errors.New("Lol ti daun")
	}
	p.mx.Lock()
	defer p.mx.Unlock()
	if p.subscriptions[subj] == nil {
		p.subscriptions[subj] = &list.List{}
	}
	l := p.subscriptions[subj]
	l.PushBack(cb)
	sub := &MySubscription{mx: &p.mx, list: l, element: l.Back()}
	return sub, nil
}

func (p *MyPubSub) Publish(subj string, msg interface{}) error {
	if p.isClosed.Load() {
		return errors.New("Ti loh!!")
	}
	p.mx.Lock()
	defer p.mx.Unlock()
	l := p.subscriptions[subj]
	for cur := l.Front(); cur != nil; cur = cur.Next() {
		cur := cur
		timer, closeTimer := context.WithTimeout(context.Background(), time.Millisecond*10)
		done := make(chan struct{})
		go func() {
			cur.Value.(MsgHandler)(msg)
			close(done)
		}()
		select {
		case <-timer.Done():
		case <-done:
		}
		closeTimer()
		if p.isClosed.Load() {
			select {
			case <-p.ctx.Done():
				return nil
			default:
			}
		}
	}
	return nil
}

func (p *MyPubSub) Close(ctx context.Context) error {
	p.ctx = ctx
	p.isClosed.Store(true)
	return nil
}
