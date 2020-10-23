package wconnectors

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/webediads/adsgolib/wlog"
)

var memcacheConnections map[string]*MemcacheConnection
var memcacheOnce map[string]bool
var memcacheOnceMutex sync.Mutex

// memcacheHostSettings is the struct that is used for registering a connection
type memcacheHostSettings struct {
	Host string
	Port int
}

type memcacheConnectionSettings []memcacheHostSettings

var allMemcacheSettings = make(map[string]memcacheConnectionSettings)

// MemcacheConnection is our abstraction to memcache.Client
type MemcacheConnection struct {
	settings memcacheConnectionSettings
	client   *memcache.Client
}

// Memcache returns a memcache client connection
func Memcache(name string) *MemcacheConnection {

	if name == "" {
		log.Println("Memcache's name cannot be empty")
		return nil
	}

	if len(memcacheOnce) == 0 {
		memcacheOnce = make(map[string]bool, 15)
		memcacheConnections = make(map[string]*MemcacheConnection, 15)
	}

	memcacheOnceMutex.Lock()
	if !memcacheOnce[name] {
		memcacheOnce[name] = true
		mcConnection := new(MemcacheConnection)
		var connectionStrings []string
		for _, memcacheSettingsForHost := range allMemcacheSettings[name] {
			connectionStrings = append(connectionStrings, memcacheSettingsForHost.Host+":"+strconv.Itoa(memcacheSettingsForHost.Port))
		}
		memcacheClient := memcache.New(strings.Join(connectionStrings, ","))
		mcConnection.client = memcacheClient
		mcConnection.settings = allMemcacheSettings[name]
		memcacheConnections[name] = mcConnection
		memcacheOnceMutex.Unlock()
	} else {
		memcacheOnceMutex.Unlock()
	}

	return memcacheConnections[name]
}

// RegisterMemcache registers the settings for a connection name
// ex : wconnectors.RegisterMemcache("global", wconfig.Config.GetUnsafe("cache", "memcache.global"))
func RegisterMemcache(name string, settingsString string) {
	var newMemcacheConnectionSettings memcacheConnectionSettings
	settingsArr := strings.Split(settingsString, ",")
	for _, settingsEntry := range settingsArr {
		settingsEntry = strings.TrimSpace(settingsEntry)
		settingsEntryArr := strings.Split(settingsEntry, ":")
		port, _ := strconv.Atoi(settingsEntryArr[1])
		newMemcacheConnectionSettings = append(newMemcacheConnectionSettings, memcacheHostSettings{Host: settingsEntryArr[0], Port: port})
	}
	allMemcacheSettings[name] = newMemcacheConnectionSettings
}

// Set stores a value
func (memcacheConnection MemcacheConnection) Set(key string, value []byte, expirationSecondsOpt ...int32) error {
	var expirationSeconds int32
	if len(expirationSecondsOpt) > 0 {
		expirationSeconds = expirationSecondsOpt[0]
	} else {
		expirationSeconds = 3600
	}
	mcErr := memcacheConnection.client.Set(&memcache.Item{Key: key, Value: []byte(value), Expiration: int32(expirationSeconds)})
	if mcErr != nil {
		wlog.GetLogger().Notice("memcache error set", nil, nil)
	}
	return mcErr
}

// Get stores a value
func (memcacheConnection MemcacheConnection) Get(key string) ([]byte, error) {
	i, err := memcacheConnection.client.Get(key)
	if err != nil {
		if err != memcache.ErrCacheMiss {
			wlog.GetLogger().Notice("memcache error get: "+err.Error(), nil, nil)
			return []byte(""), err
		}
	}
	return i.Value, nil
}

// GetClient returns the original Memcache client
func (memcacheConnection MemcacheConnection) GetClient() *memcache.Client {
	return memcacheConnection.client
}
