Feature: Get all orders by shop
  As a shop owner
  I want to retrieve all orders for my shop
  So that I can manage them from the dashboard

  Background:
    Given the user is authenticated for order operations

  Scenario: Successfully retrieve orders for a shop
    Given orders exist in shop 1
    When I send a get all orders request for shop 1
    Then the response status should be 200
    And the response should contain a list of orders

  Scenario: Empty orders list for shop
    Given no orders exist in shop 1
    When I send a get all orders request for shop 1
    Then the response status should be 200
    And the response should contain an empty list of orders

  Scenario: Access denied to shop not owned by user
    When I send a get all orders request for shop 999
    Then the response status should be 403
    And the user should receive an error message "access_denied_to_shop"

  Scenario: Invalid shop ID format
    When I send a get all orders request with invalid shop ID "abc"
    Then the response status should be 400
    And the user should receive an error message "invalid_shop_id_format"
