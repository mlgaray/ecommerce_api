package steps

import (
	"fmt"

	"github.com/cucumber/godog"
)

type CommonSteps struct{}

func NewCommonSteps() *CommonSteps {
	return &CommonSteps{}
}

func (c *CommonSteps) theResponseStatusShouldBe(expectedStatus int) error {
	ctx := GetTestContext()

	if ctx.response == nil {
		return fmt.Errorf("no response received")
	}

	if ctx.response.StatusCode != expectedStatus {
		return fmt.Errorf("expected status code %d, got %d", expectedStatus, ctx.response.StatusCode)
	}
	return nil
}

func (c *CommonSteps) iShouldReceiveAnErrorMessage(expectedMessage string) error {
	ctx := GetTestContext()

	if ctx.errorMessage == "" {
		return fmt.Errorf("expected error message in response, got no error")
	}
	if ctx.errorMessage != expectedMessage {
		return fmt.Errorf("expected error message '%s', got '%s'", expectedMessage, ctx.errorMessage)
	}
	return nil
}

func (c *CommonSteps) iShouldReceiveASuccessMessage(expectedMessage string) error {
	ctx := GetTestContext()

	if ctx.successMessage == "" {
		return fmt.Errorf("expected success message in response, got no message")
	}
	if ctx.successMessage != expectedMessage {
		return fmt.Errorf("expected success message '%s', got '%s'", expectedMessage, ctx.successMessage)
	}
	return nil
}

// RegisterSteps registers all common step definitions
func (c *CommonSteps) RegisterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the response status should be (\d+)$`, c.theResponseStatusShouldBe)
	sc.Step(`^the user should receive an error message "([^"]*)"$`, c.iShouldReceiveAnErrorMessage)
	sc.Step(`^the user should receive a success message "([^"]*)"$`, c.iShouldReceiveASuccessMessage)
}
