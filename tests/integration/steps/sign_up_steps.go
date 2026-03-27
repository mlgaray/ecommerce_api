package steps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cucumber/godog"
	"github.com/lib/pq"

	"github.com/mlgaray/ecommerce_api/internal/infraestructure/adapters/http/contracts/requests"
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
		ctx.mockSQLMock.ExpectQuery("INSERT INTO users \\(name, last_name, email, password, phone, is_active\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6\\) RETURNING id").
			WithArgs("John", "Doe", "newuser@example.com", sqlmock.AnyArg(), "+1234567890", false).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// 3. Get default timezone (ShopRepo.Create fetches timezone before INSERT)
		ctx.mockSQLMock.ExpectQuery("SELECT id FROM timezones WHERE identifier = 'America/Buenos_Aires' LIMIT 1").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// 4. Shop creation (ShopRepo.Create — no user_id)
		ctx.mockSQLMock.ExpectQuery("INSERT INTO shops \\(.+\\) VALUES \\(.+\\) RETURNING id").
			WithArgs("John's Shop", "johns-shop", "shop@example.com", "+1234567890", "", 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// 5. Get all payment methods (PaymentMethodRepo.GetAll)
		ctx.mockSQLMock.ExpectQuery("SELECT .+ FROM payment_methods WHERE is_active \\= true").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "description", "is_active"}).
				AddRow(1, "Transferencia", "transfer", "Transferencia bancaria", true).
				AddRow(2, "MercadoPago", "mercadopago", "Pago con MercadoPago", true))

		// 6. Create shop_payment_methods for each payment method
		ctx.mockSQLMock.ExpectExec("INSERT INTO shop_payment_methods").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		ctx.mockSQLMock.ExpectExec("INSERT INTO shop_payment_methods").
			WithArgs(1, 2).
			WillReturnResult(sqlmock.NewResult(2, 1))

		// 7. Get all delivery methods (DeliveryMethodRepo.GetAll)
		ctx.mockSQLMock.ExpectQuery("SELECT .+ FROM delivery_methods WHERE is_active \\= true").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code", "description", "is_active"}).
				AddRow(1, "Envío a domicilio", "delivery", "Envío a domicilio", true).
				AddRow(2, "Retiro en local", "pickup", "Retiro en local", true))

		// 8. Create shop_delivery_methods for each delivery method
		ctx.mockSQLMock.ExpectExec("INSERT INTO shop_delivery_methods").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		ctx.mockSQLMock.ExpectExec("INSERT INTO shop_delivery_methods").
			WithArgs(1, 2).
			WillReturnResult(sqlmock.NewResult(2, 1))

		// 9. Create staff entry (StaffRepo.Create — links user to shop)
		ctx.mockSQLMock.ExpectQuery("INSERT INTO staff").
			WithArgs(1, 1).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

		// 10. Get owner role (RoleRepo.GetByName)
		ctx.mockSQLMock.ExpectQuery("SELECT (.+) FROM roles WHERE name = \\$1").
			WithArgs("owner").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "owner", "Shop owner"))

		// 11. Assign staff_role (StaffRepo.AssignRole)
		ctx.mockSQLMock.ExpectExec("INSERT INTO staff_roles").
			WithArgs(1, 1).
			WillReturnResult(sqlmock.NewResult(1, 1))

		// 12. Commit transaction
		ctx.mockSQLMock.ExpectCommit()

	case existingUserScenario:
		// Begin transaction first
		ctx.mockSQLMock.ExpectBegin()

		// User creation will fail with duplicate key error
		ctx.mockSQLMock.ExpectQuery("INSERT INTO users \\(name, last_name, email, password, phone, is_active\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, \\$6\\) RETURNING id").
			WithArgs("John", "Doe", "existing@example.com", sqlmock.AnyArg(), "+1234567890", false).
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
	ctx.requestBody = requests.SignUpRequest{
		User: requests.UserRequest{
			Name:     "John",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: requests.SignUpShopRequest{
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
	ctx.requestBody = requests.SignUpRequest{
		User: requests.UserRequest{
			Name:     "John",
			LastName: "Doe",
			Email:    "",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: requests.SignUpShopRequest{
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
	ctx.requestBody = requests.SignUpRequest{
		User: requests.UserRequest{
			Name:     "John",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "",
			Phone:    "+1234567890",
		},
		Shop: requests.SignUpShopRequest{
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
	ctx.requestBody = requests.SignUpRequest{
		User: requests.UserRequest{
			Name:     "",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: requests.SignUpShopRequest{
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
	ctx.requestBody = requests.SignUpRequest{
		User: requests.UserRequest{
			Name:     "John",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: requests.SignUpShopRequest{
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
	ctx.requestBody = requests.SignUpRequest{
		User: requests.UserRequest{
			Name:     "John",
			LastName: "Doe",
			Email:    "invalid-email",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: requests.SignUpShopRequest{
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
	ctx.requestBody = requests.SignUpRequest{
		User: requests.UserRequest{
			Name:     "John",
			LastName: "Doe",
			Email:    "existing@example.com",
			Password: "SecurePassword123!",
			Phone:    "+1234567890",
		},
		Shop: requests.SignUpShopRequest{
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
	ctx.requestBody = requests.SignUpRequest{
		User: requests.UserRequest{
			Name:     "John",
			LastName: "Doe",
			Email:    "newuser@example.com",
			Password: "123",
			Phone:    "+1234567890",
		},
		Shop: requests.SignUpShopRequest{
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
