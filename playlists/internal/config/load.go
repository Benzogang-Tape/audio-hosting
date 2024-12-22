package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Load loads the configuration from the specified paths.
//
// The following order is used to find the config file:
//  1. The path specified by the `config-path` flag
//  2. The path specified by the `CONFIG_PATH` environment variable
//  4. The default paths passed into the function.
//
// Returns:
// - The configuration
// - If the config file is not found, [os.ErrNotExist] is returned.
// - If the config file is invalid, an error is returned.
func Load(paths ...string) (*Config, error) {
	var config Config

	osStatExist := func(name string) bool {
		_, err := os.Stat(name)
		return !os.IsNotExist(err)
	}

	path, err := getPath(osStatExist, paths...)
	if err != nil {
		return nil, err
	}

	// ReadConfig ovverides values by the envs and uses env-default if not set
	err = cleanenv.ReadConfig(path, &config)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &config, nil
}

// existFunc is a function that checks if a file exists.
type existFunc func(name string) bool

// findFile finds the config file in the following order
// until `exist` returns true:
//  1. The path specified by the `config-path` flag
//  2. The path specified by the `CONFIG_PATH` environment variable
//  3. The default paths
//
// Returns:
// - The path to the config file
// - If the file is not found, [os.ErrNotExist] is returned.
func getPath(exist existFunc, defaults ...string) (string, error) {
	var configPath string

	flag.StringVar(&configPath, "config-path", "", "path to config file")
	flag.Parse()

	if exist(configPath) {
		return configPath, nil
	}

	configPath = os.Getenv("CONFIG_PATH")
	if exist(configPath) {
		return configPath, nil
	}

	for _, path := range defaults {
		if exist(path) {
			return path, nil
		}
	}

	return "", os.ErrNotExist
}
