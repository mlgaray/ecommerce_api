Feature: Get Store by Slug
  As a customer
  I want to retrieve a store by its slug
  So that I can view the store's public information

  Scenario: Successfully get a store by slug
    Given a store with slug "test-shop" exists
    When I send a get store by slug request for "test-shop"
    Then the response status should be 200
    And the response should contain the store details
    And the response should contain the store images
    And the response should contain the store address
    And the response should contain the store payment methods
    And the response should contain the store delivery methods

  Scenario: Successfully get a store without images
    Given a store with slug "shop-no-images" exists without images
    When I send a get store by slug request for "shop-no-images"
    Then the response status should be 200
    And the response should contain the store details
    And the response should not contain store images

  Scenario: Get non-existent store returns 404
    Given a store with slug "non-existent" does not exist
    When I send a get store by slug request for "non-existent"
    Then the response status should be 404
    And the user should receive an error message "store_not_found"

  Scenario: Get store with empty slug returns 404
    When I send a get store by slug request for ""
    Then the response status should be 404

  Scenario: Get store does not require authentication
    Given a store with slug "public-shop" exists
    When I send an unauthenticated get store by slug request for "public-shop"
    Then the response status should be 200
    And the response should contain the store details
