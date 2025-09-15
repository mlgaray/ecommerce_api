Feature: User Sign Up
  As a new user
  I want to create an account
  So that I can access the system
  

  Scenario: Successfully create a new user account
    Given the user has valid registration data
    When the user sends a sign up request
    And the response status should be 200


  Scenario: Sign up with empty email
    Given the user has registration data with empty email
    When the user sends a sign up request
    Then the response status should be 400
    And the user should receive an error message "user_email_is_required"


  Scenario: Sign up with empty password
    Given the user has registration data with empty password
    When the user sends a sign up request
    Then the response status should be 400
    And the user should receive an error message "user_password_is_required"

  Scenario: Sign up with empty name
    Given the user has registration data with empty name
    When the user sends a sign up request
    Then the response status should be 400
    And the user should receive an error message "user_name_is_required"

  Scenario: Sign up with invalid email format
    Given the user has registration data with invalid email format
    When the user sends a sign up request
    Then the response status should be 400
    And the user should receive an error message "invalid_email_format"

  Scenario: Sign up with existing email
    Given the user has registration data with existing email
    When the user sends a sign up request
    Then the response status should be 409
    And the user should receive an error message "user_already_exists"

  Scenario: Sign up with empty slug
    Given the user has registration data with empty shop slug
    When the user sends a sign up request
    Then the response status should be 400
    And the user should receive an error message "shop_slug_is_required"

#  Scenario: Sign up with weak password
#    Given I have registration data with weak password
#    When I send a sign up request
#    Then the response status should be 400
#    And I should receive an error message "password too weak"