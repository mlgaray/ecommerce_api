Feature: Get Category by ID
  As a shop owner
  I want to retrieve a specific category by its ID
  So that I can view category details in my admin panel

  Background:
    Given the user is authenticated for category operations

  Scenario: Successfully get a category by ID
    Given a category with ID 1 exists
    When I send a get category by ID request for category 1
    Then the response status should be 200
    And the response should contain the category details
    And the response should contain the category image

  Scenario: Successfully get a category without image
    Given a category with ID 2 exists without image
    When I send a get category by ID request for category 2
    Then the response status should be 200
    And the response should contain the category details
    And the response should not contain a category image

  Scenario: Get non-existent category returns 404
    Given a category with ID 999 does not exist
    When I send a get category by ID request for category 999
    Then the response status should be 404
    And the user should receive an error message "category_not_found"

  Scenario: Get category with invalid ID format returns 400
    When I send a get category by ID request with invalid ID "abc"
    Then the response status should be 400
    And the user should receive an error message "invalid_category_id_format"

  Scenario: Get category without authentication returns 401
    Given the user is not authenticated
    When I send an unauthenticated get category by ID request for category 1
    Then the response status should be 401
