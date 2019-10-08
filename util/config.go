package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
)

// Config contains our ini files appended/overwritten and the current environment
type Config struct {
	cfg         *ini.File
	environment string
}

// ReadConfigFile reads the ini file
func (config *Config) ReadConfigFile(baseFolder string, envFlag string) error {
	var err error

	config.cfg, err = ini.Load(filepath.FromSlash(baseFolder)+"config.common.ini", filepath.FromSlash(baseFolder)+"config."+envFlag+".ini")
	if err != nil {
		fmt.Printf("Failed to read config file: %v", err)
		os.Exit(1)
	}
	return err
}

// Get returns the value of a key (as a string)
func (config *Config) Get(section string, key string) (string, error) {
	val, err := config.cfg.Section(section).GetKey(key)
	if err != nil {
		return "", err
	}
	return val.String(), nil
}

// GetArray returns an array of values from a key (as a array of string)
func (config *Config) GetArray(section string, key string) ([]string, error) {
	val, err := config.cfg.Section(section).GetKey(key)
	if err != nil {
		return nil, err
	}
	return val.Strings(","), nil
}

// GetPrefixedMap returns a hash of the keys and values beginning with a prefix (as an array of strings)
func (config *Config) GetPrefixedMap(section string, prefix string) map[string]string {
	vals := config.cfg.Section(section).KeysHash()
	var output = make(map[string]string)
	for key, val := range vals {
		if strings.HasPrefix(key, prefix) {
			output[strings.Replace(key, prefix+".", "", 1)] = val
		}
	}
	return output
}
