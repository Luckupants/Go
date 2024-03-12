//go:build !solution

package lrucache

import "container/list"

type Node struct {
	key   int
	value int
}

type LRU struct {
	values *list.List
	recent map[int]*list.Element
	cap    int
}

func (l LRU) Get(key int) (int, bool) {
	node, ok := l.recent[key]
	if !ok {
		return 0, false
	}
	l.values.MoveBefore(node, l.values.Front())
	if l.values.Len() > l.cap {
		delete(l.recent, l.values.Back().Value.(Node).key)
		l.values.Remove(l.values.Back())
	}
	return (l.values.Front().Value).(Node).value, true
}

func (l LRU) Set(key, value int) {
	node, ok := l.recent[key]
	if ok {
		l.values.Remove(node)
	}
	l.values.PushFront(Node{key, value})
	l.recent[key] = l.values.Front()
	if l.values.Len() > l.cap {
		delete(l.recent, l.values.Back().Value.(Node).key)
		l.values.Remove(l.values.Back())
	}
}

func (l LRU) Range(f func(key, value int) bool) {
	for cur := l.values.Back(); cur != nil; cur = cur.Prev() {
		if !f(cur.Value.(Node).key, cur.Value.(Node).value) {
			return
		}
	}
}

func (l LRU) Clear() {
	clear(l.recent)
	for l.values.Front() != nil {
		l.values.Remove(l.values.Front())
	}
}

func New(cap int) Cache {
	return LRU{list.New(), make(map[int]*list.Element), cap}
}
