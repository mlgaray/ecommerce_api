Feature: Category Update
  As a shop owner
  I want to update categories
  So that I can keep category information current

  Background:
    Given the user is authenticated for category update operations

  # Success Scenarios
  Scenario: Successfully update a category keeping existing image
    Given I have a category with id 1 and valid update data with existing image
    When I send an update category request
    Then the response status should be 200
    And the user should receive a success message "category_updated_successfully"

  Scenario: Successfully update a category with new image
    Given I have a category with id 1 and valid update data with new image
    When I send an update category request
    Then the response status should be 200
    And the user should receive a success message "category_updated_successfully"

  # HTTP Validations (Infrastructure layer)
  Scenario: Update category with empty name
    Given I have a category with id 1 and empty name for update
    When I send an update category request
    Then the response status should be 400
    And the user should receive an error message "category_name_is_required"

  Scenario: Update category with empty description
    Given I have a category with id 1 and empty description for update
    When I send an update category request
    Then the response status should be 400
    And the user should receive an error message "category_description_is_required"

  Scenario: Update category without any image
    Given I have a category with id 1 and no image for update
    When I send an update category request
    Then the response status should be 400
    And the user should receive an error message "category_image_required"

  Scenario: Update category with existing image missing ID
    Given I have a category with id 1 and existing image without ID
    When I send an update category request
    Then the response status should be 400
    And the user should receive an error message "existing_image_must_have_valid_id"

  Scenario: Update category with existing image missing URL
    Given I have a category with id 1 and existing image without URL
    When I send an update category request
    Then the response status should be 400
    And the user should receive an error message "existing_image_must_have_url"

  Scenario: Update category with oversized new image
    Given I have a category with id 1 and oversized new image for update
    When I send an update category request
    Then the response status should be 400
    And the user should receive an error message "image_size_too_large_max_3mb"

  Scenario: Update category with invalid new image type
    Given I have a category with id 1 and invalid image type for update
    When I send an update category request
    Then the response status should be 400
    And the user should receive an error message "invalid_image_type_only_jpeg_png_allowed"

  Scenario: Update category with invalid category_id format
    Given I have an invalid category_id "abc" for update
    When I send an update category request with invalid id
    Then the response status should be 400
    And the user should receive an error message "invalid_category_id_format"

  # Repository/Domain Errors
  Scenario: Update non-existent category returns 404
    Given I have a category with id 999 that does not exist
    When I send an update category request
    Then the response status should be 404
    And the user should receive an error message "category_not_found"

  Scenario: Update category with duplicate name returns 409
    Given I have a category with id 1 and a name that already exists
    When I send an update category request
    Then the response status should be 409
    And the user should receive an error message "category_already_exists_in_shop"

  # Authentication
  Scenario: Update category without authentication returns 401
    Given the user is not authenticated for category update
    When I send an unauthenticated update category request for category 1
    Then the response status should be 401
