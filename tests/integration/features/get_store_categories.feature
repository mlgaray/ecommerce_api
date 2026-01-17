Feature: Get Store Categories by Slug
  As a customer
  I want to retrieve categories for a store by its slug
  So that I can browse the store's product categories

  Scenario: Successfully get categories for a store
    Given a store with slug "test-shop" exists with categories
    When I send a get store categories request for "test-shop"
    Then the response status should be 200
    And the response should contain categories
    And the response should include pagination info

  Scenario: Successfully get categories with search filter
    Given a store with slug "test-shop" exists with categories
    When I send a get store categories request for "test-shop" with search "Electronics"
    Then the response status should be 200
    And the response should contain categories

  Scenario: Get categories for store without categories returns empty list
    Given a store with slug "empty-shop" exists without categories
    When I send a get store categories request for "empty-shop"
    Then the response status should be 200
    And the response should contain empty categories

  Scenario: Get categories for non-existent store returns 404
    Given a store with slug "non-existent" does not exist
    When I send a get store categories request for "non-existent"
    Then the response status should be 404
    And the user should receive an error message "store_not_found"

  Scenario: Get categories does not require authentication
    Given a store with slug "public-shop" exists with categories
    When I send an unauthenticated get store categories request for "public-shop"
    Then the response status should be 200
    And the response should contain categories
