Feature: Delete Product
  As a shop owner
  I want to delete products
  So that I can remove products that are no longer available

  Background:
    Given the user is authenticated for product delete operations

  # Success Scenarios
  Scenario: Successfully delete a product with images
    Given I have a product with id 1 that has images
    When I send a delete product request
    Then the response status should be 204

  Scenario: Successfully delete a product without images
    Given I have a product with id 2 that has no images
    When I send a delete product request
    Then the response status should be 204

  # Repository/Domain Errors
  Scenario: Delete non-existent product returns 404
    Given I have a product with id 999 that does not exist for deletion
    When I send a delete product request
    Then the response status should be 404
    And the user should receive an error message "product_not_found"

  Scenario: Delete product from different shop returns 404
    Given I have a product with id 1 that belongs to a different shop
    When I send a delete product request
    Then the response status should be 404
    And the user should receive an error message "product_not_found"

  # HTTP Validations
  Scenario: Delete product with invalid product_id format
    Given I have an invalid product_id "abc" for deletion
    When I send a delete product request with invalid id
    Then the response status should be 400
    And the user should receive an error message "invalid_product_id_format"

  # Authentication
  Scenario: Delete product without authentication returns 401
    Given the user is not authenticated for product deletion
    When I send an unauthenticated delete product request for product 1
    Then the response status should be 401
