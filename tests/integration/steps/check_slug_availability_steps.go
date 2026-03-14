package steps

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"
)

const (
	scenarioSlugAvailable = "slug-available"
	scenarioSlugTaken     = "slug-taken"
)

type CheckSlugAvailabilitySteps struct{}

func NewCheckSlugAvailabilitySteps() *CheckSlugAvailabilitySteps {
	return &CheckSlugAvailabilitySteps{}
}

// ===== Given Steps =====

func (s *CheckSlugAvailabilitySteps) theSlugDoesNotExistInTheDatabase(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioSlugAvailable
	return nil
}

func (s *CheckSlugAvailabilitySteps) theSlugAlreadyExistsInTheDatabase(slug string) error {
	ctx := GetTestContext()
	ctx.scenario = scenarioSlugTaken
	return nil
}

// ===== When Steps =====

func (s *CheckSlugAvailabilitySteps) iSendACheckSlugAvailabilityRequestFor(slug string) error {
	ctx := GetTestContext()

	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupStoreTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	s.setupSQLExpectations(slug)

	// Build URL and make request (public endpoint - no auth required)
	url := fmt.Sprintf("%s/stores/slugs/%s/check-availability", ctx.server.URL, slug)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	ctx.response = resp
	s.parseResponse(ctx, resp)

	return nil
}

// ===== SQL Mock Setup =====

func (s *CheckSlugAvailabilitySteps) setupSQLExpectations(slug string) {
	ctx := GetTestContext()

	switch ctx.scenario {
	case scenarioSlugAvailable:
		ctx.mockSQLMock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(slug).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	case scenarioSlugTaken:
		ctx.mockSQLMock.ExpectQuery(`SELECT EXISTS`).
			WithArgs(slug).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	}
}

func (s *CheckSlugAvailabilitySteps) parseResponse(ctx *TestContext, resp *http.Response) {
	if resp.Body == nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errorResponse map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			ctx.errorMessage = errorResponse["error"]
		}
	} else {
		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
			ctx.responseBody = response
		}
	}
}

// ===== Then Steps =====

func (s *CheckSlugAvailabilitySteps) theSlugShouldBeAvailable() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map response, got: %T", ctx.responseBody)
	}
	available, ok := response["available"].(bool)
	if !ok {
		return fmt.Errorf("expected 'available' bool field in response")
	}
	if !available {
		return fmt.Errorf("expected slug to be available, got available=false")
	}
	return nil
}

func (s *CheckSlugAvailabilitySteps) theSlugShouldNotBeAvailable() error {
	ctx := GetTestContext()
	response, ok := ctx.responseBody.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map response, got: %T", ctx.responseBody)
	}
	available, ok := response["available"].(bool)
	if !ok {
		return fmt.Errorf("expected 'available' bool field in response")
	}
	if available {
		return fmt.Errorf("expected slug to not be available, got available=true")
	}
	return nil
}

// ===== Register Steps =====

func (s *CheckSlugAvailabilitySteps) RegisterSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^the slug "([^"]*)" does not exist in the database$`, s.theSlugDoesNotExistInTheDatabase)
	sc.Step(`^the slug "([^"]*)" already exists in the database$`, s.theSlugAlreadyExistsInTheDatabase)

	// When steps
	sc.Step(`^I send a check slug availability request for "([^"]*)"$`, s.iSendACheckSlugAvailabilityRequestFor)

	// Then steps
	sc.Step(`^the slug should be available$`, s.theSlugShouldBeAvailable)
	sc.Step(`^the slug should not be available$`, s.theSlugShouldNotBeAvailable)
}
