Feature: Update Shop
  As a shop owner
  I want to update my shop details
  So that I can keep my shop information current

  Background:
    Given the user is authenticated for shop operations

  # Successful update scenarios
  Scenario: Successfully update shop metadata without new images
    Given a shop with ID 1 exists for update
    And I have valid shop update data without new images
    When I send an update shop request for shop 1
    Then the response status should be 204

  Scenario: Successfully update shop with new logo image
    Given a shop with ID 1 exists for update
    And I have valid shop update data with a new logo image
    When I send an update shop request for shop 1
    Then the response status should be 204

  Scenario: Successfully update shop with new cover image
    Given a shop with ID 1 exists for update
    And I have valid shop update data with a new cover image
    When I send an update shop request for shop 1
    Then the response status should be 204

  Scenario: Successfully update shop with both new images
    Given a shop with ID 1 exists for update
    And I have valid shop update data with both new images
    When I send an update shop request for shop 1
    Then the response status should be 204

  # HTTP Validation errors (Infrastructure layer - Contracts)
  Scenario: Update shop without name returns 400
    Given a shop with ID 1 exists for update
    And I have shop update data without name
    When I send an update shop request for shop 1
    Then the response status should be 400
    And the user should receive an error message "shop_name_is_required"

  Scenario: Update shop without slug returns 400
    Given a shop with ID 1 exists for update
    And I have shop update data without slug
    When I send an update shop request for shop 1
    Then the response status should be 400
    And the user should receive an error message "invalid_slug_format"

  Scenario: Update shop with invalid image type returns 400
    Given a shop with ID 1 exists for update
    And I have shop update data with an invalid image type
    When I send an update shop request for shop 1
    Then the response status should be 400
    And the user should receive an error message "invalid_logo_image_type_only_jpeg_png_allowed"

  # Business Validation errors (Domain layer - Model.Validate())
  Scenario: Update shop with invalid transfer CBU returns 400
    Given a shop with ID 1 exists for update
    And I have shop update data with invalid transfer CBU
    When I send an update shop request for shop 1
    Then the response status should be 400
    And the user should receive an error message "transfer_cbu_invalid_length"

  # Authorization errors
  Scenario: Update shop without authentication returns 401
    Given the user is not authenticated
    When I send an unauthenticated update shop request for shop 1
    Then the response status should be 401

  Scenario: Update shop that user does not own returns 403
    Given a shop with ID 999 exists but user does not own it
    When I send an update shop request for shop 999
    Then the response status should be 403
    And the user should receive an error message "access_denied_to_shop"

  # Not found errors
  Scenario: Update non-existent shop returns 404
    Given a shop with ID 888 does not exist
    When I send an update shop request for shop 888
    Then the response status should be 404
    And the user should receive an error message "shop_not_found"
