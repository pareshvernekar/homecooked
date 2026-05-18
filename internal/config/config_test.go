package config

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
)

// TestInit ensures that the Init function sets up Viper correctly
func TestInit(t *testing.T) {
	Init()
	// Check if the Config variable is set
	if Config == nil {
		t.Fatalf("Config should not be nil after Init")
	}
}

// TestRead ensures that the Read function reads the configuration file correctly
func TestRead(t *testing.T) {

	Init()
	// Check if the Config variable is set
	if Config == nil {
		t.Fatalf("Config should not be nil after Init")
	}
	// Read the configuration
	if err := Read(); err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	fmt.Printf("Loaded Configuration: %+v\n", Config.AllSettings()) // Debug: Print all loaded settings

	// Check if the loaded configuration matches the expected values
	checkLoadedValues(t)
}

// checkLoadedValues checks if the loaded values from a YAML file are correct
func checkLoadedValues(t *testing.T) {
	expectedLoadedValues := map[string]interface{}{
		"APP.DEBUG":       true,
		"SERVER.PORT":     8080,
		"DATABASE.DRIVER": "postgres",
		"DATABASE.HOST":   "localhost",
	}

	for key, expectedValue := range expectedLoadedValues {
		value := Config.Get(key)
		if value != expectedValue {
			t.Errorf("Config.Get(%s) = %v, want %v", key, value, expectedValue)
		}
	}
}

// InitWithCleanViper resets Viper configuration and initializes it properly for tests
func InitWithCleanViper(tempDir string) {
	fmt.Println("Initializing Viper with a clean state for testing...", tempDir) // Debug: Indicate initialization
	viper.Reset()
	Config = viper.New()
	// Additional initialization logic if needed
	Config.SetConfigName("config")
	Config.SetConfigType("yaml")
	Config.AddConfigPath(tempDir)
	Config.AutomaticEnv()
}
