Feature: Get revenue trend
  As a shop owner
  I want to view my shop's revenue trend over time
  So that I can track sales patterns

  Background:
    Given the user is authenticated for metrics operations

  Scenario: Successfully retrieve revenue trend with default dates
    Given the shop 1 has revenue trend data
    When I send a GET request to "/shops/1/metrics/revenue/trend"
    Then the response status should be 200
    And the response should contain revenue trend data

  Scenario: Successfully retrieve revenue trend with date filters
    Given the shop 1 has revenue trend data
    When I send a GET request to "/shops/1/metrics/revenue/trend?date_from=2024-01-01&date_to=2024-01-31"
    Then the response status should be 200

  Scenario: Invalid date format returns 400
    When I send a GET request to "/shops/1/metrics/revenue/trend?date_from=invalid"
    Then the response status should be 400
    And the user should receive an error message "invalid_date_from_format"

  Scenario: Access denied to shop not owned by user
    When I send a GET request to "/shops/999/metrics/revenue/trend"
    Then the response status should be 403
