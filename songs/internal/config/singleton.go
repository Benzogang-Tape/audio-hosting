package config

var instance Config

// Get returns the singleton instance of the configuration.
//
// Make sure, you have called [Load] before this,
// otherwise it will return nil.
func Get() Config {
	return instance
}

// Set sets the singleton instance of the configuration.
func Set(cfg Config) {
	instance = cfg
}
