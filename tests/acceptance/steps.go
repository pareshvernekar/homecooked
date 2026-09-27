//go:build acceptance

package acceptance

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cucumber/godog"
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
	}
	return nil
}

func (w *apiWorld) iCreateAFoodItemNamedWithPrice(name string, price float64) error {
	categoryName := w.lastUniqueName
	if categoryName == "" {
		if err := w.createCategoryViaAPI("vegetarian"); err != nil {
			return err
		}
		categoryName = w.lastUniqueName
	}
	unique := w.uniqueName(name)
	body := fmt.Sprintf(
		`{"name":%q,"description":%q,"price":%v,"category_name":%q,"availability_status":"available"}`,
		unique, name+" description", price, categoryName,
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

func InitializeScenario(ctx *godog.ScenarioContext, w *apiWorld) {
	ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		w.resetScenario()
		return ctx, nil
	})

	ctx.Step(`^the API is ready$`, w.theAPIIsReady)
	ctx.Step(`^the tenant header is "([^"]*)"$`, w.theTenantHeaderIs)
	ctx.Step(`^a food category exists via the API with name "([^"]*)"$`, w.aFoodCategoryExistsViaTheAPIWithName)
	ctx.Step(`^a food item exists via the API named "([^"]*)"$`, w.aFoodItemExistsViaTheAPINamed)
	ctx.Step(`^I create a food category named "([^"]*)" with description "([^"]*)"$`, w.iCreateAFoodCategoryNamedWithDescription)
	ctx.Step(`^I create a food item named "([^"]*)" with price (\d+(?:\.\d+)?)$`, w.iCreateAFoodItemNamedWithPrice)
	ctx.Step(`^I send a (GET|POST|PUT|DELETE) request to "([^"]*)" with JSON:$`, w.iSendARequestWithJSON)
	ctx.Step(`^I send a GET request to "([^"]*)"$`, w.iSendAGETRequestTo)
	ctx.Step(`^I send a PUT request to the created category with JSON:$`, w.iSendAPUTRequestToTheCreatedCategoryWithJSON)
	ctx.Step(`^I update the created category description to "([^"]*)"$`, w.iUpdateTheCreatedCategoryDescriptionTo)
	ctx.Step(`^I send a DELETE request to the created category$`, w.iSendADELETERequestToTheCreatedCategory)
	ctx.Step(`^I send a PUT request to the created food item with JSON:$`, w.iSendAPUTRequestToTheCreatedFoodItemWithJSON)
	ctx.Step(`^I send a DELETE request to the created food item$`, w.iSendADELETERequestToTheCreatedFoodItem)
	ctx.Step(`^the response status code should be (\d+)$`, w.theResponseStatusCodeShouldBe)
	ctx.Step(`^the response success flag should be true$`, w.theResponseSuccessFlagShouldBeTrue)
}
