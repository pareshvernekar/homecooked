//go:build acceptance

package acceptance

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

func TestAcceptanceFeatures(t *testing.T) {
	world := newAPIWorld()

	if err := world.pingServer(); err != nil {
		t.Fatalf(
			"acceptance tests require a running local server at %s\n"+
				"Start one first, e.g.:\n"+
				"  make local\n"+
				"  # or: go run ./cmd/main.go\n"+
				"Optional overrides: ACCEPTANCE_BASE_URL, ACCEPTANCE_TENANT_ID\n"+
				"Error: %v",
			world.baseURL, err,
		)
	}

	opts := godog.Options{
		Output:   colors.Colored(os.Stdout),
		Format:   "pretty",
		Paths:    []string{"features"},
		TestingT: t,
		Strict:   true,
	}

	status := godog.TestSuite{
		Name: "homecooked-acceptance",
		ScenarioInitializer: func(ctx *godog.ScenarioContext) {
			InitializeScenario(ctx, world)
		},
		Options: &opts,
	}.Run()

	if status != 0 {
		t.Fatalf("acceptance features failed with status %d", status)
	}
}
