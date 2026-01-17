Feature: Get Store Featured Products by Slug
  As a customer
  I want to retrieve featured products for a store by its slug
  So that I can see highlighted products on the store homepage

  Scenario: Successfully get featured products for a store
    Given a store with slug "test-shop" exists with featured products
    When I send a get store featured products request for "test-shop"
    Then the response status should be 200
    And the response should contain featured products
    And the response should include products pagination info

  Scenario: Get featured products for store without featured products returns empty list
    Given a store with slug "empty-shop" exists without featured products
    When I send a get store featured products request for "empty-shop"
    Then the response status should be 200
    And the response should contain empty products

  Scenario: Get featured products for non-existent store returns 404
    Given a store with slug "non-existent" does not exist
    When I send a get store featured products request for "non-existent"
    Then the response status should be 404
    And the user should receive an error message "store_not_found"

  Scenario: Get featured products does not require authentication
    Given a store with slug "public-shop" exists with featured products
    When I send an unauthenticated get store featured products request for "public-shop"
    Then the response status should be 200
    And the response should contain featured products

  Scenario: Featured products are filtered by is_highlighted and is_active
    Given a store with slug "test-shop" exists with mixed products
    When I send a get store featured products request for "test-shop"
    Then the response status should be 200
    And the response should only contain highlighted active products
