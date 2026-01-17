Feature: Get Shop by ID
  As a shop owner
  I want to retrieve my shop details by ID
  So that I can view and manage my shop settings

  Background:
    Given the user is authenticated for shop operations

  Scenario: Successfully get a shop by ID
    Given a shop with ID 1 exists
    When I send a get shop by ID request for shop 1
    Then the response status should be 200
    And the response should contain the shop details
    And the response should contain the shop images
    And the response should contain the shop address
    And the response should contain the shop payment methods
    And the response should contain the shop delivery methods

  Scenario: Successfully get a shop without images
    Given a shop with ID 2 exists without images
    When I send a get shop by ID request for shop 2
    Then the response status should be 200
    And the response should contain the shop details
    And the response should not contain shop images

  Scenario: Get non-existent shop returns 404
    Given a shop with ID 888 does not exist
    When I send a get shop by ID request for shop 888
    Then the response status should be 404
    And the user should receive an error message "shop_not_found"

  Scenario: Get shop with invalid ID format returns 400
    When I send a get shop by ID request with invalid ID "abc"
    Then the response status should be 400
    And the user should receive an error message "invalid_shop_id_format"

  Scenario: Get shop without authentication returns 401
    Given the user is not authenticated
    When I send an unauthenticated get shop by ID request for shop 1
    Then the response status should be 401

  Scenario: Get shop that user does not own returns 403
    Given a shop with ID 999 exists but user does not own it
    When I send a get shop by ID request for shop 999
    Then the response status should be 403
    And the user should receive an error message "access_denied_to_shop"
