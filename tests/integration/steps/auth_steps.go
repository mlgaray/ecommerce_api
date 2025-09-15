package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
)

type AuthSteps struct {
	// Empty - everything is now in TestContext
}

func NewAuthSteps() *AuthSteps {
	return &AuthSteps{}
}

func (a *AuthSteps) setupSQLExpectations() {
	const (
		validUserScenario       = "valid-user"
		nonExistentUserScenario = "non-existent-user"
		wrongPasswordScenario   = "wrong-password"
	)

	ctx := GetTestContext()
	switch ctx.scenario {
	case validUserScenario:
		// Mock successful user lookup (using direct DB query, not transaction)
		rows := sqlmock.NewRows([]string{"id", "name", "email", "phone", "password", "is_active", "role_id", "role_name"}).
			AddRow(1, "Test User", "user@example.com", "+1234567890", "password123", true, 0, "")
		ctx.mockSQLMock.ExpectQuery("SELECT\\s+u\\.id, u\\.name, u\\.email, u\\.phone, u\\.password, u\\.is_active,\\s+COALESCE\\(r\\.id, 0\\) as role_id,\\s+COALESCE\\(r\\.name, ''\\) as role_name\\s+FROM users u\\s+LEFT JOIN user_roles ur ON u\\.id = ur\\.user_id\\s+LEFT JOIN roles r ON ur\\.role_id = r\\.id\\s+WHERE u\\.email = \\$1\\s+ORDER BY u\\.id, r\\.id").
			WithArgs("user@example.com").
			WillReturnRows(rows)

	case nonExistentUserScenario:
		// Mock user not found (using direct DB query, not transaction)
		// Return empty rows instead of sql.ErrNoRows to trigger the !rows.Next() condition
		emptyRows := sqlmock.NewRows([]string{"id", "name", "email", "phone", "password", "is_active", "role_id", "role_name"})
		ctx.mockSQLMock.ExpectQuery("SELECT\\s+u\\.id, u\\.name, u\\.email, u\\.phone, u\\.password, u\\.is_active,\\s+COALESCE\\(r\\.id, 0\\) as role_id,\\s+COALESCE\\(r\\.name, ''\\) as role_name\\s+FROM users u\\s+LEFT JOIN user_roles ur ON u\\.id = ur\\.user_id\\s+LEFT JOIN roles r ON ur\\.role_id = r\\.id\\s+WHERE u\\.email = \\$1\\s+ORDER BY u\\.id, r\\.id").
			WithArgs("nonexistent@example.com").
			WillReturnRows(emptyRows)

	case wrongPasswordScenario:
		// Mock user found but with different password (using direct DB query, not transaction)
		rows := sqlmock.NewRows([]string{"id", "name", "email", "phone", "password", "is_active", "role_id", "role_name"}).
			AddRow(1, "Test User", "user@example.com", "+1234567890", "correctpassword", true, 0, "")
		ctx.mockSQLMock.ExpectQuery("SELECT\\s+u\\.id, u\\.name, u\\.email, u\\.phone, u\\.password, u\\.is_active,\\s+COALESCE\\(r\\.id, 0\\) as role_id,\\s+COALESCE\\(r\\.name, ''\\) as role_name\\s+FROM users u\\s+LEFT JOIN user_roles ur ON u\\.id = ur\\.user_id\\s+LEFT JOIN roles r ON ur\\.role_id = r\\.id\\s+WHERE u\\.email = \\$1\\s+ORDER BY u\\.id, r\\.id").
			WithArgs("user@example.com").
			WillReturnRows(rows)
	}
}

func (a *AuthSteps) iHaveValidUserCredentials() error {
	ctx := GetTestContext()
	ctx.scenario = "valid-user"
	ctx.signInRequest = contracts.SignInRequest{
		Email:    "user@example.com",
		Password: "password123",
	}

	return nil
}

func (a *AuthSteps) iHaveCredentialsWithEmptyEmail() error {
	ctx := GetTestContext()
	ctx.signInRequest = contracts.SignInRequest{
		Email:    "",
		Password: "password123",
	}
	return nil
}

func (a *AuthSteps) iHaveCredentialsWithEmptyPassword() error {
	ctx := GetTestContext()
	ctx.signInRequest = contracts.SignInRequest{
		Email:    "user@example.com",
		Password: "",
	}
	return nil
}

func (a *AuthSteps) iHaveCredentialsForANonExistentUser() error {
	ctx := GetTestContext()
	ctx.scenario = "non-existent-user"
	ctx.signInRequest = contracts.SignInRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	return nil
}

func (a *AuthSteps) iHaveCredentialsWithWrongPassword() error {
	ctx := GetTestContext()
	ctx.scenario = "wrong-password"
	ctx.signInRequest = contracts.SignInRequest{
		Email:    "user@example.com",
		Password: "wrongpassword",
	}
	return nil
}

func (a *AuthSteps) iHaveCredentialsWithInvalidEmailFormat() error {
	ctx := GetTestContext()
	ctx.signInRequest = contracts.SignInRequest{
		Email:    "invalid-email-format",
		Password: "password123",
	}
	return nil
}

func (a *AuthSteps) iSendASignInRequest() error {
	ctx := GetTestContext()
	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	a.setupSQLExpectations()

	// Convert request body to JSON
	jsonBody, err := json.Marshal(ctx.signInRequest)
	if err != nil {
		return err
	}

	// Make actual HTTP request to test server
	url := ctx.server.URL + "/auth/signin"
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	ctx.response = resp

	// Parse response body
	if resp.Body != nil {
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			// Parse error response
			var errorResponse map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
				ctx.errorMessage = errorResponse["error"]
			}
		} else {
			// Parse success response
			var responseBody contracts.SignInResponse
			if err := json.NewDecoder(resp.Body).Decode(&responseBody); err == nil {
				ctx.signInResponse = responseBody
			}
		}
	}

	return nil
}

func (a *AuthSteps) iShouldReceiveAToken() error {
	ctx := GetTestContext()
	if ctx.signInResponse.Token == "" {
		return fmt.Errorf("expected token in response, got: %v", ctx.signInResponse)
	}
	return nil
}

// RegisterSteps registers all step definitions
func (a *AuthSteps) RegisterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the user has valid credentials$`, a.iHaveValidUserCredentials)
	sc.Step(`^the user has credentials with empty email$`, a.iHaveCredentialsWithEmptyEmail)
	sc.Step(`^the user has credentials with empty password$`, a.iHaveCredentialsWithEmptyPassword)
	sc.Step(`^the user has credentials for a non-existent user$`, a.iHaveCredentialsForANonExistentUser)
	sc.Step(`^the user has credentials with wrong password$`, a.iHaveCredentialsWithWrongPassword)
	sc.Step(`^the user has credentials with invalid email format$`, a.iHaveCredentialsWithInvalidEmailFormat)
	sc.Step(`^the user sends a sign in request$`, a.iSendASignInRequest)
	sc.Step(`^the user should receive a token$`, a.iShouldReceiveAToken)
}
