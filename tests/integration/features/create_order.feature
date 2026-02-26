Feature: Create order
  As a customer
  I want to create an order in a store
  So that I can purchase products

  Scenario: Successfully create an order with valid items
    Given a valid order request for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 201
    And the response should contain the created order
    And the created order should have status "pending"

  Scenario: Successfully create an order with delivery zone
    Given a valid order request with delivery zone for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 201
    And the created order should have status "pending"

  Scenario: Successfully create an order with customer address
    Given a valid order request with customer address for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 201
    And the created order should have status "pending"

  Scenario: Successfully create an order with selected options
    Given a valid order request with selected options for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 201
    And the created order should have status "pending"

  Scenario: Missing customer name
    Given an order request without customer name for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 400
    And the user should receive an error message "customer_name_is_required"

  Scenario: Missing payment method ID
    Given an order request without payment method for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 400
    And the user should receive an error message "payment_method_id_is_required"

  Scenario: Empty items array
    Given an order request without items for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 400
    And the user should receive an error message "order_must_have_at_least_one_item"

  Scenario: Missing delivery method ID
    Given an order request without delivery method for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 400
    And the user should receive an error message "delivery_method_id_is_required"

  Scenario: Missing delivery method name
    Given an order request without delivery method name for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 400
    And the user should receive an error message "delivery_method_name_is_required"

  Scenario: Invalid delivery zone ID
    Given an order request with invalid delivery zone for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 400
    And the user should receive an error message "delivery_zone_id_is_required"

  Scenario: Subtotal mismatch
    Given an order request with wrong subtotal for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 400
    And the user should receive an error message "subtotal_does_not_match_calculated_subtotal"

  Scenario: Total mismatch
    Given an order request with wrong total for store "test-store"
    When I send a create order request to store "test-store"
    Then the response status should be 400
    And the user should receive an error message "total_does_not_match_calculated_total"

  Scenario: Store not found
    Given a valid order request for a nonexistent store
    When I send a create order request to store "nonexistent-store"
    Then the response status should be 404
    And the user should receive an error message "store_not_found"
