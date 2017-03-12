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

type goodiesError struct {
	err string
}

func (e *goodiesError) Error() string {
	return e.err
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

func (g *Goodies) newItem(value interface{}, ttl time.Duration) gItem {
	return gItem{
		Value:  value,
		Expiry: getExpiry(ttl, g.defaultExpiry),
	}
}
func (g *Goodies) newItemWithExpiry(value interface{}, expiry int64) gItem {
	return gItem{
		Value:  value,
		Expiry: expiry,
	}
}

// Set Method
func (g *Goodies) Set(key string, value interface{}, ttl time.Duration) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.storage[key] = g.newItem(value, ttl)
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
		delete(g.storage, key) //remove outdated value
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

// ListPush Adds a value into the end of list. Creates a list if it doesn't exist
// An error will be returned in case
func (g *Goodies) ListPush(key string, value interface{}, ttl time.Duration) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	found, err := g.checkListExists(key)
	if err != nil {
		return err
	}
	if !found {
		g.storage[key] = g.newItem(createList(value), ttl)
	} else {
		if expired := checkExpiry(g.storage[key].Expiry); expired {
			g.storage[key] = g.newItem(createList(value), ttl)
		} else {
			list := g.storage[key].Value
			list = append(list.([]interface{}), value)
			g.storage[key] = g.newItemWithExpiry(list, g.storage[key].Expiry)
		}
	}
	return nil
}

// ListLen Returns the length of list. Returns 0 if list not found
// Returns error if value stored is not a list
func (g *Goodies) ListLen(key string) (int, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	found, err := g.checkListExists(key)
	if err != nil {
		return -1, err
	}
	if !found {
		return 0, nil
	}
	if expired := checkExpiry(g.storage[key].Expiry); expired {
		delete(g.storage, key)
		return 0, nil
	}
	return len(g.storage[key].Value.([]interface{})), nil
}

// ListRemove Placeholder for remove from list function
func (g *Goodies) ListRemove(key string, index int) error {
	return nil
}

// DictSet Placeholder for add to dictionary function
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
			g.cleanupOutdated()
			if err := p.Save(g.getBlob()); err != nil {
				fmt.Printf("Backup not saved %v\n", err)
			}
		case <-g.stop:
			g.cleanupOutdated()
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

func (g *Goodies) checkListExists(key string) (bool, error) {
	val, ok := g.storage[key]
	if !ok {
		return false, nil
	}
	switch val.Value.(type) {
	case []interface{}:
		return true, nil
	}
	return false, &goodiesError{fmt.Sprintf("Item %v is not a list", key)}
}

func createList(value interface{}) []interface{} {
	newList := make([]interface{}, 1)
	newList[0] = value
	return newList
}
