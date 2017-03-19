package goodies

import (
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

type TypeMismatchError struct {
	err string
}

func (e TypeMismatchError) Error() string {
	return e.err
}

type NotFoundError struct {
	key string
}

func (e NotFoundError) Error() string {
	return fmt.Sprintf("Item for key: %v not found", e.key)
}

type DictKeyNotFound struct {
	key string
}

func (e DictKeyNotFound) Error() string {
	return fmt.Sprintf("Item for key: %v was not found in a dictionary", e.key)
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
		_ = persister.Load(&initialStorage) //we don't really care about whether there was some data or not
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

func (g Goodies) newItem(value interface{}, ttl time.Duration) gItem {
	return gItem{
		Value:  value,
		Expiry: getExpiry(ttl, g.defaultExpiry),
	}
}

func newItemWithExpiry(value interface{}, expiry int64) gItem {
	return gItem{
		Value:  value,
		Expiry: expiry,
	}
}

// Set Method
func (g *Goodies) Set(key string, value string, ttl time.Duration) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.internalSet(key, value, ttl)
}

func (g *Goodies) internalSet(key string, value string, ttl time.Duration) {
	g.storage[key] = g.newItem(value, ttl)
}

// Get Method
func (g *Goodies) Get(key string) (string, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.internalGetString(key)
}

// Update method
// Returns NotFoundError if an item doesn't exist
// Returns TypeMismatchError if an item is not of a string type
func (g *Goodies) Update(key string, value string, ttl time.Duration) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	_, err := g.internalGetString(key)
	if err != nil {
		return err
	}
	g.internalSet(key, value, ttl)
	return nil
}

// Remove key from storage (removes item of any type)
func (g *Goodies) Remove(key string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.internalRemove(key)
}

func (g *Goodies) internalRemove(key string) {
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
func (g *Goodies) ListPush(key string, value string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	list, err := g.internalGetList(key)
	if err != nil {
		switch err.(type) {
		case NotFoundError:
			g.storage[key] = g.newItem(createList(value), g.defaultExpiry)
			return nil
		default:
			return err
		}
	}

	list = append(list, value)
	g.storage[key] = newItemWithExpiry(list, g.storage[key].Expiry)
	return nil
}

// ListLen Returns the length of list. Returns 0 if list not found
// Returns error if value stored is not a list
func (g *Goodies) ListLen(key string) (int, error) {
	list, err := g.internalGetList(key)
	if err != nil {
		switch err.(type) {
		case NotFoundError:
			return 0, nil //No error if the list is just not found
		default:
			return 0, err
		}
	}

	return len(list), nil
}

// ListRemoveIndex Removes list entry
// Returns no error if item was removed or not found
// Returns error if key is not pointing to a list
func (g *Goodies) ListRemoveIndex(key string, index int) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	list, err := g.internalGetList(key)
	if err != nil {
		return err
	}
	if list == nil {
		return nil
	}
	if len(list) <= index {
		return nil
	}
	g.storage[key] = newItemWithExpiry(
		append(list[:index], list[index+1:]...),
		g.storage[key].Expiry)
	return nil
}

//ListRemoveValue Removes all value occurences from the list
//Return error only if the referenced item is not a list
func (g *Goodies) ListRemoveValue(key string, value string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	list, err := g.internalGetList(key)
	if err != nil {
		return err
	}
	if list == nil {
		return nil
	}

	var result []string
	for _, val := range list {
		if val == value {
			continue
		}
		result = append(result, val)
	}

	g.storage[key] = newItemWithExpiry(result, g.storage[key].Expiry)
	return nil
}

// DictSet Sets a value for a specific dictionary key in storage
// Returns an error if referenced item is not a dictionary
func (g *Goodies) DictSet(key string, dictKey string, value string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	dict, err := g.internalGetDict(key)
	if err != nil {
		return err
	}
	if dict == nil {
		dict := make(map[string]interface{}, 1)
		dict[dictKey] = value
		g.storage[key] = g.newItem(dict, g.defaultExpiry)
		return nil
	}
	dict[key] = value
	g.storage[key] = newItemWithExpiry(dict, g.storage[key].Expiry)
	return nil
}

// DictGet returns a value for a dictionary by a key
// Returns an error if referenced item is not a dictionary
func (g *Goodies) DictGet(key string, dictKey string) (string, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	dict, err := g.internalGetDict(key)
	if err != nil {
		return "", err
	}
	val, ok := dict[dictKey]
	if !ok {
		return "", &DictKeyNotFound{dictKey}
	}
	return val, nil
}

// DictRemove Remove a specific key from a dictionary
// Returns an error if referenced item is not a dictionary
func (g *Goodies) DictRemove(key string, dictKey string) error {
	return nil
}

// DictHasKey Can be used to retreive key existence in a dictionary
func (g *Goodies) DictHasKey(key string, dictKey string) (bool, error) {
	return false, nil
}

// SetExpiry Updates item expiry to the specified ttl value
// Returns error in case if item was not found
func (g *Goodies) SetExpiry(key string, ttl time.Duration) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	value, ok := g.internalGet(key)

	if !ok {
		return &TypeMismatchError{fmt.Sprintf("Item %v doesn't exist", key)}
	}
	g.storage[key] = g.newItem(value, ttl)
	return nil
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

func (g *Goodies) internalGetString(key string) (string, error) {
	val, found := g.internalGet(key)
	if !found {
		return "", NotFoundError{key}
	}
	isString := checkValueIsString(key)
	if !isString {
		return "", TypeMismatchError{fmt.Sprintf("Requested item is not a string")}
	}
	return val.(string), nil
}

func (g *Goodies) internalGet(key string) (interface{}, bool) {
	val, found := g.storage[key]
	if !found {
		return nil, false
	}
	if expired := checkExpiry(val.Expiry); expired {
		g.internalRemove(key)
		return nil, false
	}
	return val.Value, found
}

func (g *Goodies) internalGetList(key string) ([]string, error) {
	value, found := g.internalGet(key)
	if !found {
		return nil, NotFoundError{key}
	}
	isList := checkValueIsList(value)
	if !isList {
		return nil, TypeMismatchError{fmt.Sprintf("Item %v is not a list", key)}
	}
	return value.([]string), nil
}

func (g *Goodies) internalGetDict(key string) (map[string]string, error) {
	value, found := g.internalGet(key)
	if !found {
		return nil, NotFoundError{key}
	}
	isDict := checkValueIsDict(value)
	if !isDict {
		return nil, TypeMismatchError{fmt.Sprintf("Item %v is not a dictionary", key)}
	}
	return value.(map[string]string), nil
}

// TODO: make cleanup strategy to run every 2 x defaultExpiration or each 10k items
func (g *Goodies) cleanupOutdated() {
	g.lock.Lock()
	defer g.lock.Unlock()
	for key, value := range g.storage {
		if checkExpiry(value.Expiry) {
			g.internalRemove(key)
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

func checkValueIsString(value interface{}) bool {
	switch value.(type) {
	case string:
		return true
	}
	return false
}

func checkValueIsList(value interface{}) bool {
	switch value.(type) {
	case []string:
		return true
	}
	return false
}

func checkValueIsDict(value interface{}) bool {
	switch value.(type) {
	case map[string]string:
		return true
	}
	return false
}

func createList(value string) []string {
	newList := make([]string, 1)
	newList[0] = value
	return newList
}
