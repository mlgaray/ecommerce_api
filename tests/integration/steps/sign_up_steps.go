package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"
	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/core/models"
	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts"
)

type SignUpSteps struct {
	// Empty - everything is now in TestContext
}

func NewSignUpSteps() *SignUpSteps {
	return &SignUpSteps{}
}

func (s *SignUpSteps) setupSQLExpectations() {
	const (
		validRegistrationScenario = "valid-registration"
		existingUserScenario      = "existing-user"
	)

	ctx := GetTestContext()
	switch ctx.scenario {
	case validRegistrationScenario:
		// 1. Begin transaction
		ctx.mockSQLMock.ExpectBegin()

		// 2. User creation (UserRepo.Create)
		ctx.mockSQLMock.ExpectQuery("INSERT INTO users \\(name, last_name, email, password, phone\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\) RETURNING id").
			WithArgs("John", "Doe", "newuser@example.com", sqlmock.AnyArg(), "+1234567890").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		// 3. Get admin role (RoleRepo.GetByName)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM roles WHERE name = \\$1").
			WithArgs("admin").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "admin", "Administrator role"))

		// 4. Assign role (UserRepo.AssignRole)
		ctx.mockSQLMock.ExpectExec("INSERT INTO user_roles \\(user_id, role_id, created_at\\) VALUES \\(\\$1, \\$2, now\\(\\)\\)").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 5. Shop creation (ShopRepo.Create)
		ctx.mockSQLMock.ExpectQuery("INSERT INTO shops \\(.+\\) VALUES \\(.+\\) RETURNING id").
			WithArgs(1, "John's Shop", sqlmock.AnyArg(), "shop@example.com", "+1234567890", sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// 6. Commit transaction
		ctx.mockSQLMock.ExpectCommit()

	case existingUserScenario:
		// Begin transaction first
		ctx.mockSQLMock.ExpectBegin()

		// User creation will fail with duplicate key error
		ctx.mockSQLMock.ExpectQuery("INSERT INTO users \\(name, last_name, email, password, phone\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5\\) RETURNING id").
			WithArgs("John", "Doe", "existing@example.com", sqlmock.AnyArg(), "+1234567890").
			WillReturnError(&pq.Error{
				Code:       "23505",
				Constraint: "users_email_key",
			}) // or a proper duplicate key error

		// Rollback transaction
		ctx.mockSQLMock.ExpectRollback()
	}
}

// Step definitions

func (s *SignUpSteps) iHaveValidUserRegistrationData() error {
	ctx := GetTestContext()
	ctx.scenario = "valid-registration"
	ctx.requestBody = contracts.SignUpRequest{
		User: models.User{
			Name:     "John",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: models.Shop{
			Name:  "John's Shop",
			Slug:  "johns-shop",
			Email: "shop@example.com",
			Phone: "+1234567890",
		},
	}
	return nil
}

func (s *SignUpSteps) iHaveRegistrationDataWithEmptyEmail() error {
	ctx := GetTestContext()
	ctx.requestBody = contracts.SignUpRequest{
		User: models.User{
			Name:     "John",
			LastName: "Doe",
			Email:    "",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: models.Shop{
			Name:  "John's Shop",
			Slug:  "johns-shop",
			Email: "shop@example.com",
			Phone: "+1234567890",
		},
	}
	return nil
}

func (s *SignUpSteps) iHaveRegistrationDataWithEmptyPassword() error {
	ctx := GetTestContext()
	ctx.requestBody = contracts.SignUpRequest{
		User: models.User{
			Name:     "John",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "",
			Phone:    "+1234567890",
		},
		Shop: models.Shop{
			Name:  "John's Shop",
			Slug:  "johns-shop",
			Email: "shop@example.com",
			Phone: "+1234567890",
		},
	}
	return nil
}

func (s *SignUpSteps) iHaveRegistrationDataWithEmptyName() error {
	ctx := GetTestContext()
	ctx.requestBody = contracts.SignUpRequest{
		User: models.User{
			Name:     "",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: models.Shop{
			Name:  "John's Shop",
			Slug:  "johns-shop",
			Email: "shop@example.com",
			Phone: "+1234567890",
		},
	}
	return nil
}

func (s *SignUpSteps) iHaveRegistrationDataWithEmptyShopSlug() error {
	ctx := GetTestContext()
	ctx.requestBody = contracts.SignUpRequest{
		User: models.User{
			Name:     "John",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: models.Shop{
			Name:  "John's Shop",
			Email: "shop@example.com",
			Phone: "+1234567890",
			Slug:  "",
		},
	}
	return nil
}

func (s *SignUpSteps) iHaveRegistrationDataWithInvalidEmailFormat() error {
	ctx := GetTestContext()
	ctx.requestBody = contracts.SignUpRequest{
		User: models.User{
			Name:     "John",
			LastName: "Doe",
			Email:    "invalid-email",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: models.Shop{
			Name:  "John's Shop",
			Slug:  "johns-shop",
			Email: "shop@example.com",
			Phone: "+1234567890",
		},
	}
	return nil
}

func (s *SignUpSteps) iHaveRegistrationDataWithExistingEmail() error {
	ctx := GetTestContext()
	ctx.scenario = "existing-user"
	ctx.requestBody = contracts.SignUpRequest{
		User: models.User{
			Name:     "John",
			LastName: "Doe",
			Email:    "existing@example.com",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: models.Shop{
			Name:  "John's Shop",
			Slug:  "johns-shop",
			Email: "shop@example.com",
			Phone: "+1234567890",
		},
	}
	return nil
}

func (s *SignUpSteps) iHaveRegistrationDataWithWeakPassword() error {
	ctx := GetTestContext()
	ctx.requestBody = contracts.SignUpRequest{
		User: models.User{
			Name:     "John",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "123",
			Phone:    "+1234567890",
		},
		Shop: models.Shop{
			Name:  "John's Shop",
			Slug:  "johns-shop",
			Email: "shop@example.com",
			Phone: "+1234567890",
		},
	}
	return nil
}

func (s *SignUpSteps) iSendASignUpRequest() error {
	ctx := GetTestContext()
	// Setup test app if not already done
	if ctx.app == nil {
		if err := ctx.SetupTestApp(); err != nil {
			return err
		}
	}

	// Setup SQL expectations based on scenario
	s.setupSQLExpectations()

	// Convert request body to JSON
	jsonBody, err := json.Marshal(ctx.requestBody)
	if err != nil {
		return err
	}

	// Make actual HTTP request to test server
	url := ctx.server.URL + "/auth/signup"
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
			// Parse success response (simple message)
			var responseBody map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&responseBody); err == nil {
				ctx.successMessage = responseBody["message"]
			}
		}
	}

	return nil
}

func (s *SignUpSteps) iShouldReceiveASuccessMessage() error {
	ctx := GetTestContext()
	if ctx.successMessage == "" {
		return fmt.Errorf("expected success message in response, got empty message")
	}
	return nil
}

// GetResponse returns the current response (public method for common steps)
func (s *SignUpSteps) GetResponse() *http.Response {
	ctx := GetTestContext()
	return ctx.response
}

// GetErrorMessage returns the current error message (public method for common steps)
func (s *SignUpSteps) GetErrorMessage() string {
	ctx := GetTestContext()
	return ctx.errorMessage
}

// RegisterSteps registers all step definitions
func (s *SignUpSteps) RegisterSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the user has valid registration data$`, s.iHaveValidUserRegistrationData)
	sc.Step(`^the user has registration data with empty email$`, s.iHaveRegistrationDataWithEmptyEmail)
	sc.Step(`^the user has registration data with empty password$`, s.iHaveRegistrationDataWithEmptyPassword)
	sc.Step(`^the user has registration data with empty name$`, s.iHaveRegistrationDataWithEmptyName)
	sc.Step(`^the user has registration data with empty shop slug`, s.iHaveRegistrationDataWithEmptyShopSlug)
	sc.Step(`^the user has registration data with invalid email format$`, s.iHaveRegistrationDataWithInvalidEmailFormat)
	sc.Step(`^the user has registration data with existing email$`, s.iHaveRegistrationDataWithExistingEmail)
	sc.Step(`^the user has registration data with weak password$`, s.iHaveRegistrationDataWithWeakPassword)
	sc.Step(`^the user sends a sign up request$`, s.iSendASignUpRequest)
	sc.Step(`^the user should receive a success message$`, s.iShouldReceiveASuccessMessage)
}
