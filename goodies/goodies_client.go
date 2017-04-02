package goodies

import (
	"strconv"
	"strings"
	"time"
)

type goodiesClient struct {
	transport CommandProcessor
}

func internalProcess(req CommandRequest, c goodiesClient) CommandResponse {
	var res CommandResponse
	err := c.transport.Process(req, &res)
	if err != nil {
		return CommandResponse{false, "", ErrInternalError{err.Error()}}
	}
	return res
}

func (c goodiesClient) Set(key string, value string, ttl time.Duration) error {
	req := CommandRequest{"Set", []string{key, value, ttlAsString(ttl)}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c goodiesClient) Get(key string) (string, error) {
	req := CommandRequest{"Get", []string{key}}
	res := internalProcess(req, c)
	if !res.Success {
		return "", res.Err
	}
	return res.Result, nil
}

func (c goodiesClient) Update(key string, value string, ttl time.Duration) error {
	req := CommandRequest{"Update", []string{key, value, ttlAsString(ttl)}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c goodiesClient) Remove(key string) error {
	req := CommandRequest{"Remove", []string{key}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c goodiesClient) Keys() ([]string, error) {
	req := CommandRequest{"Keys", []string{}}
	res := internalProcess(req, c)
	if !res.Success {
		return nil, res.Err
	}
	return strings.Split(res.Result, ":"), nil
}

func (c goodiesClient) ListPush(key string, value string) error {
	req := CommandRequest{"ListPush", []string{key, value}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c goodiesClient) ListLen(key string) (int, error) {
	req := CommandRequest{"ListLen", []string{key}}
	res := internalProcess(req, c)
	if !res.Success {
		return 0, res.Err
	}
	len, _ := strconv.Atoi(res.Result)
	return len, nil
}

func (c goodiesClient) ListRemoveIndex(key string, index int) error {
	req := CommandRequest{"ListRemoveIndex", []string{key, strconv.Itoa(index)}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c goodiesClient) ListRemoveValue(key string, value string) error {
	req := CommandRequest{"ListRemoveIndex", []string{key, value}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c goodiesClient) ListGetByIndex(key string, index int) (string, error) {
	req := CommandRequest{"ListGetByIndex", []string{key, strconv.Itoa(index)}}
	res := internalProcess(req, c)
	if !res.Success {
		return "", res.Err
	}
	return res.Result, nil
}

func (c goodiesClient) DictSet(key string, dictKey string, value string) error {
	req := CommandRequest{"DictSet", []string{key, dictKey, value}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c goodiesClient) DictGet(key string, dictKey string) (string, error) {
	req := CommandRequest{"DictGet", []string{key, dictKey}}
	res := internalProcess(req, c)
	if !res.Success {
		return "", res.Err
	}
	return res.Result, nil
}

func (c goodiesClient) DictRemove(key string, dictKey string) error {
	req := CommandRequest{"DictRemove", []string{key, dictKey}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func (c goodiesClient) DictHasKey(key string, dictKey string) (bool, error) {
	req := CommandRequest{"DictHasKey", []string{key, dictKey}}
	res := internalProcess(req, c)
	if !res.Success {
		return false, res.Err
	}
	if res.Result == "1" {
		return true, nil
	}
	return false, nil
}

func (c goodiesClient) SetExpiry(key string, ttl time.Duration) error {
	req := CommandRequest{"SetExpiry", []string{key, ttlAsString(ttl)}}
	res := internalProcess(req, c)
	if !res.Success {
		return res.Err
	}
	return nil
}

func ttlAsString(ttl time.Duration) string {
	if ttl == ExpireDefault {
		return "-2"
	}
	if ttl == ExpireNever {
		return "-1"
	}
	return strconv.FormatInt(int64(ttl.Seconds()), 10)
}
