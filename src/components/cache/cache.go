package cache

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/fx"
)

type CacheEntry struct {
	Expire time.Time `json:"e"`
	Value  string    `json:"v"`
}

type CacheResult struct {
	fx.Out

	Cache *Cache
}

type Cache struct {
	Root string
	Log  *slog.Logger
}

type CacheParams struct {
	fx.In

	Log *slog.Logger
}

func NewCache(p CacheParams) CacheResult {
	c := &Cache{
		Root: filepath.Join(os.TempDir(), "gspot.cache"),
		Log:  p.Log,
	}
	return CacheResult{
		Cache: c,
	}
}

func (c *Cache) load() (map[string]CacheEntry, error) {
	out := map[string]CacheEntry{}
	cache, err := os.Open(c.Root)
	if err != nil {
		return nil, err
	}
	if err := json.NewDecoder(cache).Decode(&out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Cache) save(m map[string]CacheEntry) error {
	payload, err := json.Marshal(m)
	if err != nil {
		return err
	}
	slog.Debug("CACHE", "saving", string(payload))
	err = os.WriteFile(c.Root, payload, 0o600)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cache) GetOrDo(key string, do func() (string, error), ttl time.Duration) (string, error) {
	conf, err := c.load()
	if err != nil {
		slog.Debug("CACHE", "failed read", err)
		return c.Do(key, do, ttl)
	}
	val, ok := conf[key]
	if !ok {
		return c.Do(key, do, ttl)
	}
	if time.Now().After(val.Expire) {
		return c.Do(key, do, ttl)
	}
	return val.Value, nil
}

func (c *Cache) Do(key string, do func() (string, error), ttl time.Duration) (string, error) {
	if do == nil {
		return "", nil
	}
	res, err := do()
	if err != nil {
		return "", err
	}
	return c.Put(key, res, ttl)
}

func (c *Cache) Put(key string, value string, ttl time.Duration) (string, error) {
	conf, err := c.load()
	if err != nil {
		conf = map[string]CacheEntry{}
	}
	conf[key] = CacheEntry{
		Expire: time.Now().Add(ttl),
		Value:  value,
	}
	slog.Debug("CACHE", "new item", fmt.Sprintf("%s: %s", key, value))
	err = c.save(conf)
	if err != nil {
		slog.Debug("CACHE", "failed to save", err)
	}
	return value, nil
}

func (c *Cache) Clear() error {
	return os.Remove(c.Root)
}
