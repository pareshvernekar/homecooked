//go:build acceptance

package acceptance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/uuid"
)

func (w *apiWorld) theAPIIsReady() error {
	return w.pingServer()
}

func (w *apiWorld) theTenantHeaderIs(tenantID string) error {
	w.tenantID = tenantID
	return nil
}

func (w *apiWorld) aFoodCategoryExistsViaTheAPIWithName(name string) error {
	return w.createCategoryViaAPI(name)
}

func (w *apiWorld) aFoodItemExistsViaTheAPINamed(name string) error {
	return w.createFoodItemViaAPI(name)
}

func (w *apiWorld) anotherFoodItemExistsViaTheAPINamed(name string) error {
	return w.createSecondFoodItemViaAPI(name)
}

func (w *apiWorld) iCreateADraftDailyMenuNamedWithCategory(name, category string) error {
	return w.createDraftDailyMenu(name, category)
}

func (w *apiWorld) iAddAComboMenuItemNamedPricedAnd(name string, priceA, priceB float64) error {
	return w.addComboMenuItem(name, priceA, priceB)
}

func (w *apiWorld) iAddASimpleMenuItemNamedPriced(name string, price float64) error {
	return w.addSimpleMenuItem(name, price)
}

func (w *apiWorld) iPublishTheCreatedMenu() error {
	if w.lastMenuID == "" {
		return fmt.Errorf("no created menu id")
	}
	return w.doRequest(http.MethodPost, "/api/v1/menus/"+w.lastMenuID+"/publish", "")
}

func (w *apiWorld) iUnpublishTheCreatedMenu() error {
	if w.lastMenuID == "" {
		return fmt.Errorf("no created menu id")
	}
	return w.doRequest(http.MethodPost, "/api/v1/menus/"+w.lastMenuID+"/unpublish", "")
}

func (w *apiWorld) iGetTheCreatedMenuTree() error {
	if w.lastMenuID == "" {
		return fmt.Errorf("no created menu id")
	}
	return w.doRequest(http.MethodGet, "/api/v1/menus/"+w.lastMenuID, "")
}

func (w *apiWorld) theMenuStatusShouldBe(status string) error {
	got, err := w.menuStatusFromLastBody()
	if err != nil {
		return err
	}
	if got != status {
		return fmt.Errorf("expected menu status %q, got %q body=%s", status, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theFirstMenuItemDefaultTotalShouldBe(expected float64) error {
	got, err := w.firstItemDefaultTotal()
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("expected default_total %v, got %v body=%s", expected, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) iTryToAddASimpleItemToThePublishedMenu() error {
	return w.addSimpleMenuItem("Should Fail", 1)
}

func (w *apiWorld) theCreatedMenuShouldNotAppearInTheList() error {
	found, err := w.listContainsCreatedMenu()
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf("created menu %s unexpectedly present in list body=%s", w.lastMenuID, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theCreatedMenuShouldAppearInTheList() error {
	found, err := w.listContainsCreatedMenu()
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("created menu %s missing from list body=%s", w.lastMenuID, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theSizeUnitListShouldIncludeCode(code string) error {
	return w.sizeUnitListHasCode(code)
}

func (w *apiWorld) iCreateACustomSizeUnitWithDisplayName(displayName string) error {
	code := "acc-" + uuid.New().String()[:8]
	body := fmt.Sprintf(`{"code":%q,"display_name":%q}`, code, displayName)
	if err := w.doRequest(http.MethodPost, "/api/v1/size-units", body); err != nil {
		return err
	}
	if w.lastStatus == http.StatusCreated {
		if id, err := w.extractDataID(); err == nil {
			w.lastSizeUnitID = id
		}
	}
	return nil
}

func (w *apiWorld) theCreatedSizeUnitIsSystemShouldBeFalse() error {
	var payload map[string]interface{}
	if err := json.Unmarshal(w.lastBody, &payload); err != nil {
		return err
	}
	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("data is not an object: %s", string(w.lastBody))
	}
	isSystem, ok := data["is_system"].(bool)
	if !ok {
		return fmt.Errorf("is_system missing: %s", string(w.lastBody))
	}
	if isSystem {
		return fmt.Errorf("expected is_system=false")
	}
	return nil
}

func (w *apiWorld) iCreateAFoodCategoryNamedWithDescription(name, description string) error {
	unique := w.uniqueName(name)
	body := fmt.Sprintf(`{"name":%q,"description":%q}`, unique, description)
	if err := w.doRequest(http.MethodPost, "/api/v1/categories", body); err != nil {
		return err
	}
	if w.lastStatus == http.StatusCreated {
		if id, err := w.extractDataID(); err == nil {
			w.lastCategoryID = id
		}
		w.lastCategoryName = unique
	}
	return nil
}

func (w *apiWorld) iCreateAFoodItemNamed(name string) error {
	categoryName := w.lastCategoryName
	if categoryName == "" {
		if err := w.createCategoryViaAPI("vegetarian"); err != nil {
			return err
		}
		categoryName = w.lastCategoryName
	}
	unique := w.uniqueName(name)
	body := fmt.Sprintf(
		`{"name":%q,"description":%q,"category_name":%q,"availability_status":"available"}`,
		unique, name+" description", categoryName,
	)
	if err := w.doRequest(http.MethodPost, "/api/v1/food-items", body); err != nil {
		return err
	}
	if w.lastStatus == http.StatusCreated {
		if id, err := w.extractDataID(); err == nil {
			w.lastFoodItemID = id
		}
	}
	return nil
}

func (w *apiWorld) iSendARequestWithJSON(method, path string, body *godog.DocString) error {
	payload := ""
	if body != nil {
		payload = body.Content
	}
	return w.doRequest(method, path, payload)
}

func (w *apiWorld) iSendAGETRequestTo(path string) error {
	return w.doRequest(http.MethodGet, path, "")
}

func (w *apiWorld) iSendAPUTRequestToTheCreatedCategoryWithJSON(body *godog.DocString) error {
	if w.lastCategoryID == "" {
		return fmt.Errorf("no created category id available")
	}
	payload := ""
	if body != nil {
		payload = body.Content
	}
	return w.doRequest(http.MethodPut, "/api/v1/categories/"+w.lastCategoryID, payload)
}

func (w *apiWorld) iUpdateTheCreatedCategoryDescriptionTo(description string) error {
	if w.lastCategoryID == "" {
		return fmt.Errorf("no created category id available")
	}
	if w.lastUniqueName == "" {
		return fmt.Errorf("no created category name available")
	}
	body := fmt.Sprintf(`{"name":%q,"description":%q,"is_active":true}`, w.lastUniqueName, description)
	return w.doRequest(http.MethodPut, "/api/v1/categories/"+w.lastCategoryID, body)
}

func (w *apiWorld) iSendADELETERequestToTheCreatedCategory() error {
	if w.lastCategoryID == "" {
		return fmt.Errorf("no created category id available")
	}
	return w.doRequest(http.MethodDelete, "/api/v1/categories/"+w.lastCategoryID, "")
}

func (w *apiWorld) iSendAPUTRequestToTheCreatedFoodItemWithJSON(body *godog.DocString) error {
	if w.lastFoodItemID == "" {
		return fmt.Errorf("no created food item id available")
	}
	payload := ""
	if body != nil {
		payload = body.Content
	}
	return w.doRequest(http.MethodPut, "/api/v1/food-items/"+w.lastFoodItemID, payload)
}

func (w *apiWorld) iSendADELETERequestToTheCreatedFoodItem() error {
	if w.lastFoodItemID == "" {
		return fmt.Errorf("no created food item id available")
	}
	return w.doRequest(http.MethodDelete, "/api/v1/food-items/"+w.lastFoodItemID, "")
}

func (w *apiWorld) theResponseStatusCodeShouldBe(code string) error {
	expected, err := strconv.Atoi(code)
	if err != nil {
		return fmt.Errorf("invalid status code %q: %w", code, err)
	}
	if w.lastStatus != expected {
		return fmt.Errorf("expected status %d, got %d body=%s", expected, w.lastStatus, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theResponseSuccessFlagShouldBeTrue() error {
	ok, err := w.successFlag()
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("expected success=true, body=%s", string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) aPublishedDailyMenuExistsWithASimpleItemPriced(price float64) error {
	return w.seedPublishedDailySimpleMenu(price)
}

func (w *apiWorld) iCreateAnOrderForThePublishedMenuForCustomerPhoneExpectedInHours(customer, phone string, hours int) error {
	return w.createOrderForPublishedMenu(customer, phone, hours)
}

func (w *apiWorld) iAddALineForTheMenuItemWithQuantity(quantity int) error {
	return w.addOrderLine(quantity)
}

func (w *apiWorld) iGetTheCreatedOrder() error {
	return w.getCreatedOrder()
}

func (w *apiWorld) iUpdateThePublishedSizeOptionPriceTo(price float64) error {
	return w.updatePublishedSizeOptionPrice(price)
}

func (w *apiWorld) iRecordACashPaymentOfOnTheOrder(amount float64) error {
	return w.recordCashPayment(amount)
}

func (w *apiWorld) iAcceptTheOrder() error { return w.orderAction("accept") }

func (w *apiWorld) iStartPreparingTheOrder() error { return w.orderAction("start-preparing") }

func (w *apiWorld) iMarkTheOrderReady() error { return w.orderAction("ready") }

func (w *apiWorld) iMarkTheOrderAsPICKEDUP() error {
	return w.markOrderPickedUp()
}

func (w *apiWorld) iRefuseTheOrderWithReason(reason string) error {
	return w.refuseOrder(reason)
}

func (w *apiWorld) theCookAdminPhoneIsCleared() error {
	if err := w.setCookAdminPhone(""); err != nil {
		return err
	}
	if w.lastStatus != http.StatusOK {
		return fmt.Errorf("clear cook admin phone expected 200, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theCookAdminPhoneIs(phone string) error {
	if err := w.setCookAdminPhone(phone); err != nil {
		return err
	}
	if w.lastStatus != http.StatusOK {
		return fmt.Errorf("set cook admin phone expected 200, got %d body=%s", w.lastStatus, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) iSetTheCookAdminPhoneTo(phone string) error {
	return w.setCookAdminPhone(phone)
}

func (w *apiWorld) iClearTheCookAdminPhone() error {
	return w.setCookAdminPhone("")
}

func (w *apiWorld) theCookAdminPhoneShouldBe(phone string) error {
	got, ok, err := w.getCookAdminPhone()
	if err != nil {
		return err
	}
	if !ok || got != phone {
		return fmt.Errorf("expected cook_admin_phone %q, got present=%v value=%q body=%s", phone, ok, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theCookAdminPhoneShouldBeEmpty() error {
	got, ok, err := w.getCookAdminPhone()
	if err != nil {
		return err
	}
	if ok && got != "" {
		return fmt.Errorf("expected cook_admin_phone empty/null, got %q body=%s", got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) iListNotificationsForTheCreatedOrder() error {
	return w.listOrderNotifications(w.lastOrderID)
}

func (w *apiWorld) iListNotificationsForOrder(orderID string) error {
	return w.listOrderNotifications(orderID)
}

func (w *apiWorld) theOrderShouldHaveNNotifications(n int) error {
	rows, err := w.notificationRowsFromLastBody()
	if err != nil {
		return err
	}
	if len(rows) != n {
		return fmt.Errorf("expected %d notifications, got %d body=%s", n, len(rows), string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theOrderShouldHaveANotificationForEventTo(event, phone string) error {
	row, err := w.notificationByEvent(event)
	if err != nil {
		return err
	}
	got, _ := row["recipient_phone"].(string)
	if got != phone {
		return fmt.Errorf("event %q recipient=%q want %q body=%s", event, got, phone, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theOrderShouldNotHaveANotificationForEvent(event string) error {
	rows, err := w.notificationRowsFromLastBody()
	if err != nil {
		return err
	}
	for _, r := range rows {
		if r["event_type"] == event {
			return fmt.Errorf("unexpected notification for event %q body=%s", event, string(w.lastBody))
		}
	}
	return nil
}

func (w *apiWorld) theNotificationBodyShouldContain(event, fragment string) error {
	row, err := w.notificationByEvent(event)
	if err != nil {
		return err
	}
	body, _ := row["body"].(string)
	if !strings.Contains(body, fragment) {
		return fmt.Errorf("event %q body %q does not contain %q", event, body, fragment)
	}
	return nil
}

func (w *apiWorld) iRememberTheNotificationCountForTheCreatedOrder() error {
	if err := w.listOrderNotifications(w.lastOrderID); err != nil {
		return err
	}
	rows, err := w.notificationRowsFromLastBody()
	if err != nil {
		return err
	}
	w.rememberedNotificationCount = len(rows)
	return nil
}

func (w *apiWorld) theOrderNotificationCountShouldBeUnchanged() error {
	rows, err := w.notificationRowsFromLastBody()
	if err != nil {
		return err
	}
	if len(rows) != w.rememberedNotificationCount {
		return fmt.Errorf("expected notification count %d, got %d body=%s", w.rememberedNotificationCount, len(rows), string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) iWaitUntilTheNotificationStatusIs(event, status string) error {
	return w.waitForNotificationStatus(event, status)
}

func (w *apiWorld) theOrderStatusShouldBe(status string) error {
	got, err := w.orderStringField("status")
	if err != nil {
		return err
	}
	if got != status {
		return fmt.Errorf("expected order status %q, got %q body=%s", status, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theOrderChargedTotalShouldBe(expected float64) error {
	got, err := w.orderFloatField("charged_total")
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("expected charged_total %v, got %v body=%s", expected, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theOrderPaidAmountShouldBe(expected float64) error {
	got, err := w.orderFloatField("paid_amount")
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("expected paid_amount %v, got %v body=%s", expected, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theOrderBalanceShouldBe(expected float64) error {
	got, err := w.orderFloatField("balance")
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("expected balance %v, got %v body=%s", expected, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theOrderPaymentReceivedShouldBe(expected string) error {
	want := expected == "true"
	got, err := w.orderBoolField("payment_received")
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("expected payment_received=%v, got %v body=%s", want, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theOrderLineUnitPriceShouldBe(expected float64) error {
	got, err := w.orderFloatField("unit_price")
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("expected line unit_price %v, got %v body=%s", expected, got, string(w.lastBody))
	}
	return nil
}

func (w *apiWorld) theOrderLineExtendedAmountShouldBe(expected float64) error {
	got, err := w.orderFloatField("extended_amount")
	if err != nil {
		return err
	}
	if got != expected {
		return fmt.Errorf("expected line extended_amount %v, got %v body=%s", expected, got, string(w.lastBody))
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext, w *apiWorld) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		w.resetScenario()
		return ctx, nil
	})

	ctx.Step(`^the API is ready$`, w.theAPIIsReady)
	ctx.Step(`^the tenant header is "([^"]*)"$`, w.theTenantHeaderIs)
	ctx.Step(`^a food category exists via the API with name "([^"]*)"$`, w.aFoodCategoryExistsViaTheAPIWithName)
	ctx.Step(`^a food item exists via the API named "([^"]*)"$`, w.aFoodItemExistsViaTheAPINamed)
	ctx.Step(`^another food item exists via the API named "([^"]*)"$`, w.anotherFoodItemExistsViaTheAPINamed)
	ctx.Step(`^I create a draft daily menu named "([^"]*)" with category "([^"]*)"$`, w.iCreateADraftDailyMenuNamedWithCategory)
	ctx.Step(`^I add a combo menu item named "([^"]*)" priced (\d+(?:\.\d+)?) and (\d+(?:\.\d+)?)$`, w.iAddAComboMenuItemNamedPricedAnd)
	ctx.Step(`^I add a simple menu item named "([^"]*)" priced (\d+(?:\.\d+)?)$`, w.iAddASimpleMenuItemNamedPriced)
	ctx.Step(`^I publish the created menu$`, w.iPublishTheCreatedMenu)
	ctx.Step(`^I unpublish the created menu$`, w.iUnpublishTheCreatedMenu)
	ctx.Step(`^I get the created menu tree$`, w.iGetTheCreatedMenuTree)
	ctx.Step(`^the menu status should be "([^"]*)"$`, w.theMenuStatusShouldBe)
	ctx.Step(`^the first menu item default_total should be (\d+(?:\.\d+)?)$`, w.theFirstMenuItemDefaultTotalShouldBe)
	ctx.Step(`^I try to add a simple item to the published menu$`, w.iTryToAddASimpleItemToThePublishedMenu)
	ctx.Step(`^the created menu should not appear in the list$`, w.theCreatedMenuShouldNotAppearInTheList)
	ctx.Step(`^the created menu should appear in the list$`, w.theCreatedMenuShouldAppearInTheList)
	ctx.Step(`^the size unit list should include code "([^"]*)"$`, w.theSizeUnitListShouldIncludeCode)
	ctx.Step(`^I create a custom size unit with display name "([^"]*)"$`, w.iCreateACustomSizeUnitWithDisplayName)
	ctx.Step(`^the created size unit is_system should be false$`, w.theCreatedSizeUnitIsSystemShouldBeFalse)
	ctx.Step(`^I create a food category named "([^"]*)" with description "([^"]*)"$`, w.iCreateAFoodCategoryNamedWithDescription)
	ctx.Step(`^I create a food item named "([^"]*)"$`, w.iCreateAFoodItemNamed)
	ctx.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)" with JSON:$`, w.iSendARequestWithJSON)
	ctx.Step(`^I send a GET request to "([^"]*)"$`, w.iSendAGETRequestTo)
	ctx.Step(`^I send a PUT request to the created category with JSON:$`, w.iSendAPUTRequestToTheCreatedCategoryWithJSON)
	ctx.Step(`^I update the created category description to "([^"]*)"$`, w.iUpdateTheCreatedCategoryDescriptionTo)
	ctx.Step(`^I send a DELETE request to the created category$`, w.iSendADELETERequestToTheCreatedCategory)
	ctx.Step(`^I send a PUT request to the created food item with JSON:$`, w.iSendAPUTRequestToTheCreatedFoodItemWithJSON)
	ctx.Step(`^I send a DELETE request to the created food item$`, w.iSendADELETERequestToTheCreatedFoodItem)
	ctx.Step(`^the response status code should be (\d+)$`, w.theResponseStatusCodeShouldBe)
	ctx.Step(`^the response success flag should be true$`, w.theResponseSuccessFlagShouldBeTrue)

	// Order intake acceptance (REQORDER*, REQOLINE*, REQPAY*, REQITEM006S03)
	ctx.Step(`^a published daily menu exists with a simple item priced (\d+(?:\.\d+)?)$`, w.aPublishedDailyMenuExistsWithASimpleItemPriced)
	ctx.Step(`^I create an order for the published menu for customer "([^"]*)" phone "([^"]*)" expected in (\d+) hours$`, w.iCreateAnOrderForThePublishedMenuForCustomerPhoneExpectedInHours)
	ctx.Step(`^I add a line for the menu item with quantity (\d+)$`, w.iAddALineForTheMenuItemWithQuantity)
	ctx.Step(`^I get the created order$`, w.iGetTheCreatedOrder)
	ctx.Step(`^I update the published size option price to (\d+(?:\.\d+)?)$`, w.iUpdateThePublishedSizeOptionPriceTo)
	ctx.Step(`^I record a cash payment of (\d+(?:\.\d+)?) on the order$`, w.iRecordACashPaymentOfOnTheOrder)
	ctx.Step(`^I accept the order$`, w.iAcceptTheOrder)
	ctx.Step(`^I start preparing the order$`, w.iStartPreparingTheOrder)
	ctx.Step(`^I mark the order as READY$`, w.iMarkTheOrderReady)
	ctx.Step(`^I mark the order as PICKEDUP$`, w.iMarkTheOrderAsPICKEDUP)
	ctx.Step(`^the order status should be "([^"]*)"$`, w.theOrderStatusShouldBe)
	ctx.Step(`^the order charged_total should be (\d+(?:\.\d+)?)$`, w.theOrderChargedTotalShouldBe)
	ctx.Step(`^the order paid_amount should be (\d+(?:\.\d+)?)$`, w.theOrderPaidAmountShouldBe)
	ctx.Step(`^the order balance should be (\d+(?:\.\d+)?)$`, w.theOrderBalanceShouldBe)
	ctx.Step(`^the order payment_received should be (true|false)$`, w.theOrderPaymentReceivedShouldBe)
	ctx.Step(`^the order line unit_price should be (\d+(?:\.\d+)?)$`, w.theOrderLineUnitPriceShouldBe)
	ctx.Step(`^the order line extended_amount should be (\d+(?:\.\d+)?)$`, w.theOrderLineExtendedAmountShouldBe)

	// Order notifications acceptance (REQNOTIF001–REQNOTIF005)
	ctx.Step(`^the cook admin phone is cleared$`, w.theCookAdminPhoneIsCleared)
	ctx.Step(`^the cook admin phone is "([^"]*)"$`, w.theCookAdminPhoneIs)
	ctx.Step(`^I set the cook admin phone to "([^"]*)"$`, w.iSetTheCookAdminPhoneTo)
	ctx.Step(`^I clear the cook admin phone$`, w.iClearTheCookAdminPhone)
	ctx.Step(`^the cook admin phone should be "([^"]*)"$`, w.theCookAdminPhoneShouldBe)
	ctx.Step(`^the cook admin phone should be empty$`, w.theCookAdminPhoneShouldBeEmpty)
	ctx.Step(`^I refuse the order with reason "([^"]*)"$`, w.iRefuseTheOrderWithReason)
	ctx.Step(`^I list notifications for the created order$`, w.iListNotificationsForTheCreatedOrder)
	ctx.Step(`^I list notifications for order "([^"]*)"$`, w.iListNotificationsForOrder)
	ctx.Step(`^the order should have (\d+) notifications?$`, w.theOrderShouldHaveNNotifications)
	ctx.Step(`^the order should have a notification for event "([^"]*)" to "([^"]*)"$`, w.theOrderShouldHaveANotificationForEventTo)
	ctx.Step(`^the order should not have a notification for event "([^"]*)"$`, w.theOrderShouldNotHaveANotificationForEvent)
	ctx.Step(`^the "([^"]*)" notification body should contain "([^"]*)"$`, w.theNotificationBodyShouldContain)
	ctx.Step(`^I remember the notification count for the created order$`, w.iRememberTheNotificationCountForTheCreatedOrder)
	ctx.Step(`^the order notification count should be unchanged$`, w.theOrderNotificationCountShouldBeUnchanged)
	ctx.Step(`^I wait until the "([^"]*)" notification status is "([^"]*)"$`, w.iWaitUntilTheNotificationStatusIs)
}
