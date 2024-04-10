//go:build !solution

package keylock

import (
	"sort"
	"sync"
)

type KeyLock struct {
	mutexes map[string]chan struct{}
	mx      sync.Mutex
}

func New() *KeyLock {
	return &KeyLock{mutexes: make(map[string]chan struct{})}
}

func (l *KeyLock) LockKeys(keys []string, cancel <-chan struct{}) (canceled bool, unlock func()) {
	copyKeys := make([]string, len(keys))
	copy(copyKeys, keys)
	sort.Strings(copyKeys)
	keys = copyKeys
	unlock = func() {
		l.mx.Lock()
		for _, key := range keys {
			l.mutexes[key] <- struct{}{}
		}
		l.mx.Unlock()
	}
	for i, key := range keys {
		l.mx.Lock()
		if _, ok := l.mutexes[key]; !ok {
			l.mutexes[key] = make(chan struct{}, 1)
			l.mutexes[key] <- struct{}{}
		}
		mutex := l.mutexes[key]
		l.mx.Unlock()
		select {
		case <-mutex:
		case <-cancel:
			l.mx.Lock()
			for j := 0; j < i; j++ {
				l.mutexes[keys[j]] <- struct{}{}
			}
			l.mx.Unlock()
			return true, func() {}
		}
	}
	return false, unlock
}
