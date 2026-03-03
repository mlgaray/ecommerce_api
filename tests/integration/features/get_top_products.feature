Feature: Get top products
  As a shop owner
  I want to view my shop's top selling products
  So that I can optimize my inventory

  Background:
    Given the user is authenticated for metrics operations

  Scenario: Successfully retrieve top products
    Given the shop 1 has top products data
    When I send a GET request to "/shops/1/metrics/products/top"
    Then the response status should be 200
    And the response should contain top products data

  Scenario: Successfully retrieve with filters
    Given the shop 1 has top products data
    When I send a GET request to "/shops/1/metrics/products/top?sort_by=revenue&limit=5"
    Then the response status should be 200

  Scenario: Invalid date format returns 400
    When I send a GET request to "/shops/1/metrics/products/top?date_from=bad-date"
    Then the response status should be 400

  Scenario: Access denied to shop not owned by user
    When I send a GET request to "/shops/999/metrics/products/top"
    Then the response status should be 403
