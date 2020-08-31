package container

import (
	"encoding/json"
	"sync"
)

var SHARD_COUNT = 16

type ConcurrentTable []*ConcurrentMapShared

type ConcurrentMapShared struct {
	items map[string]interface{}
	sync.RWMutex
}

func NewConcurrentTable() ConcurrentTable {
	m := make(ConcurrentTable, SHARD_COUNT)
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]interface{}, 50)}
	}
	return m
}

func NewConcurrentTableWithSize(size int) ConcurrentTable {
	m := make(ConcurrentTable, SHARD_COUNT)
	size = size / SHARD_COUNT
	for i := 0; i < SHARD_COUNT; i++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]interface{}, size)}
	}
	return m
}

func (m ConcurrentTable) getShard(key string) *ConcurrentMapShared {
	return m[uint(fnv32(key))%uint(SHARD_COUNT)]
}

func (m ConcurrentTable) MSet(data map[string]interface{}) {
	for key, value := range data {
		shard := m.getShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

func (m ConcurrentTable) Set(key string, value interface{}) {
	// Get map shard.
	shard := m.getShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

func (m ConcurrentTable) Incr(key string, step int) {
	// Get map shard.
	shard := m.getShard(key)
	shard.Lock()
	ai, ok := shard.items[key]
	var count int
	if ok {
		count = ai.(int)
	}
	count += step
	shard.items[key] = count
	shard.Unlock()
}

type UpsertCb func(exist bool, valueInMap interface{}, newValue interface{}) interface{}

func (m ConcurrentTable) UpSert(key string, value interface{}, cb UpsertCb) (res interface{}) {
	shard := m.getShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

func (m ConcurrentTable) SetIfAbsent(key string, value interface{}) bool {
	// Get map shard.
	shard := m.getShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

func (m ConcurrentTable) Get(key string) (interface{}, bool) {
	// Get shard
	shard := m.getShard(key)
	shard.RLock()
	// Get item from shard.
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

func (m ConcurrentTable) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

func (m ConcurrentTable) Has(key string) bool {
	// Get shard
	shard := m.getShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

func (m ConcurrentTable) Remove(key string) {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

type RemoveCb func(key string, v interface{}, exists bool) bool

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

// RemoveCb locks the shard containing the key, retrieves its current value and calls the callback with those params
// If callback returns true and element exists, it will remove it from the map
// Returns the value returned by the callback (even if element was not present in the map)
func (m ConcurrentTable) RemoveCb(key string, cb RemoveCb) bool {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

func (m ConcurrentTable) Pop(key string) (v interface{}, exists bool) {
	// Try to get shard.
	shard := m.getShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

func (m ConcurrentTable) IsEmpty() bool {
	return m.Count() == 0
}

type Tuple struct {
	Key string
	Val interface{}
}

// Deprecated: using IterBuffered() will get a better performence
func (m ConcurrentTable) Iter() <-chan Tuple {
	chans := snapshot(m)
	ch := make(chan Tuple)
	go fanIn(chans, ch)
	return ch
}

// IterBuffered returns a buffered iterator which could be used in a for range loop.
func (m ConcurrentTable) IterBuffered() <-chan Tuple {
	chans := snapshot(m)
	total := 0
	for _, c := range chans {
		total += cap(c)
	}
	ch := make(chan Tuple, total)
	go fanIn(chans, ch)
	return ch
}

func snapshot(m ConcurrentTable) (chans []chan Tuple) {
	chans = make([]chan Tuple, SHARD_COUNT)
	wg := sync.WaitGroup{}
	wg.Add(SHARD_COUNT)
	for index, shard := range m {
		go func(index int, shard *ConcurrentMapShared) {
			shard.RLock()
			chans[index] = make(chan Tuple, len(shard.items))
			wg.Done()
			for key, val := range shard.items {
				chans[index] <- Tuple{key, val}
			}
			shard.RUnlock()
			close(chans[index])
		}(index, shard)
	}
	wg.Wait()
	return chans
}

func fanIn(chans []chan Tuple, out chan Tuple) {
	wg := sync.WaitGroup{}
	wg.Add(len(chans))
	for _, ch := range chans {
		go func(ch chan Tuple) {
			for t := range ch {
				out <- t
			}
			wg.Done()
		}(ch)
	}
	wg.Wait()
	close(out)
}

func (m ConcurrentTable) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}

	return tmp
}

type IterCb func(key string, v interface{})

// Callback based iterator, cheapest way to read
// all elements in a map.
func (m ConcurrentTable) IterCb(fn IterCb) {
	for idx := range m {
		shard := (m)[idx]
		shard.RLock()
		for key, value := range shard.items {
			fn(key, value)
		}
		shard.RUnlock()
	}
}

func (m ConcurrentTable) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(SHARD_COUNT)
		for _, shard := range m {
			go func(shard *ConcurrentMapShared) {
				// Foreach key, value pair.
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
				wg.Done()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}

func (m ConcurrentTable) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}
