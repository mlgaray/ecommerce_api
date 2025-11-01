package integration

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/mlgaray/ecommerce_api/tests/integration/steps"
)

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty", // can be "pretty", "progress", "junit", "cucumber"
	Tags:   "~@wip",  // "@wip",  F   //"~@wip",
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestFeatures(t *testing.T) {
	o := opts
	o.TestingT = t

	status := godog.TestSuite{
		Name:                "integration",
		ScenarioInitializer: InitializeScenario,
		Options:             &o,
	}.Run()

	if status == 2 {
		t.SkipNow()
	}

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}

func InitializeScenario(sc *godog.ScenarioContext) {
	// Initialize step definitions
	authSteps := steps.NewAuthSteps()
	signUpSteps := steps.NewSignUpSteps()
	productSteps := steps.NewProductSteps()
	getProductsByShopIDSteps := steps.NewGetProductsByShopIDSteps()
	commonSteps := steps.NewCommonSteps()

	// Register steps
	authSteps.RegisterSteps(sc)
	signUpSteps.RegisterSteps(sc)
	productSteps.RegisterSteps(sc)
	getProductsByShopIDSteps.RegisterSteps(sc)
	commonSteps.RegisterSteps(sc)

	// Setup hooks
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// Reset shared test context before each scenario
		testCtx := steps.GetTestContext()
		testCtx.Reset()

		// fmt.Printf("Setting up scenario: %s\n", sc.Name)
		return ctx, nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Cleanup after scenario
		// fmt.Printf("Cleaning up scenario: %s\n", sc.Name)

		// Clean up shared test context
		testCtx := steps.GetTestContext()
		if err := testCtx.TeardownTestApp(); err != nil {
			// TODO: Log error but continue
			_ = err
		}

		return ctx, nil
	})
}

func TestMain(m *testing.M) {
	flag.Parse()
	opts.Paths = []string{"features"}

	status := godog.TestSuite{
		Name:                "ecommerce_api",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	os.Exit(status)
}
