package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/pareshvernekar/homecooked/internal/config"
)

func TestTenantMiddleware(t *testing.T) {
	// Initialize the configuration
	config.Init()

	// Create a test handler that retrieves the tenant ID from the context
	testHandler := func(c *gin.Context) {
		tenantID, exists := c.Get(TenantIDKey)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tenant ID not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
	}

	// Create a test router and add the TenantMiddleware
	router := gin.Default()
	router.Use(TenantMiddleware())
	router.GET("/", testHandler)

	// Test cases
	tests := []struct {
		name     string
		header   string
		expected int
	}{
		{
			name:     "Tenant ID present",
			header:   "tenant1",
			expected: http.StatusOK,
		},
		{
			name:     "Tenant ID missing",
			header:   "",
			expected: http.StatusBadRequest,
		},
	}

	// Run the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the specified header
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("X-Tenant-ID", tt.header)

			// Create a response recorder
			recorder := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(recorder, req)

			// Check the response status code
			if recorder.Code != tt.expected {
				t.Errorf("Expected status code %d, but got %d", tt.expected, recorder.Code)
			}

			// Check the response body
			if tt.expected == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(recorder.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response body: %v", err)
				}
				if response["tenant_id"] != "tenant1" {
					t.Errorf("Expected tenant ID 'tenant1', but got %v", response["tenant_id"])
				}
			}
		})
	}
}