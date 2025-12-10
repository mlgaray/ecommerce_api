Feature: Delete Category
  As a shop owner
  I want to delete categories
  So that I can remove categories that are no longer needed

  Background:
    Given the user is authenticated for category delete operations

  # Success Scenarios
  Scenario: Successfully delete a category without products
    Given I have a category with id 1 that has no products
    When I send a delete category request
    Then the response status should be 204

  # Repository/Domain Errors
  Scenario: Delete category that has products returns 409
    Given I have a category with id 1 that has products assigned
    When I send a delete category request
    Then the response status should be 409
    And the user should receive an error message "category_has_products"

  Scenario: Delete non-existent category returns 404
    Given I have a category with id 999 that does not exist for deletion
    When I send a delete category request
    Then the response status should be 404
    And the user should receive an error message "category_not_found"

  # HTTP Validations
  Scenario: Delete category with invalid category_id format
    Given I have an invalid category_id "abc" for deletion
    When I send a delete category request with invalid id
    Then the response status should be 400
    And the user should receive an error message "invalid_category_id_format"

  # Authentication
  Scenario: Delete category without authentication returns 401
    Given the user is not authenticated for category deletion
    When I send an unauthenticated delete category request for category 1
    Then the response status should be 401
