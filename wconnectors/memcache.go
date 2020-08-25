package wconnectors

import (
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/bradfitz/gomemcache/memcache"
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
func (memcacheConnection MemcacheConnection) Set(key string, value []byte) {
	memcacheConnection.client.Set(&memcache.Item{Key: key, Value: []byte(value)})
}

// Get stores a value
func (memcacheConnection MemcacheConnection) Get(key string, value []byte) (item *memcache.Item, err error) {
	return memcacheConnection.client.Get(key)
}

// GetClient returns the original Memcache client
func (memcacheConnection MemcacheConnection) GetClient() *memcache.Client {
	return memcacheConnection.client
}
