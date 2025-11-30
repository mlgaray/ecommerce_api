Feature: Category Creation
  As a shop owner
  I want to create categories
  So that I can organize my products

  Scenario: Successfully create a category with valid data
    Given I have valid category data with image
    When I send a create category request
    Then the response status should be 201
    And the category should be created successfully

  # HTTP Validations (Infrastructure layer - Contracts)
  Scenario: Create category without image
    Given I have valid category data without image
    When I send a create category request
    Then the response status should be 400
    And the user should receive an error message "category_image_required"

  Scenario: Create category with empty name
    Given I have category data with empty name
    When I send a create category request
    Then the response status should be 400
    And the user should receive an error message "category_name_is_required"

  Scenario: Create category with empty description
    Given I have category data with empty description
    When I send a create category request
    Then the response status should be 400
    And the user should receive an error message "category_description_is_required"

  Scenario: Create category with invalid shop_id
    Given I have category data with invalid shop_id
    When I send a create category request
    Then the response status should be 400
    And the user should receive an error message "shop_id_is_required"

  Scenario: Create category with image size too large
    Given I have category data with oversized image
    When I send a create category request
    Then the response status should be 400
    And the user should receive an error message "image_size_too_large_max_3mb"

  Scenario: Create category with invalid image type
    Given I have category data with invalid image type
    When I send a create category request
    Then the response status should be 400
    And the user should receive an error message "invalid_image_type_only_jpeg_png_allowed"

  # Database Validations
  Scenario: Create category with duplicate name in same shop
    Given I have category data with duplicate name
    When I send a create category request
    Then the response status should be 409
    And the user should receive an error message "category_already_exists_in_shop"