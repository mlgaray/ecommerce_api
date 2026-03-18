Feature: Check Slug Availability
  As a new user during registration
  I want to check if a store slug is available
  So that I can choose a unique slug before completing sign up

  Scenario: Successfully check an available slug
    Given the slug "brand-new-store" does not exist in the database
    When I send a check slug availability request for "brand-new-store"
    Then the response status should be 200
    And the slug should be available

  Scenario: Check slug that is already taken
    Given the slug "taken-store" already exists in the database
    When I send a check slug availability request for "taken-store"
    Then the response status should be 200
    And the slug should not be available

  Scenario: Check slug with invalid format returns 400
    When I send a check slug availability request for "INVALID!"
    Then the response status should be 400
    And the user should receive an error message "invalid_slug_format"

  Scenario: Check slug that is too short returns 400
    When I send a check slug availability request for "ab"
    Then the response status should be 400
    And the user should receive an error message "invalid_slug_format"

  Scenario: Check slug with hyphens at start returns 400
    When I send a check slug availability request for "-my-store"
    Then the response status should be 400
    And the user should receive an error message "invalid_slug_format"

  Scenario: Check slug does not require authentication
    Given the slug "public-check" does not exist in the database
    When I send a check slug availability request for "public-check"
    Then the response status should be 200
    And the slug should be available
