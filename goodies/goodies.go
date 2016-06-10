package goodies

import "sync"

import "time"

// Goodies bag
type Goodies struct {
	Storage       map[string]GItem
	Lock          sync.RWMutex
	DefaultExpiry time.Duration
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
		Storage:       make(map[string]GItem),
		DefaultExpiry: ttl,
	}
}

// Set Method
func (g *Goodies) Set(key string, value interface{}, ttl time.Duration) {
	g.Lock.Lock()
	defer g.Lock.Unlock()
	g.Storage[key] = GItem{
		Value:  value,
		Expiry: getExpiry(ttl, g.DefaultExpiry),
	}
}

// Get Method
func (g *Goodies) Get(key string) (interface{}, bool) {
	g.Lock.RLock()
	defer g.Lock.RUnlock()

	val, found := g.Storage[key]
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
	g.Lock.Lock()
	defer g.Lock.Unlock()
	delete(g.Storage, key)
}

// Keys returns list of keys
func (g *Goodies) Keys() []string {
	g.Lock.RLock()
	defer g.Lock.RUnlock()
	keys := make([]string, len(g.Storage))
	i := 0
	for k := range g.Storage {
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
		expiry = time.Now().Add(ttl).UnixNano()
	}
	return expiry
}

func checkExpiry(expiry int64) bool {
	if expiry <= 0 {
		//never expires
		return false
	}
	if time.Now().UnixNano() > expiry {
		return true
	}
	return false
}
