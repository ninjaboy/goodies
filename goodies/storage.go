package goodies

import (
	"fmt"
	"sync"
)

import "time"

// Storage bag
type Storage struct {
	storage       map[string]goodiesItem
	lock          sync.RWMutex
	defaultExpiry time.Duration
}

// goodiesItem is internal Goodies item
type goodiesItem struct {
	Value  interface{}
	Expiry int64
}

// NewGoodiesStorage creates new instance of goodiebag
func NewGoodiesStorage(ttl time.Duration) *Storage {
	initialStorage := make(map[string]goodiesItem)
	goodies := &Storage{
		storage:       initialStorage,
		defaultExpiry: ttl,
	}
	return goodies
}

func (g *Storage) newItem(value interface{}, ttl time.Duration) goodiesItem {
	return goodiesItem{
		Value:  value,
		Expiry: getExpiry(ttl, g.defaultExpiry),
	}
}

func newItemWithExpiry(value interface{}, expiry int64) goodiesItem {
	return goodiesItem{
		Value:  value,
		Expiry: expiry,
	}
}

// Set Method
func (g *Storage) Set(key string, value string, ttl time.Duration) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	//TODO: disallow key to contain ',' for keys serialisation simplification
	g.internalSet(key, value, ttl)
	return nil
}

func (g *Storage) internalSet(key string, value string, ttl time.Duration) {
	g.storage[key] = g.newItem(value, ttl)
}

// Get Method
func (g *Storage) Get(key string) (string, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	return g.internalGetString(key)
}

// Update method
// Returns ErrNotFound if an item doesn't exist
// Returns ErrTypeMismatch if an item is not of a string type
func (g *Storage) Update(key string, value string, ttl time.Duration) error {
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
func (g *Storage) Remove(key string) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.internalRemove(key)
	return nil
}

func (g *Storage) internalRemove(key string) {
	delete(g.storage, key)
}

// Keys returns list of keys
func (g *Storage) Keys() ([]string, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()
	keys := make([]string, len(g.storage))
	i := 0
	for k := range g.storage {
		keys[i] = k
		i++
	}
	return keys, nil
}

// ListPush Adds a value into the end of list. Creates a list if it doesn't exist
// An error will be returned in case
func (g *Storage) ListPush(key string, value string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	list, err := g.internalGetList(key)
	if err != nil {
		switch err.(type) {
		case ErrNotFound:
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
func (g *Storage) ListLen(key string) (int, error) {
	list, err := g.internalGetList(key)
	if err != nil {
		switch err.(type) {
		case ErrNotFound:
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
func (g *Storage) ListRemoveIndex(key string, index int) error {
	g.lock.Lock()
	defer g.lock.Unlock()
	list, err := g.internalGetList(key)
	if err != nil {
		switch err.(type) {
		case ErrNotFound:
			return nil //No error if the list is just not found
		default:
			return err
		}
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
func (g *Storage) ListRemoveValue(key string, value string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	list, err := g.internalGetList(key)
	if err != nil {
		switch err.(type) {
		case ErrNotFound:
			return nil //No error if the list is just not found
		default:
			return err
		}
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

// ListGetByIndex Returns an item from a referenced list by index
// Returns ErrNotFound in case if list was not found, ErrTypeMismatch in case if referenced item is not a list
func (g *Storage) ListGetByIndex(key string, index int) (string, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	list, err := g.internalGetList(key)
	if err != nil {
		return "", err
	}

	if len(list) <= index {
		return "", nil
	}
	return list[index], nil
}

// DictSet Sets a value for a specific dictionary key in storage
// Returns an error if referenced item is not a dictionary
func (g *Storage) DictSet(key string, dictKey string, value string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	dict, err := g.internalGetDict(key)
	if err != nil {
		switch err.(type) {
		case ErrNotFound:
			dict := make(map[string]string, 1)
			dict[dictKey] = value
			g.storage[key] = g.newItem(dict, g.defaultExpiry)
			return nil
		default:
			return err
		}
	}

	dict[key] = value
	g.storage[key] = newItemWithExpiry(dict, g.storage[key].Expiry)
	return nil
}

// DictGet returns a value for a dictionary by a key
// Returns an error if referenced item is not a dictionary
func (g *Storage) DictGet(key string, dictKey string) (string, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	dict, err := g.internalGetDict(key)
	if err != nil {
		return "", err
	}
	val, ok := dict[dictKey]
	if !ok {
		return "", ErrDictKeyNotFound{dictKey}
	}
	return val, nil
}

// DictRemove Remove a specific key from a dictionary
// Returns an error if referenced item is not a dictionary
func (g *Storage) DictRemove(key string, dictKey string) error {
	g.lock.RLock()
	defer g.lock.RUnlock()

	dict, err := g.internalGetDict(key)
	if err != nil {
		return err
	}
	delete(dict, dictKey)
	return nil
}

// DictHasKey Can be used to retreive key existence in a dictionary
func (g *Storage) DictHasKey(key string, dictKey string) (bool, error) {
	g.lock.RLock()
	defer g.lock.RUnlock()

	dict, err := g.internalGetDict(key)
	if err != nil {
		return false, err
	}
	_, ok := dict[dictKey]
	return ok, nil
}

// SetExpiry Updates item expiry to the specified ttl value
// Returns error in case if item was not found
func (g *Storage) SetExpiry(key string, ttl time.Duration) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	value, ok := g.internalGet(key)

	if !ok {
		return &ErrTypeMismatch{fmt.Sprintf("Item %v doesn't exist", key)}
	}
	g.storage[key] = g.newItem(value, ttl)
	return nil
}

func (g *Storage) internalGetString(key string) (string, error) {
	val, found := g.internalGet(key)
	if !found {
		return "", ErrNotFound{key}
	}
	isString := checkValueIsString(key)
	if !isString {
		return "", ErrTypeMismatch{fmt.Sprintf("Requested item is not a string")}
	}
	return val.(string), nil
}

func (g *Storage) internalGet(key string) (interface{}, bool) {
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

func (g *Storage) internalGetList(key string) ([]string, error) {
	value, found := g.internalGet(key)
	if !found {
		return nil, ErrNotFound{key}
	}
	isList := checkValueIsList(value)
	if !isList {
		return nil, ErrTypeMismatch{fmt.Sprintf("Item %v is not a list", key)}
	}
	return value.([]string), nil
}

func (g *Storage) internalGetDict(key string) (map[string]string, error) {
	value, found := g.internalGet(key)
	if !found {
		return nil, ErrNotFound{key}
	}
	isDict := checkValueIsDict(value)
	if !isDict {
		return nil, ErrTypeMismatch{fmt.Sprintf("Item %v is not a dictionary", key)}
	}
	return value.(map[string]string), nil
}

// TODO: make cleanup strategy to run every 2 x defaultExpiration or each 10k items
func (g *Storage) cleanupOutdated() {
	g.lock.Lock()
	defer g.lock.Unlock()
	for key, value := range g.storage {
		if checkExpiry(value.Expiry) {
			g.internalRemove(key)
		}
	}
}

func (g *Storage) getBlob() *map[string]goodiesItem {
	g.lock.Lock()
	defer g.lock.Unlock()
	return &g.storage
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
