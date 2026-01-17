Feature: Get Store Products by Slug
  As a customer
  I want to retrieve products for a store by its slug
  So that I can browse the store's product catalog with infinite scroll

  Scenario: Successfully get products for a store
    Given a store with slug "test-shop" exists with products
    When I send a get store products request for "test-shop"
    Then the response status should be 200
    And the response should contain products
    And the response should include products pagination info

  Scenario: Successfully get products with search filter
    Given a store with slug "test-shop" exists with products
    When I send a get store products request for "test-shop" with search "Laptop"
    Then the response status should be 200
    And the response should contain products

  Scenario: Successfully get products with category filter
    Given a store with slug "test-shop" exists with products
    When I send a get store products request for "test-shop" with category_id 1
    Then the response status should be 200
    And the response should contain products

  Scenario: Get products for store without products returns empty list
    Given a store with slug "empty-shop" exists without products
    When I send a get store products request for "empty-shop"
    Then the response status should be 200
    And the response should contain empty products

  Scenario: Get products for non-existent store returns 404
    Given a store with slug "non-existent" does not exist
    When I send a get store products request for "non-existent"
    Then the response status should be 404
    And the user should receive an error message "store_not_found"

  Scenario: Get products does not require authentication
    Given a store with slug "public-shop" exists with products
    When I send an unauthenticated get store products request for "public-shop"
    Then the response status should be 200
    And the response should contain products

  Scenario: Get products supports pagination
    Given a store with slug "test-shop" exists with products
    When I send a get store products request for "test-shop" with limit 2
    Then the response status should be 200
    And the response should contain products
    And the response should include products pagination info

  Scenario: Get products only returns active products
    Given a store with slug "test-shop" exists with mixed active and inactive products
    When I send a get store products request for "test-shop"
    Then the response status should be 200
    And the response should contain products
    And all returned products should have is_active true

  Scenario: Get products with search and category_id combined
    Given a store with slug "test-shop" exists with products
    When I send a get store products request for "test-shop" with search "Product" and category_id 1
    Then the response status should be 200
    And the response should contain products

  Scenario: Get products with slug containing special characters is handled
    Given a store with slug "test-shop" exists with products
    When I send a get store products request for "test'; DROP TABLE shops;--"
    Then the response status should be 404
    And the user should receive an error message "store_not_found"

  Scenario: Get products with sort by price ascending
    Given a store with slug "test-shop" exists with products
    When I send a get store products request for "test-shop" with sort "price" and order "asc"
    Then the response status should be 200
    And the response should contain products
