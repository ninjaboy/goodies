package goodies

import (
	"errors"
	"fmt"
	"sync"
)

import "time"

// Goodies bag
type Goodies struct {
	storage       map[string]gItem
	lock          sync.RWMutex
	defaultExpiry time.Duration
	persister     *Persister
	stop          chan bool
}

// gItem is internal Goodies item
type gItem struct {
	Value  interface{}
	Expiry int64
}

const (
	//ExpireNever Use this value when adding value to cache to make element last forever
	ExpireNever time.Duration = -1
	//ExpireDefault Use this value to use default cache expiration
	ExpireDefault time.Duration = -2
)

// NewGoodies creates new instance of goodiebag
func NewGoodies(ttl time.Duration, filename string, persistInterval time.Duration) *Goodies {
	var persister *Persister
	initialStorage := make(map[string]gItem)
	if filename != "" {
		persister = NewPersister(filename, persistInterval)
		err := persister.Load(&initialStorage)
		if err != nil {
			fmt.Println(err)
		}
	}

	goodies := &Goodies{
		storage:       initialStorage,
		defaultExpiry: ttl,
		stop:          make(chan bool),
	}

	if filename != "" {
		go goodies.runPersister(persister)
	}

	return goodies
}

// Set Method
func (g *Goodies) Set(key string, value interface{}, ttl time.Duration) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.storage[key] = gItem{
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

// Update method
func (g *Goodies) Update(key string, value interface{}, ttl time.Duration) (interface{}, error) {
	if _, found := g.Get(key); !found {
		return nil, errors.New("Key " + key + " doesn't exist")
	}
	g.Set(key, value, ttl)
	return value, nil
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

func (g *Goodies) ListAdd(key string, value interface{}, ttl time.Duration) {

}

func (g *Goodies) ListRemove(key string, index int) error {
	return nil
}

func (g *Goodies) DictSet(key string, value interface{}) {

}

//Stop method is a nice way to clearly stop the cache
func (g *Goodies) Stop() {
	g.stop <- true
}

func (g *Goodies) runPersister(p *Persister) {
	persistTrigger := time.NewTicker(p.interval)
	for {
		select {
		case <-persistTrigger.C:
			//fmt.Println("Saving blob")
			g.cleanupOutdated()
			if err := p.Save(g.getBlob()); err != nil {
				fmt.Printf("Backup not saved %v\n", err)
			}
		case <-g.stop:
			g.cleanupOutdated()
			//fmt.Println("Saving blob")
			if err := p.Save(g.getBlob()); err != nil {
				fmt.Printf("Backup not saved %v\n", err)
			}
			return
		}
	}
}

func (g *Goodies) cleanupOutdated() {
	g.lock.Lock()
	defer g.lock.Unlock()
	for key, value := range g.storage {
		if checkExpiry(value.Expiry) {
			delete(g.storage, key)
		}
	}
}

func (g *Goodies) getBlob() map[string]gItem {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.storage
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
