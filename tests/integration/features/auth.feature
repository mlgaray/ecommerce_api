Feature: Authentication
  As a user
  I want to sign in to the application
  So that I can access protected resources

  Scenario: Successful sign in
    Given the user has valid credentials
    When the user sends a sign in request
    Then the response status should be 200
    And the user should receive a token

  Scenario: Sign in with empty email
    Given the user has credentials with empty email
    When the user sends a sign in request
    Then the response status should be 400
    And the user should receive an error message "email_is_required"

  Scenario: Sign in with empty password
    Given the user has credentials with empty password
    When the user sends a sign in request
    Then the response status should be 400
    And the user should receive an error message "password_is_required"

  Scenario: Sign in with non-existent user
    Given the user has credentials for a non-existent user
    When the user sends a sign in request
    Then the response status should be 404
    And the user should receive an error message "user_not_found"

  Scenario: Sign in with wrong password
    Given the user has credentials with wrong password
    When the user sends a sign in request
    Then the response status should be 401
    And the user should receive an error message "invalid_credentials"
  @wip
  Scenario: Sign in with invalid email format
    Given the user has credentials with invalid email format
    When the user sends a sign in request
    Then the response status should be 400
    And the user should receive an error message "invalid_email_format"