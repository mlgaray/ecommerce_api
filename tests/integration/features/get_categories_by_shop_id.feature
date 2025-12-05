Feature: Get Categories by Shop ID with Filters
  As a shop owner
  I want to list and filter categories by shop
  So that I can see all my shop categories with pagination and filters

  Scenario: Successfully get categories with default pagination
    Given a shop with ID 1 has categories
    When I send a get categories request for shop 1
    Then the response status should be 200
    And the response should contain a list of categories
    And the response should contain category pagination metadata
    And the response should contain category total count on first page

  Scenario: Get categories with custom limit
    Given a shop with ID 1 has categories
    When I send a get categories request for shop 1 with limit 5
    Then the response status should be 200
    And the response should contain at most 5 categories

  Scenario: Get categories with cursor pagination
    Given a shop with ID 1 has categories with cursor pagination
    When I send a get categories request for shop 1 with a valid cursor
    Then the response status should be 200
    And the response should contain categories from next page
    And the response should not contain category total count

  Scenario: Get categories for shop with no categories
    Given a shop with ID 999 has no categories
    When I send a get categories request for shop 999
    Then the response status should be 200
    And the response should contain an empty list of categories
    And the response should have category hasMore as false

  Scenario: Get categories with negative limit
    When I send a get categories request for shop 1 with limit -1
    Then the response status should be 400
    And the user should receive an error message "limit_cannot_be_negative"

  Scenario: Search categories by name
    Given a shop with ID 1 has categories with different names
    When I send a get categories request for shop 1 with search term "Electronics"
    Then the response status should be 200
    And the response should contain categories matching search term

  Scenario: Sort categories by name ascending
    Given a shop with ID 1 has categories with different names
    When I send a get categories request for shop 1 sorted by name ascending
    Then the response status should be 200
    And the response categories should be sorted by name ascending

  Scenario: Sort categories by created_at descending
    Given a shop with ID 1 has categories
    When I send a get categories request for shop 1 sorted by created_at descending
    Then the response status should be 200
    And the response categories should be sorted by created_at descending
