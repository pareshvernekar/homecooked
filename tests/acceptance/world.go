//go:build acceptance

package acceptance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pareshvernekar/homecooked/internal/models"
)

const (
	defaultBaseURL  = "http://localhost:8080"
	defaultTenantID = "1"
)

// apiWorld holds per-scenario HTTP acceptance state against a running local server.
type apiWorld struct {
	baseURL    string
	tenantID   string
	httpClient *http.Client

	lastStatus int
	lastBody   []byte

	lastCategoryID string
	lastFoodItemID string
	lastUniqueName string
}

func newAPIWorld() *apiWorld {
	baseURL := os.Getenv("ACCEPTANCE_BASE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	tenantID := os.Getenv("ACCEPTANCE_TENANT_ID")
	if tenantID == "" {
		tenantID = defaultTenantID
	}

	return &apiWorld{
		baseURL:  strings.TrimRight(baseURL, "/"),
		tenantID: tenantID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *apiWorld) resetScenario() {
	w.lastStatus = 0
	w.lastBody = nil
	w.lastCategoryID = ""
	w.lastFoodItemID = ""
	w.lastUniqueName = ""
}

func (w *apiWorld) uniqueName(prefix string) string {
	name := fmt.Sprintf("%s-%s", prefix, uuid.New().String()[:8])
	w.lastUniqueName = name
	return name
}

func (w *apiWorld) doRequest(method, path, body string) error {
	url := w.baseURL + path

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", w.tenantID)

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request to %s failed (is the local server running at %s?): %w", url, w.baseURL, err)
	}
	defer resp.Body.Close()

	w.lastStatus = resp.StatusCode
	w.lastBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	// Keep a stable copy even if caller mutates later.
	w.lastBody = bytes.Clone(w.lastBody)
	return nil
}

func (w *apiWorld) pingServer() error {
	// Categories list is a cheap readiness probe for the local API.
	if err := w.doRequest(http.MethodGet, "/api/v1/categories", ""); err != nil {
		return err
	}
	if w.lastStatus >= 500 {
		return fmt.Errorf("local server at %s returned %d for GET /api/v1/categories: %s", w.baseURL, w.lastStatus, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) successFlag() (bool, error) {
	if len(w.lastBody) == 0 {
		return false, fmt.Errorf("empty response body")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return false, err
	}
	success, ok := payload["success"].(bool)
	if !ok {
		return false, fmt.Errorf("success flag missing or not bool in body: %s", string(w.lastBody))
	}
	return success, nil
}

func (w *apiWorld) extractDataID() (string, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return "", err
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("response data is not an object: %s", string(w.lastBody))
	}
	id, ok := data["id"].(string)
	if !ok || id == "" {
		return "", fmt.Errorf("response data.id missing: %s", string(w.lastBody))
	}
	return id, nil
}

func (w *apiWorld) createCategoryViaAPI(namePrefix string) error {
	name := namePrefix
	if models.IsValidCategory(namePrefix) {
		w.lastUniqueName = namePrefix
	} else {
		name = w.uniqueName(namePrefix)
	}

	body := fmt.Sprintf(`{"name":%q,"description":%q}`, name, namePrefix+" description")
	if err := w.doRequest(http.MethodPost, "/api/v1/categories", body); err != nil {
		return err
	}

	if w.lastStatus == http.StatusCreated {
		id, err := w.extractDataID()
		if err != nil {
			return err
		}
		w.lastCategoryID = id
		w.lastUniqueName = name
		return nil
	}

	// Valid enum names are globally unique; reuse an existing row when re-running against a local DB.
	if models.IsValidCategory(namePrefix) {
		id, err := w.findCategoryIDByName(namePrefix)
		if err != nil {
			return fmt.Errorf("create category got %d and lookup failed: %w (body=%s)", w.lastStatus, err, string(w.lastBody))
		}
		w.lastCategoryID = id
		w.lastUniqueName = namePrefix
		return nil
	}

	return fmt.Errorf("create category expected 201, got %d body=%s", w.lastStatus, string(w.lastBody))
}

func (w *apiWorld) findCategoryIDByName(name string) (string, error) {
	if err := w.doRequest(http.MethodGet, "/api/v1/categories", ""); err != nil {
		return "", err
	}
	if w.lastStatus != http.StatusOK {
		return "", fmt.Errorf("list categories returned %d: %s", w.lastStatus, string(w.lastBody))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return "", err
	}
	items, ok := payload["data"].([]interface{})
	if !ok {
		return "", fmt.Errorf("categories data is not an array: %s", string(w.lastBody))
	}
	for _, item := range items {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if obj["name"] == name {
			id, _ := obj["id"].(string)
			if id != "" {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("category %q not found", name)
}

func (w *apiWorld) createFoodItemViaAPI(namePrefix string) error {
	categoryName := w.lastUniqueName
	if categoryName == "" {
		if err := w.createCategoryViaAPI("vegetarian"); err != nil {
			return err
		}
		categoryName = w.lastUniqueName
	}

	itemName := w.uniqueName(namePrefix)
	body := fmt.Sprintf(
		`{"name":%q,"description":%q,"price":10.0,"category_name":%q,"availability_status":"available"}`,
		itemName, namePrefix+" description", categoryName,
	)
	if err := w.doRequest(http.MethodPost, "/api/v1/food-items", body); err != nil {
		return err
	}
	if w.lastStatus != http.StatusCreated {
		return fmt.Errorf("create food item expected 201, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	id, err := w.extractDataID()
	if err != nil {
		return err
	}
	w.lastFoodItemID = id
	return nil
}
