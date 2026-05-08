package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds the application configuration
var Config *viper.Viper

// Init initializes the Viper configuration
func Init() {
	Config = viper.New()
	Config.SetConfigName("config")
	Config.SetConfigType("yaml")
	Config.AddConfigPath("internal/config/configs")
	Config.AddConfigPath("configs")
	Config.AddConfigPath(".")
	Config.AddConfigPath("../config/configs")
	Config.AutomaticEnv()
	Config.SetEnvPrefix("APP")
	Config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
}

// Read reads the configuration file
func Read() error {
	if err := Config.ReadInConfig(); err != nil {
		return err
	}
	return nil
}

// Get retrieves a configuration value
func Get(key string) interface{} {
	return Config.Get(key)
}

// GetString retrieves a configuration value as string
func GetString(key string) string {
	return Config.GetString(key)
}

// GetInt retrieves a configuration value as integer
func GetInt(key string) int {
	return Config.GetInt(key)
}

// GetBool retrieves a configuration value as boolean
func GetBool(key string) bool {
	return Config.GetBool(key)
}

// GetFloat64 retrieves a configuration value as float64
func GetFloat64(key string) float64 {
	return Config.GetFloat64(key)
}

// GetStringSlice retrieves a configuration value as string slice
func GetStringSlice(key string) []string {
	return Config.GetStringSlice(key)
}

// Write writes the current configuration to a file
func Write(filePath string) error {
	return Config.WriteConfigAs(filePath)
}
