Feature: Get dashboard metrics
  As a shop owner
  I want to view my shop's dashboard metrics
  So that I can understand my business performance

  Background:
    Given the user is authenticated for metrics operations

  Scenario: Successfully retrieve dashboard metrics
    Given the shop 1 has metrics data
    When I send a GET request to "/shops/1/metrics/dashboard"
    Then the response status should be 200
    And the response should contain dashboard metrics

  Scenario: Successfully retrieve dashboard with timezone
    Given the shop 1 has metrics data
    When I send a GET request to "/shops/1/metrics/dashboard?tz=America/New_York"
    Then the response status should be 200

  Scenario: Invalid shop ID format
    When I send a GET request to "/shops/abc/metrics/dashboard"
    Then the response status should be 400
    And the user should receive an error message "invalid_shop_id_format"

  Scenario: Access denied to shop not owned by user
    When I send a GET request to "/shops/999/metrics/dashboard"
    Then the response status should be 403
