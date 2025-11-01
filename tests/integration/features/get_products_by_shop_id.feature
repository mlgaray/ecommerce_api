Feature: Get Products by Shop ID
  As a shop owner
  I want to list products by shop
  So that I can see all my shop products with pagination

  Scenario: Successfully get products with default pagination
    Given a shop with ID 1 has products
    When I send a get products request for shop 1
    Then the response status should be 200
    And the response should contain a list of products
    And the response should contain pagination metadata

  Scenario: Get products with custom limit
    Given a shop with ID 1 has products
    When I send a get products request for shop 1 with limit 5
    Then the response status should be 200
    And the response should contain at most 5 products

  Scenario: Get products with cursor pagination
    Given a shop with ID 1 has products
    When I send a get products request for shop 1 with cursor 10
    Then the response status should be 200
    And the response should contain products after cursor 10

  Scenario: Get products for shop with no products
    Given a shop with ID 999 has no products
    When I send a get products request for shop 999
    Then the response status should be 200
    And the response should contain an empty list of products
    And the response should have hasMore as false

  Scenario: Get products with negative limit
    When I send a get products request for shop 1 with limit -1
    Then the response status should be 400
    And the user should receive an error message "invalid limit format"

  Scenario: Get products with negative cursor
    When I send a get products request for shop 1 with cursor -1
    Then the response status should be 400
    And the user should receive an error message "invalid cursor format"
