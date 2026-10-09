package cache

import (
	"container/list"
	"time"
)

type entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
}

// shard is a plain LRU with TTL. The caller holds the lock.
type shard[K comparable, V any] struct {
	maxSize int
	order   *list.List
	items   map[K]*list.Element
}

func newShard[K comparable, V any](maxSize int) *shard[K, V] {
	return &shard[K, V]{maxSize: maxSize, order: list.New(), items: make(map[K]*list.Element)}
}

func (s *shard[K, V]) get(key K, now time.Time) (*entry[K, V], bool) {
	el, ok := s.items[key]
	if !ok {
		return nil, false
	}
	e := el.Value.(*entry[K, V])
	if !now.Before(e.expiresAt) {
		s.remove(el)
		return nil, false
	}
	s.order.MoveToFront(el)
	return e, true
}

func (s *shard[K, V]) set(key K, value V, expiresAt time.Time) {
	if el, ok := s.items[key]; ok {
		e := el.Value.(*entry[K, V])
		e.value, e.expiresAt = value, expiresAt
		s.order.MoveToFront(el)
		return
	}
	s.items[key] = s.order.PushFront(&entry[K, V]{key: key, value: value, expiresAt: expiresAt})
	for s.order.Len() > s.maxSize {
		s.remove(s.order.Back())
	}
}

func (s *shard[K, V]) delete(key K) bool {
	el, ok := s.items[key]
	if ok {
		s.remove(el)
	}
	return ok
}

func (s *shard[K, V]) remove(el *list.Element) {
	s.order.Remove(el)
	delete(s.items, el.Value.(*entry[K, V]).key)
}

// sweep checks up to limit entries from the cold end and drops the expired ones.
func (s *shard[K, V]) sweep(now time.Time, limit int) int {
	removed := 0
	el := s.order.Back()
	for scanned := 0; el != nil && scanned < limit; scanned++ {
		prev := el.Prev()
		if !now.Before(el.Value.(*entry[K, V]).expiresAt) {
			s.remove(el)
			removed++
		}
		el = prev
	}
	return removed
}
