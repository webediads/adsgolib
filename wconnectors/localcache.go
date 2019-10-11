package wconnectors

// fork from https://github.com/facebookarchive/inmem/blob/master/inmem.go

import (
	"reflect"
	"strconv"
	"strings"
	"sync"

	"container/list"
	"time"

	"git.webedia-group.net/tools/adsgolib/wconfig"
)

// Cache of things.
type Cache interface {
	Set(key, value interface{})
	Get(key interface{}) (interface{}, bool)
	Remove(key interface{})
	Len() int
}

// cache implements a non-thread safe fixed size cache.
type cache struct {
	size  int
	lru   *list.List
	items map[interface{}]*list.Element
	ttl   int
}

// entry in the cache.
type entry struct {
	key       interface{}
	value     interface{}
	expiresAt time.Time
}

// NewUnlocked constructs a new Cache of the given size that is not safe for
// concurrent use. If will panic if size is not a positive integer.
func NewUnlocked(size int, ttl int) Cache {
	if size <= 0 {
		panic("must provide a positive size")
	}
	return &cache{
		size:  size,
		lru:   list.New(),
		items: make(map[interface{}]*list.Element),
		ttl:   ttl,
	}
}

func (c *cache) Set(key, value interface{}) {

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Ptr {
		panic("Cannot set a pointer as a cache entry")
	}

	expiresAt := time.Now().Add(time.Duration(c.ttl) * time.Second)
	if ent, ok := c.items[key]; ok {
		// update existing entry
		c.lru.MoveToFront(ent)
		v := ent.Value.(*entry)
		v.value = value
		v.expiresAt = expiresAt
		return
	}

	// set new entry
	c.items[key] = c.lru.PushFront(&entry{
		key:       key,
		value:     value,
		expiresAt: expiresAt,
	})

	// remove oldest
	if c.lru.Len() > c.size {
		ent := c.lru.Back()
		if ent != nil {
			c.removeElement(ent)
		}
	}
}

func (c *cache) Get(key interface{}) (interface{}, bool) {

	if ent, ok := c.items[key]; ok {
		v := ent.Value.(*entry)

		if v.expiresAt.After(time.Now()) {
			// found good entry
			// on commente cette ligne, en multi ça fout la merde
			// c.lru.MoveToFront(ent)
			return v.value, true
		}

		// ttl expired
		c.removeElement(ent)
	}
	return nil, false
}

func (c *cache) Remove(key interface{}) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
	}
}

func (c *cache) Len() int {
	return c.lru.Len()
}

// removeElement is used to remove a given list element from the cache
func (c *cache) removeElement(e *list.Element) {
	c.lru.Remove(e)
	kv := e.Value.(*entry)
	delete(c.items, kv.key)
}

type lockedCache struct {
	c cache
	m sync.Mutex
}

// NewLocked constructs a new Cache of the given size that is safe for
// concurrent use. If will panic if size is not a positive integer.
func NewLocked(size int, ttl int) Cache {
	if size <= 0 {
		panic("must provide a positive size")
	}
	return &lockedCache{
		c: cache{
			size:  size,
			lru:   list.New(),
			items: make(map[interface{}]*list.Element),
			ttl:   ttl,
		},
	}
}

func (l *lockedCache) Set(key, value interface{}) {
	l.m.Lock()
	l.c.Set(key, value)
	l.m.Unlock()
}

func (l *lockedCache) Get(key interface{}) (interface{}, bool) {
	l.m.Lock()
	v, f := l.c.Get(key)
	l.m.Unlock()
	return v, f
}

func (l *lockedCache) Remove(key interface{}) {
	l.m.Lock()
	l.c.Remove(key)
	l.m.Unlock()
}

func (l *lockedCache) Len() int {
	l.m.Lock()
	c := l.c.Len()
	l.m.Unlock()
	return c
}

// map of all caches
var cacheConnections map[string]Cache
var cacheOnce map[string]bool
var cacheOnceMutex sync.Mutex

// LocalCacheSettings is the struct that is used for registering a connection
type LocalCacheSettings struct {
	Size int
	TTL  int
}

var allLocalCacheSettings = make(map[string]LocalCacheSettings)

// RegisterLocalCache registers a db connection
func RegisterLocalCache(name string, settings LocalCacheSettings) {
	allLocalCacheSettings[name] = settings
}

// RegisterLocalCaches registers all the entries from the config or a map
func RegisterLocalCaches(localCacheConfigEntries map[string]string) {
	newCacheEntries := make(map[string]bool)
	for configCacheKey := range localCacheConfigEntries {
		cacheName := strings.Split(configCacheKey, ".")[1]
		if !newCacheEntries[cacheName] {
			newCacheEntries[cacheName] = true
			cacheSize, _ := strconv.Atoi(wconfig.Config.GetUnsafe("cache", "localcache."+cacheName+".size"))
			cacheTTL, _ := strconv.Atoi(wconfig.Config.GetUnsafe("cache", "localcache."+cacheName+".ttl"))
			RegisterLocalCache(cacheName, LocalCacheSettings{
				Size: cacheSize,
				TTL:  cacheTTL,
			})
		}
	}
}

// LocalCache return a cache
func LocalCache(name string) Cache {

	if name == "" {
		panic("LocalCache's name cannot be empty")
	}

	localCacheSettings, ok := allLocalCacheSettings[name]
	if !ok {
		panic("This LocalCache '" + name + "' was not registered")
	}

	if len(cacheOnce) == 0 {
		cacheOnce = make(map[string]bool, 50)
		cacheConnections = make(map[string]Cache, 50)
	}
	cacheOnceMutex.Lock()
	if !cacheOnce[name] {
		cacheOnce[name] = true
		cacheConnections[name] = NewLocked(localCacheSettings.Size, localCacheSettings.TTL)
	}
	cacheOnceMutex.Unlock()
	return cacheConnections[name]

}
