Feature: Get order by ID
  As a shop owner
  I want to retrieve a specific order by its ID
  So that I can see its full details

  Background:
    Given the user is authenticated for order operations

  Scenario: Successfully retrieve an existing order
    Given an order with ID 1 exists in shop 1
    When I send a get order by ID request for shop 1 and order 1
    Then the response status should be 200
    And the response should contain the order details
    And the response should contain the order items

  Scenario: Order not found
    Given an order with ID 999 does not exist in shop 1
    When I send a get order by ID request for shop 1 and order 999
    Then the response status should be 404
    And the user should receive an error message "order_not_found"

  Scenario: Access denied to shop not owned by user
    When I send a get order by ID request for shop 999 and order 1
    Then the response status should be 403
    And the user should receive an error message "access_denied_to_shop"

  Scenario: Invalid shop ID format
    When I send a get order by ID request with invalid shop ID "abc" and order 1
    Then the response status should be 400
    And the user should receive an error message "invalid_shop_id_format"

  Scenario: Invalid order ID - zero
    When I send a get order by ID request for shop 1 and order 0
    Then the response status should be 400
    And the user should receive an error message "invalid_order_id_format"
