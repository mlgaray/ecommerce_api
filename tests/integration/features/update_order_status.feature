Feature: Update order status
  As a shop owner
  I want to update the status of an order
  So that I can manage the order lifecycle

  Background:
    Given the user is authenticated for order operations

  Scenario: Successfully transition from pending to confirmed
    Given an order with ID 1 in shop 1 has status "pending"
    When I send an update order status request for shop 1 and order 1 with status "confirmed"
    Then the response status should be 200
    And the response should contain the order with status "confirmed"

  Scenario: Successfully transition from pending to cancelled
    Given an order with ID 1 in shop 1 has status "pending"
    When I send an update order status request for shop 1 and order 1 with status "cancelled"
    Then the response status should be 200
    And the response should contain the order with status "cancelled"

  Scenario: Successfully transition from confirmed to completed
    Given an order with ID 1 in shop 1 has status "confirmed"
    When I send an update order status request for shop 1 and order 1 with status "completed"
    Then the response status should be 200
    And the response should contain the order with status "completed"

  Scenario: Invalid transition from pending to completed
    Given an order with ID 1 in shop 1 has status "pending"
    When I send an update order status request for shop 1 and order 1 with status "completed"
    Then the response status should be 400
    And the user should receive an error message "invalid_transition"

  Scenario: Order not found
    Given an order with ID 999 does not exist in shop 1 for status update
    When I send an update order status request for shop 1 and order 999 with status "confirmed"
    Then the response status should be 404
    And the user should receive an error message "order_not_found"

  Scenario: Access denied to shop not owned by user
    When I send an update order status request for shop 999 and order 1 with status "confirmed"
    Then the response status should be 403
    And the user should receive an error message "access_denied_to_shop"

  Scenario: Concurrent modification detected
    Given an order with ID 1 in shop 1 has status "pending" but will be concurrently modified
    When I send an update order status request for shop 1 and order 1 with status "confirmed"
    Then the response status should be 409
    And the user should receive an error message "concurrent_modification"

  Scenario: Invalid order status value
    When I send an update order status request for shop 1 and order 1 with status "xyz"
    Then the response status should be 400
    And the user should receive an error message "invalid_order_status"
