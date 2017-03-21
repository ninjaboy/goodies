package goodies

import (
	"strconv"
	"strings"
	"time"
)

func internalProcess(req GoodiesRequest, c Client) GoodiesResponse {
	var res GoodiesResponse
	err := c.transport.Process(req, &res)
	if err != nil {
		return GoodiesResponse{false, "", ErrInternalError{err.Error()}}
	}
	return res
}

func (c Client) Set(key string, value string, ttl time.Duration) error {
	req := GoodiesRequest{"Set", []string{key, value, ttlAsString(ttl)}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c Client) Get(key string) (string, error) {
	req := GoodiesRequest{"Get", []string{key}}
	res := internalProcess(req, c)
	if !res.Success {
		return "", res.Err
	}
	return res.Result, nil
}

func (c Client) Update(key string, value string, ttl time.Duration) error {
	req := GoodiesRequest{"Update", []string{key, value, ttlAsString(ttl)}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c Client) Remove(key string) error {
	req := GoodiesRequest{"Remove", []string{key}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c Client) Keys() ([]string, error) {
	req := GoodiesRequest{"Keys", []string{}}
	res := internalProcess(req, c)
	if !res.Success {
		return nil, res.Err
	}
	return strings.Split(res.Result, ":"), nil
}

func (c Client) ListPush(key string, value string) error {
	req := GoodiesRequest{"ListPush", []string{key, value}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c Client) ListLen(key string) (int, error) {
	req := GoodiesRequest{"ListLen", []string{key}}
	res := internalProcess(req, c)
	if !res.Success {
		return 0, res.Err
	}
	len, _ := strconv.Atoi(res.Result)
	return len, nil
}

func (c Client) ListRemoveIndex(key string, index int) error {
	req := GoodiesRequest{"ListRemoveIndex", []string{key, string(index)}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c Client) ListRemoveValue(key string, value string) error {
	req := GoodiesRequest{"ListRemoveIndex", []string{key, value}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c Client) ListGetByIndex(key string, index int) (string, error) {
	req := GoodiesRequest{"ListGetByIndex", []string{key, string(index)}}
	res := internalProcess(req, c)
	if !res.Success {
		return "", res.Err
	}
	return res.Result, nil
}

func (c Client) DictSet(key string, dictKey string, value string) error {
	req := GoodiesRequest{"DictSet", []string{key, dictKey, value}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c Client) DictGet(key string, dictKey string) (string, error) {
	req := GoodiesRequest{"DictGet", []string{key, dictKey}}
	res := internalProcess(req, c)
	if !res.Success {
		return "", res.Err
	}
	return res.Result, nil
}

func (c Client) DictRemove(key string, dictKey string) error {
	req := GoodiesRequest{"DictRemove", []string{key, dictKey}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c Client) DictHasKey(key string, dictKey string) (bool, error) {
	req := GoodiesRequest{"DictHasKey", []string{key, dictKey}}
	res := internalProcess(req, c)
	if !res.Success {
		return false, res.Err
	}
	if res.Result == "1" {
		return true, nil
	}
	return false, nil
}

func (c Client) SetExpiry(key string, ttl time.Duration) error {
	req := GoodiesRequest{"SetExpiry", []string{key, ttlAsString(ttl)}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func ttlAsString(ttl time.Duration) string {
	return string(ttl.Nanoseconds())
}
