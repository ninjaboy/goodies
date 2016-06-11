package goodies

import "sync"

import "time"

// Goodies bag
type Goodies struct {
	storage       map[string]GItem
	lock          sync.RWMutex
	defaultExpiry time.Duration
}

// GItem is internal Goodies item
type GItem struct {
	Value  interface{}
	Expiry int64
}

const (
	//ExpireNever Use this value when adding value to cache to make element last forever
	ExpireNever time.Duration = -1
	//ExpireDefault Use this value to use default cache expiration
	ExpireDefault time.Duration = -2
)

// NewGoodies creates new isntance of goodiebag
func NewGoodies(ttl time.Duration) *Goodies {
	return &Goodies{
		storage:       make(map[string]GItem),
		defaultExpiry: ttl,
	}
}

// Set Method
func (g *Goodies) Set(key string, value interface{}, ttl time.Duration) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.storage[key] = GItem{
		Value:  value,
		Expiry: getExpiry(ttl, g.defaultExpiry),
	}
}

// Get Method
func (g *Goodies) Get(key string) (interface{}, bool) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	val, found := g.storage[key]
	if !found {
		return nil, false
	}

	if expired := checkExpiry(val.Expiry); expired {
		return nil, false
	}
	return val.Value, found
}

// Update method (at the moment not clear how it should be different to Set)
func (g *Goodies) Update(key string, value interface{}, ttl time.Duration) {
	g.Set(key, value, ttl)
}

// Remove key from storage
func (g *Goodies) Remove(key string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	delete(g.storage, key)
}

// Keys returns list of keys
func (g *Goodies) Keys() []string {
	g.lock.RLock()
	defer g.lock.RUnlock()
	keys := make([]string, len(g.storage))
	i := 0
	for k := range g.storage {
		keys[i] = k
		i++
	}
	return keys
}

func getExpiry(ttl time.Duration, def time.Duration) int64 {
	var expiry int64
	if ttl == ExpireDefault {
		ttl = def
	}
	if ttl > 0 {
		expiry = time.Now().Add(ttl).Unix()
	}
	return expiry
}

func checkExpiry(expiry int64) bool {
	if expiry <= 0 {
		//never expires
		return false
	}
	if time.Now().Unix() > expiry {
		return true
	}
	return false
}
