Feature: Get top customers
  As a shop owner
  I want to view my shop's top customers
  So that I can reward loyal buyers

  Background:
    Given the user is authenticated for metrics operations

  Scenario: Successfully retrieve top customers
    Given the shop 1 has top customers data
    When I send a GET request to "/shops/1/metrics/customers/top"
    Then the response status should be 200
    And the response should contain top customers data

  Scenario: Successfully retrieve with limit filter
    Given the shop 1 has top customers data
    When I send a GET request to "/shops/1/metrics/customers/top?limit=5"
    Then the response status should be 200

  Scenario: Invalid limit format returns 400
    When I send a GET request to "/shops/1/metrics/customers/top?limit=abc"
    Then the response status should be 400

  Scenario: Access denied to shop not owned by user
    When I send a GET request to "/shops/999/metrics/customers/top"
    Then the response status should be 403
