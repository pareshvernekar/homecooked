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

// checkDefaultValues checks if the default values are set correctly
func checkDefaultValues(t *testing.T) {
	expectedDefaults := map[string]interface{}{
		"APP.DEBUG":               false,
		"APP.NAME":                "Food Menu System",
		"SERVER.PORT":             8080,
		"DB.DRIVER":               "mysql",
		"DB.HOST":                 "localhost",
		"DB.PORT":                 3306,
		"DB.NAME":                 "food_menu_system",
		"DB.USER":                 "root",
		"DB.PASSWORD":             "password",
		"TENANT_ID_HEADER":        "X-Tenant-ID",
		"DEFAULT_TENANT_ID":       "default_tenant",
		"LOG_LEVEL":               "debug",
		"LOG_FORMAT":              "json",
		"JWT_SECRET":              "your_jwt_secret_here",
		"JWT_EXPIRATION":          "8h", // Ensure this is a string
		"CACHE_ENABLED":           false,
		"CACHE_TTL":               "5m",
		"TENANT_ISOLATION_ENABLED": true,
		"TENANT_ISOLATION_STRATEGY": "row_level_security",
	}

	for key, expectedValue := range expectedDefaults {
		value := Config.Get(key)
		if value != expectedValue {
			t.Errorf("Config.Get(%s) = %v, want %v", key, value, expectedValue)
		}
	}
}

// checkLoadedValues checks if the loaded values from a YAML file are correct
func checkLoadedValues(t *testing.T) {
	expectedLoadedValues := map[string]interface{}{
		"APP.DEBUG": false,
		"SERVER.PORT": 8080,
		"DATABASE.DRIVER": "mysql",
		"DATABASE.HOST": "localhost",
	}

	for key, expectedValue := range expectedLoadedValues {
		value := Config.Get(key)
		if value != expectedValue {
			t.Errorf("Config.Get(%s) = %v, want %v", key, value, expectedValue)
		}
	}
}

// slicesEqual checks if two string slices are equal
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
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