Feature: Get Store Product by ID
  As a customer
  I want to retrieve a specific product by its ID from a store
  So that I can view the product details including variants and options

  Scenario: Successfully get product by ID
    Given a store with slug "test-shop" exists with products
    When I send a get store product by id request for "test-shop" with product_id 1
    Then the response status should be 200
    And the response should contain product details
    And the response should include product variants and options

  Scenario: Get product for non-existent store returns 404
    Given a store with slug "non-existent" does not exist
    When I send a get store product by id request for "non-existent" with product_id 1
    Then the response status should be 404
    And the user should receive an error message "store_not_found"

  Scenario: Get non-existent product returns 404
    Given a store with slug "test-shop" exists with products
    When I send a get store product by id request for "test-shop" with product_id 99999
    Then the response status should be 404
    And the user should receive an error message "product_not_found"

  Scenario: Get inactive product returns 404
    Given a store with slug "test-shop" exists with an inactive product with id 999
    When I send a get store product by id request for "test-shop" with product_id 999
    Then the response status should be 404
    And the user should receive an error message "product_not_found"

  Scenario: Get product with invalid product_id returns 400
    Given a store with slug "test-shop" exists with products
    When I send a get store product by id request for "test-shop" with invalid product_id "abc"
    Then the response status should be 400
    And the user should receive an error message "invalid_product_id"

  Scenario: Get product does not require authentication
    Given a store with slug "public-shop" exists with products
    When I send an unauthenticated get store product by id request for "public-shop" with product_id 1
    Then the response status should be 200
    And the response should contain product details

  Scenario: Get product with negative product_id returns 400
    Given a store with slug "test-shop" exists with products
    When I send a get store product by id request for "test-shop" with product_id -1
    Then the response status should be 400
    And the user should receive an error message "invalid_product_id"

  Scenario: Get product with product_id zero returns 400
    Given a store with slug "test-shop" exists with products
    When I send a get store product by id request for "test-shop" with product_id 0
    Then the response status should be 400
    And the user should receive an error message "invalid_product_id"
