Feature: Get Products by Shop ID with Filters
  As a shop owner
  I want to list and filter products by shop
  So that I can see all my shop products with pagination and filters

  Scenario: Successfully get products with default pagination
    Given a shop with ID 1 has products
    When I send a get products request for shop 1
    Then the response status should be 200
    And the response should contain a list of products
    And the response should contain pagination metadata
    And the response should contain total count on first page

  Scenario: Get products with custom limit
    Given a shop with ID 1 has products
    When I send a get products request for shop 1 with limit 5
    Then the response status should be 200
    And the response should contain at most 5 products

  Scenario: Get products with cursor pagination
    Given a shop with ID 1 has products with cursor pagination
    When I send a get products request for shop 1 with a valid cursor
    Then the response status should be 200
    And the response should contain products from next page
    And the response should not contain total count

  Scenario: Get products for shop with no products
    Given a shop with ID 999 has no products
    When I send a get products request for shop 999
    Then the response status should be 200
    And the response should contain an empty list of products
    And the response should have hasMore as false

  Scenario: Get products with negative limit
    When I send a get products request for shop 1 with limit -1
    Then the response status should be 400
    And the user should receive an error message "limit_cannot_be_negative"

  Scenario: Search products by name
    Given a shop with ID 1 has products with different names
    When I send a get products request for shop 1 with search term "Laptop"
    Then the response status should be 200
    And the response should contain products matching search term

  Scenario: Filter products by category
    Given a shop with ID 1 has products in different categories
    When I send a get products request for shop 1 with category ID 2
    Then the response status should be 200
    And the response should contain products from category 2

  Scenario: Filter products by price range
    Given a shop with ID 1 has products with different prices
    When I send a get products request for shop 1 with min price 100 and max price 500
    Then the response status should be 200
    And the response should contain products within price range

  Scenario: Filter products by active status
    Given a shop with ID 1 has active and inactive products
    When I send a get products request for shop 1 with is_active true
    Then the response status should be 200
    And the response should contain only active products

  Scenario: Sort products by price ascending
    Given a shop with ID 1 has products with different prices
    When I send a get products request for shop 1 sorted by price ascending
    Then the response status should be 200
    And the response products should be sorted by price ascending

  Scenario: Combine multiple filters
    Given a shop with ID 1 has products
    When I send a get products request for shop 1 with category ID 1 and search term "Product"
    Then the response status should be 200
    And the response should contain filtered products
