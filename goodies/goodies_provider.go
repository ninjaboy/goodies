package goodies

import (
	"time"
)

const (
	//ExpireNever Use this value when adding value to cache to make element last forever
	ExpireNever time.Duration = -1
	//ExpireDefault Use this value to use default cache expiration
	ExpireDefault time.Duration = -2
)

// CommandProcessor generic interface implementing transport prototocol for a client
type CommandProcessor interface {
	Process(CommandRequest, *CommandResponse) error
}

// Provider generic client interface combining all available methods
// ttl can be passed as usual time.Duration or as predefined constants(ExpireNever/ExpireDefault)
type Provider interface {
	Set(key string, value string, ttl time.Duration) error
	Get(key string) (string, error)
	Update(key string, value string, ttl time.Duration) error
	Remove(key string) error
	Keys() ([]string, error)

	ListPush(key string, value string) error
	ListLen(key string) (int, error)
	ListRemoveIndex(key string, index int) error
	ListRemoveValue(key string, value string) error
	ListGetByIndex(key string, index int) (string, error)
	DictSet(key string, dictKey string, value string) error
	DictGet(key string, dictKey string) (string, error)
	DictRemove(key string, dictKey string) error
	DictHasKey(key string, dictKey string) (bool, error)
	SetExpiry(key string, ttl time.Duration) error
}
