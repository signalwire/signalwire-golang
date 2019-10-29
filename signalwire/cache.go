package signalwire

import (
	"errors"
	"time"

	bladecache "github.com/dossy/go-cache"
)

const (
	// CacheExpiry Default object expiry
	CacheExpiry = 3600 * 24 /*seconds*/
	// CacheCleaning Cleaning interval for expired cached objects
	CacheCleaning = 10 /*seconds*/
)

// BCache TODO DESCRIPTION
type BCache struct {
	p *bladecache.Cache
}

// InitCache TODO DESCRIPTION
func (cache *BCache) InitCache(expiry, clean time.Duration) error {
	if cache == nil {
		return errors.New("empty cache object")
	}

	cache.p = bladecache.New(expiry, clean)

	return nil
}

// SetCallCache TODO DESCRIPTION
func (cache *BCache) SetCallCache(callID string, sess *CallSession) error {
	if cache == nil {
		return errors.New("empty cache object")
	}

	if cache.p == nil {
		return errors.New("cache not initialized")
	}

	if sess == nil {
		return errors.New("empty session object")
	}

	cache.p.Set(callID, sess, CacheExpiry*time.Second)

	return nil
}

// GetCallCache TODO DESCRIPTION
func (cache *BCache) GetCallCache(callID string) (*CallSession, error) {
	if cache == nil {
		return nil, errors.New("empty cache object")
	}

	if cache.p == nil {
		return nil, errors.New("cache not initialized")
	}

	if v, found := cache.p.Get(callID); found {
		if _, ok := v.(*CallSession); !ok {
			return nil, errors.New("wrong cache data type")
		}

		return v.(*CallSession), nil
	}

	return nil, nil
}

// DeleteCallCache TODO DESCRIPTION
func (cache *BCache) DeleteCallCache(callID string) error {
	if cache == nil {
		return errors.New("empty cache object")
	}

	if cache.p == nil {
		return errors.New("cache not initialized")
	}

	cache.p.Delete(callID)

	return nil
}

// SetMsgCache TODO DESCRIPTION
func (cache *BCache) SetMsgCache(msgID string, sess *MsgSession) error {
	if cache == nil {
		Log.Error("empty cache object")
		return errors.New("empty cache object")
	}

	if cache.p == nil {
		Log.Error("cache not initialized")
		return errors.New("cache not initialized")
	}

	if sess == nil {
		Log.Error("empty session object")
		return errors.New("empty session object")
	}

	cache.p.Set(msgID, sess, CacheExpiry*time.Second)

	return nil
}

// GetMsgCache TODO DESCRIPTION
func (cache *BCache) GetMsgCache(msgID string) (*MsgSession, error) {
	if cache == nil {
		return nil, errors.New("empty cache object")
	}

	if cache.p == nil {
		return nil, errors.New("cache not initialized")
	}

	if v, found := cache.p.Get(msgID); found {
		if _, ok := v.(*MsgSession); !ok {
			return nil, errors.New("wrong cache data type")
		}

		return v.(*MsgSession), nil
	}

	return nil, nil
}

// DeleteMsgCache TODO DESCRIPTION
func (cache *BCache) DeleteMsgCache(msgID string) error {
	if cache == nil {
		return errors.New("empty cache object")
	}

	if cache.p == nil {
		return errors.New("cache not initialized")
	}

	cache.p.Delete(msgID)

	return nil
}
