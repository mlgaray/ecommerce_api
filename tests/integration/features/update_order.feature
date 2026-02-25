Feature: Update order
  As a shop owner
  I want to update an order's data
  So that I can modify customer info and items

  Background:
    Given the user is authenticated for order operations

  Scenario: Successfully update customer and items
    Given an editable order with ID 1 exists in shop 1
    When I send an update order request for shop 1 and order 1
    Then the response status should be 200
    And the response should contain the updated order

  Scenario: Successfully update with delivery zone
    Given an editable order with ID 1 exists in shop 1 with delivery zones
    When I send an update order request with delivery zone for shop 1 and order 1
    Then the response status should be 200
    And the response should contain the updated order

  Scenario: Order completed cannot be edited
    Given a completed order with ID 1 exists in shop 1
    When I send an update order request for shop 1 and order 1
    Then the response status should be 422
    And the user should receive an error message "order_cannot_be_edited"

  Scenario: Order cancelled cannot be edited
    Given a cancelled order with ID 1 exists in shop 1
    When I send an update order request for shop 1 and order 1
    Then the response status should be 422
    And the user should receive an error message "order_cannot_be_edited"

  Scenario: Order not found
    Given an order with ID 999 does not exist in shop 1 for update
    When I send an update order request for shop 1 and order 999
    Then the response status should be 404
    And the user should receive an error message "order_not_found"

  Scenario: Access denied to shop not owned by user
    When I send an update order request for shop 999 and order 1
    Then the response status should be 403
    And the user should receive an error message "access_denied_to_shop"

  Scenario: Subtotal mismatch on update
    Given an editable order with ID 1 exists in shop 1 for subtotal validation
    When I send an update order request with wrong subtotal for shop 1 and order 1
    Then the response status should be 400
    And the user should receive an error message "subtotal_does_not_match_calculated_subtotal"
