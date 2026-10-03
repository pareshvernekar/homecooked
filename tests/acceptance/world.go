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

	lastCategoryID   string
	lastCategoryName string
	lastFoodItemID   string
	lastFoodItemID2  string
	lastUniqueName   string
	lastMenuID       string
	lastMenuCategory string
	lastSizeUnitID   string

	lastMenuItemID    string
	lastComponentID   string
	lastSizeOptionID  string
	lastOrderID       string
	lastOrderItemID   string

	rememberedNotificationCount int
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
	w.lastCategoryName = ""
	w.lastFoodItemID = ""
	w.lastFoodItemID2 = ""
	w.lastUniqueName = ""
	w.lastMenuID = ""
	w.lastMenuCategory = ""
	w.lastSizeUnitID = ""
	w.lastMenuItemID = ""
	w.lastComponentID = ""
	w.lastSizeOptionID = ""
	w.lastOrderID = ""
	w.lastOrderItemID = ""
	w.rememberedNotificationCount = 0
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
		w.lastCategoryName = name
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
		w.lastCategoryName = namePrefix
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
	categoryName := w.lastCategoryName
	if categoryName == "" {
		if err := w.createCategoryViaAPI("vegetarian"); err != nil {
			return err
		}
		categoryName = w.lastCategoryName
	}

	itemName := w.uniqueName(namePrefix)
	body := fmt.Sprintf(
		`{"name":%q,"description":%q,"category_name":%q,"availability_status":"available"}`,
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

func (w *apiWorld) createSecondFoodItemViaAPI(namePrefix string) error {
	categoryName := w.lastCategoryName
	if categoryName == "" {
		if err := w.createCategoryViaAPI("vegetarian"); err != nil {
			return err
		}
		categoryName = w.lastCategoryName
	}
	itemName := w.uniqueName(namePrefix)
	body := fmt.Sprintf(
		`{"name":%q,"description":%q,"category_name":%q,"availability_status":"available"}`,
		itemName, namePrefix+" description", categoryName,
	)
	if err := w.doRequest(http.MethodPost, "/api/v1/food-items", body); err != nil {
		return err
	}
	if w.lastStatus != http.StatusCreated {
		return fmt.Errorf("create second food item expected 201, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	id, err := w.extractDataID()
	if err != nil {
		return err
	}
	w.lastFoodItemID2 = id
	return nil
}

func (w *apiWorld) createDraftDailyMenu(namePrefix, categoryName string) error {
	menuName := w.uniqueName(namePrefix)
	menuDate := time.Now().UTC().Format("2006-01-02")
	body := fmt.Sprintf(
		`{"name":%q,"menu_type":"daily","menu_date":%q,"categories":[{"name":%q,"sequence":1}]}`,
		menuName, menuDate, categoryName,
	)
	if err := w.doRequest(http.MethodPost, "/api/v1/menus", body); err != nil {
		return err
	}
	createStatus := w.lastStatus
	createBody := append([]byte(nil), w.lastBody...)
	if createStatus != http.StatusCreated {
		return fmt.Errorf("create menu expected 201, got %d body=%s", createStatus, string(createBody))
	}
	id, err := w.extractDataID()
	if err != nil {
		return err
	}
	w.lastMenuID = id

	if err := w.doRequest(http.MethodGet, "/api/v1/menus/"+w.lastMenuID, ""); err != nil {
		return err
	}
	if w.lastStatus != http.StatusOK {
		return fmt.Errorf("get menu tree expected 200, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return err
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("menu tree data is not an object: %s", string(w.lastBody))
	}
	cats, ok := data["categories"].([]interface{})
	if !ok || len(cats) == 0 {
		return fmt.Errorf("menu tree has no categories: %s", string(w.lastBody))
	}
	cat, ok := cats[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("first category is not an object")
	}
	catID, _ := cat["id"].(string)
	if catID == "" {
		return fmt.Errorf("menu category id missing")
	}
	w.lastMenuCategory = catID

	// Restore create response for subsequent Then assertions.
	w.lastStatus = createStatus
	w.lastBody = createBody
	return nil
}

func (w *apiWorld) addComboMenuItem(name string, priceA, priceB float64) error {
	if w.lastMenuID == "" || w.lastMenuCategory == "" {
		return fmt.Errorf("no created menu/category available")
	}
	if w.lastFoodItemID == "" || w.lastFoodItemID2 == "" {
		return fmt.Errorf("need two food items for combo")
	}
	body := fmt.Sprintf(`{
		"category_id":%q,
		"kind":"combo",
		"name":%q,
		"components":[
			{"food_item_id":%q,"default_size_index":0,"size_options":[{"size_unit_id":"su_serving","qty":1,"price":%v}]},
			{"food_item_id":%q,"default_size_index":0,"size_options":[{"size_unit_id":"su_serving","qty":1,"price":%v}]}
		]
	}`, w.lastMenuCategory, name, w.lastFoodItemID, priceA, w.lastFoodItemID2, priceB)
	return w.doRequest(http.MethodPost, "/api/v1/menus/"+w.lastMenuID+"/items", body)
}

func (w *apiWorld) addSimpleMenuItem(name string, price float64) error {
	if w.lastMenuID == "" || w.lastMenuCategory == "" {
		return fmt.Errorf("no created menu/category available")
	}
	if w.lastFoodItemID == "" {
		return fmt.Errorf("no food item available")
	}
	body := fmt.Sprintf(`{
		"category_id":%q,
		"kind":"simple",
		"name":%q,
		"components":[
			{"food_item_id":%q,"default_size_index":0,"size_options":[{"size_unit_id":"su_serving","qty":1,"price":%v}]}
		]
	}`, w.lastMenuCategory, name, w.lastFoodItemID, price)
	return w.doRequest(http.MethodPost, "/api/v1/menus/"+w.lastMenuID+"/items", body)
}

func (w *apiWorld) menuStatusFromLastBody() (string, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return "", err
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data is not an object: %s", string(w.lastBody))
	}
	status, _ := data["status"].(string)
	if status == "" {
		return "", fmt.Errorf("status missing: %s", string(w.lastBody))
	}
	return status, nil
}

func (w *apiWorld) firstItemDefaultTotal() (float64, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return 0, err
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("data is not an object")
	}
	cats, ok := data["categories"].([]interface{})
	if !ok || len(cats) == 0 {
		return 0, fmt.Errorf("no categories")
	}
	cat := cats[0].(map[string]interface{})
	items, ok := cat["items"].([]interface{})
	if !ok || len(items) == 0 {
		return 0, fmt.Errorf("no items")
	}
	item := items[0].(map[string]interface{})
	total, ok := item["default_total"].(float64)
	if !ok {
		return 0, fmt.Errorf("default_total missing: %v", item)
	}
	return total, nil
}

func (w *apiWorld) listContainsCreatedMenu() (bool, error) {
	if w.lastMenuID == "" {
		return false, fmt.Errorf("no created menu id")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return false, err
	}
	data, ok := payload["data"].([]interface{})
	if !ok {
		// empty list may be null
		if payload["data"] == nil {
			return false, nil
		}
		return false, fmt.Errorf("list data is not an array: %s", string(w.lastBody))
	}
	for _, row := range data {
		obj, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		if obj["id"] == w.lastMenuID {
			return true, nil
		}
	}
	return false, nil
}

func (w *apiWorld) sizeUnitListHasCode(code string) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return err
	}
	data, ok := payload["data"].([]interface{})
	if !ok {
		return fmt.Errorf("size units data is not an array: %s", string(w.lastBody))
	}
	for _, row := range data {
		obj, ok := row.(map[string]interface{})
		if !ok {
			continue
		}
		if obj["code"] == code {
			return nil
		}
	}
	return fmt.Errorf("size unit code %q not found in %s", code, string(w.lastBody))
}

// seedPublishedDailySimpleMenu creates food item + daily menu + simple item + publish.
// REQORDER001 background setup for acceptance.
func (w *apiWorld) seedPublishedDailySimpleMenu(price float64) error {
	if err := w.createFoodItemViaAPI("Order Rice"); err != nil {
		return err
	}
	if err := w.createDraftDailyMenu("Order Daily", "Mains"); err != nil {
		return err
	}
	if err := w.addSimpleMenuItem("Rice Bowl", price); err != nil {
		return err
	}
	if w.lastStatus != http.StatusCreated {
		return fmt.Errorf("add menu item expected 201, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	if err := w.doRequest(http.MethodPost, "/api/v1/menus/"+w.lastMenuID+"/publish", ""); err != nil {
		return err
	}
	if w.lastStatus != http.StatusOK {
		return fmt.Errorf("publish menu expected 200, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	return w.captureFirstMenuItemIDs()
}

func (w *apiWorld) captureFirstMenuItemIDs() error {
	if err := w.doRequest(http.MethodGet, "/api/v1/menus/"+w.lastMenuID, ""); err != nil {
		return err
	}
	if w.lastStatus != http.StatusOK {
		return fmt.Errorf("get menu tree expected 200, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return err
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("menu tree data is not an object")
	}
	cats, ok := data["categories"].([]interface{})
	if !ok || len(cats) == 0 {
		return fmt.Errorf("menu tree has no categories")
	}
	cat := cats[0].(map[string]interface{})
	items, ok := cat["items"].([]interface{})
	if !ok || len(items) == 0 {
		return fmt.Errorf("menu has no items")
	}
	item := items[0].(map[string]interface{})
	itemID, _ := item["id"].(string)
	comps, ok := item["components"].([]interface{})
	if !ok || len(comps) == 0 {
		return fmt.Errorf("menu item has no components")
	}
	comp := comps[0].(map[string]interface{})
	compID, _ := comp["id"].(string)
	opts, ok := comp["size_options"].([]interface{})
	if !ok || len(opts) == 0 {
		return fmt.Errorf("component has no size options")
	}
	opt := opts[0].(map[string]interface{})
	optID, _ := opt["id"].(string)
	if itemID == "" || compID == "" || optID == "" {
		return fmt.Errorf("missing menu item/component/option ids")
	}
	w.lastMenuItemID = itemID
	w.lastComponentID = compID
	w.lastSizeOptionID = optID
	return nil
}

func (w *apiWorld) createOrderForPublishedMenu(customerName, phone string, hoursAhead int) error {
	if w.lastMenuID == "" {
		return fmt.Errorf("no published menu id")
	}
	expectedAt := time.Now().UTC().Add(time.Duration(hoursAhead) * time.Hour).UnixMilli()
	body := fmt.Sprintf(
		`{"menu_id":%q,"customer_name":%q,"customer_phone":%q,"expected_at":%d}`,
		w.lastMenuID, customerName, phone, expectedAt,
	)
	if err := w.doRequest(http.MethodPost, "/api/v1/orders", body); err != nil {
		return err
	}
	if w.lastStatus == http.StatusCreated {
		id, err := w.extractDataID()
		if err != nil {
			return err
		}
		w.lastOrderID = id
	}
	return nil
}

func (w *apiWorld) addOrderLine(quantity int) error {
	if w.lastOrderID == "" || w.lastMenuItemID == "" {
		return fmt.Errorf("need order and menu item ids")
	}
	body := fmt.Sprintf(`{"menu_item_id":%q,"quantity":%d}`, w.lastMenuItemID, quantity)
	if err := w.doRequest(http.MethodPost, "/api/v1/orders/"+w.lastOrderID+"/items", body); err != nil {
		return err
	}
	if w.lastStatus == http.StatusCreated {
		id, err := w.extractDataID()
		if err != nil {
			return err
		}
		w.lastOrderItemID = id
	}
	return nil
}

func (w *apiWorld) getCreatedOrder() error {
	if w.lastOrderID == "" {
		return fmt.Errorf("no created order id")
	}
	return w.doRequest(http.MethodGet, "/api/v1/orders/"+w.lastOrderID, "")
}

func (w *apiWorld) updatePublishedSizeOptionPrice(price float64) error {
	if w.lastMenuID == "" || w.lastMenuItemID == "" || w.lastComponentID == "" || w.lastSizeOptionID == "" {
		return fmt.Errorf("missing published size option path ids")
	}
	path := fmt.Sprintf(
		"/api/v1/menus/%s/items/%s/components/%s/size-options/%s",
		w.lastMenuID, w.lastMenuItemID, w.lastComponentID, w.lastSizeOptionID,
	)
	body := fmt.Sprintf(`{"price":%v}`, price)
	if err := w.doRequest(http.MethodPut, path, body); err != nil {
		return err
	}
	if w.lastStatus != http.StatusOK {
		return fmt.Errorf("update size option price expected 200, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) recordCashPayment(amount float64) error {
	if w.lastOrderID == "" {
		return fmt.Errorf("no created order id")
	}
	body := fmt.Sprintf(`{"mode":"cash","amount":%v}`, amount)
	return w.doRequest(http.MethodPost, "/api/v1/orders/"+w.lastOrderID+"/payments", body)
}

// orderAction posts a lifecycle action (accept, start-preparing, ready, pickup) for the created order.
// REQLIFE002, REQLIFE004
func (w *apiWorld) orderAction(action string) error {
	if w.lastOrderID == "" {
		return fmt.Errorf("no created order id")
	}
	return w.doRequest(http.MethodPost, "/api/v1/orders/"+w.lastOrderID+"/"+action, "")
}

func (w *apiWorld) markOrderPickedUp() error {
	if w.lastOrderID == "" {
		return fmt.Errorf("no created order id")
	}
	return w.doRequest(http.MethodPost, "/api/v1/orders/"+w.lastOrderID+"/pickup", "")
}

func (w *apiWorld) refuseOrder(reason string) error {
	if w.lastOrderID == "" {
		return fmt.Errorf("no created order id")
	}
	body := fmt.Sprintf(`{"reason":%q}`, reason)
	return w.doRequest(http.MethodPost, "/api/v1/orders/"+w.lastOrderID+"/refuse", body)
}

// setCookAdminPhone PUT /tenant/settings. Empty phone clears (REQNOTIF001).
func (w *apiWorld) setCookAdminPhone(phone string) error {
	var body string
	if phone == "" {
		body = `{"cook_admin_phone":null}`
	} else {
		body = fmt.Sprintf(`{"cook_admin_phone":%q}`, phone)
	}
	return w.doRequest(http.MethodPut, "/api/v1/tenant/settings", body)
}

func (w *apiWorld) getCookAdminPhone() (string, bool, error) {
	if err := w.doRequest(http.MethodGet, "/api/v1/tenant/settings", ""); err != nil {
		return "", false, err
	}
	data, err := w.dataObjectFromLastBody()
	if err != nil {
		return "", false, err
	}
	v, ok := data["cook_admin_phone"]
	if !ok || v == nil {
		return "", false, nil
	}
	s, ok := v.(string)
	if !ok {
		return "", false, fmt.Errorf("cook_admin_phone not a string: %s", string(w.lastBody))
	}
	return s, true, nil
}

func (w *apiWorld) listOrderNotifications(orderID string) error {
	if orderID == "" {
		return fmt.Errorf("no order id")
	}
	return w.doRequest(http.MethodGet, "/api/v1/orders/"+orderID+"/notifications", "")
}

func (w *apiWorld) notificationRowsFromLastBody() ([]map[string]interface{}, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return nil, err
	}
	raw, ok := payload["data"].([]interface{})
	if !ok {
		if payload["data"] == nil {
			return []map[string]interface{}{}, nil
		}
		return nil, fmt.Errorf("data is not a list: %s", string(w.lastBody))
	}
	out := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("notification row is not an object: %s", string(w.lastBody))
		}
		out = append(out, m)
	}
	return out, nil
}

func (w *apiWorld) notificationByEvent(event string) (map[string]interface{}, error) {
	rows, err := w.notificationRowsFromLastBody()
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		if r["event_type"] == event {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no notification for event %q in %s", event, string(w.lastBody))
}

func (w *apiWorld) waitForNotificationStatus(event, status string) error {
	if w.lastOrderID == "" {
		return fmt.Errorf("no created order id")
	}
	deadline := time.Now().Add(10 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := w.listOrderNotifications(w.lastOrderID); err != nil {
			return err
		}
		row, err := w.notificationByEvent(event)
		if err != nil {
			lastErr = err
		} else if got, _ := row["status"].(string); got == status {
			return nil
		} else {
			lastErr = fmt.Errorf("event %q status=%v want %q", event, row["status"], status)
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timed out waiting for %s status %q: %v", event, status, lastErr)
}

func (w *apiWorld) dataObjectFromLastBody() (map[string]interface{}, error) {
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return nil, err
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not an object: %s", string(w.lastBody))
	}
	return data, nil
}

func (w *apiWorld) orderStringField(field string) (string, error) {
	data, err := w.dataObjectFromLastBody()
	if err != nil {
		return "", err
	}
	v, _ := data[field].(string)
	if v == "" {
		return "", fmt.Errorf("%s missing or empty: %s", field, string(w.lastBody))
	}
	return v, nil
}

func (w *apiWorld) orderFloatField(field string) (float64, error) {
	data, err := w.dataObjectFromLastBody()
	if err != nil {
		return 0, err
	}
	v, ok := data[field].(float64)
	if !ok {
		return 0, fmt.Errorf("%s missing or not a number: %s", field, string(w.lastBody))
	}
	return v, nil
}

func (w *apiWorld) orderBoolField(field string) (bool, error) {
	data, err := w.dataObjectFromLastBody()
	if err != nil {
		return false, err
	}
	v, ok := data[field].(bool)
	if !ok {
		return false, fmt.Errorf("%s missing or not a bool: %s", field, string(w.lastBody))
	}
	return v, nil
}
