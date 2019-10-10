package wconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
)

// Wrapper contains our ini files appended/overwritten and the current environment
type wrapper struct {
	cfg         *ini.File
	environment string
}

// Config is our application Wrapper object
var Config = wrapper{}

// ReadConfigFile reads the ini file
func (config *wrapper) ReadConfigFile(baseFolder string, envFlag string) error {
	var err error

	config.cfg, err = ini.Load(filepath.FromSlash(baseFolder)+"config.common.ini", filepath.FromSlash(baseFolder)+"config."+envFlag+".ini")
	if err != nil {
		fmt.Printf("Failed to read config file: %v", err)
		os.Exit(1)
	}
	return err
}

// Get returns the value of a key (as a string)
func (config *wrapper) Get(section string, key string) (string, error) {
	val, err := config.cfg.Section(section).GetKey(key)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

// GetUnsafe returns the value of a key (as a string)
func (config *wrapper) GetUnsafe(section string, key string) string {
	val, err := config.Get(section, key)
	if err != nil {
		panic("Config section: '" + section + "', key: '" + key + "' does not exist")
	}
	return val
}

// GetArray returns an array of values from a key (as a array of string)
func (config *wrapper) GetArray(section string, key string) ([]string, error) {
	val, err := config.cfg.Section(section).GetKey(key)
	if err != nil {
		return nil, err
	}
	return val.Strings(","), nil
}

// GetPrefixedMap returns a hash of the keys and values beginning with a prefix (as an array of strings)
func (config *wrapper) GetPrefixedMap(section string, prefix string) map[string]string {
	vals := config.cfg.Section(section).KeysHash()
	var output = make(map[string]string)
	for key, val := range vals {
		if strings.HasPrefix(key, prefix) {
			output[strings.Replace(key, prefix+".", "", 1)] = val
		}
	}
	return output
}

func (config *wrapper) SetEnvironment(environment string) {
	config.environment = environment
}

func (config *wrapper) GetEnvironment() string {
	return config.environment
}
