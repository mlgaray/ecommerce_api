Feature: Product Update
  As a shop owner
  I want to update products
  So that I can keep product information current

  Scenario: Successfully update a product with valid data
    Given I have a product with id 1 and valid update data
    When I send an update product request
    Then the response status should be 200
    And the user should receive a success message "product_updated_successfully"

  Scenario: Update product keeping existing images only
    Given I have a product with id 1 and keep existing images
    When I send an update product request
    Then the response status should be 200
    And the user should receive a success message "product_updated_successfully"

  Scenario: Update product adding new images
    Given I have a product with id 1 and add new images
    When I send an update product request
    Then the response status should be 200
    And the user should receive a success message "product_updated_successfully"

  # HTTP Validations (Infrastructure layer)
  Scenario: Update product removing all images without adding new ones
    Given I have a product with id 1 and no images
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "at_least_one_image_is_required"

  Scenario: Update product with empty name
    Given I have a product with id 1 and empty name
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "product_name_is_required"

  Scenario: Update product with empty description
    Given I have a product with id 1 and empty description
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "product_description_is_required"

  Scenario: Update product without category
    Given I have a product with id 1 and no category
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "category_id_is_required"

  Scenario: Update product with invalid shop_id
    Given I have a product with id 1 and invalid shop_id
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "shop_id_is_required"

  Scenario: Update product with oversized new image
    Given I have a product with id 1 and oversized new image
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "image_size_too_large_max_3mb"

  Scenario: Update product with invalid new image type
    Given I have a product with id 1 and invalid new image type
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "invalid_image_type_only_jpeg_png_allowed"

  Scenario: Update product with invalid product_id format
    Given I have an invalid product_id "abc"
    When I send an update product request with invalid id
    Then the response status should be 400
    And the user should receive an error message "invalid_product_id_format"

  # Business Validations (Domain layer - Product.Validate())
  Scenario: Update product with negative price
    Given I have a product with id 1 and negative price
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "product_price_must_be_positive"

  Scenario: Update product with negative stock
    Given I have a product with id 1 and negative stock
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "product_stock_cannot_be_negative"

  Scenario: Update product with negative minimum stock
    Given I have a product with id 1 and negative minimum stock
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "product_minimum_stock_cannot_be_negative"

  Scenario: Update product with minimum stock but no stock
    Given I have a product with id 1 with minimum stock but no stock
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "minimum_stock_requires_stock"

  Scenario: Update product with minimum stock greater than stock
    Given I have a product with id 1 with minimum stock greater than stock
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "product_minimum_stock_cannot_be_greater_than_stock"

  Scenario: Update product as promotional without promotional price
    Given I have a product with id 1 as promotional without promotional price
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "promotional_product_requires_promotional_price"

  Scenario: Update product with promotional price greater than or equal to regular price
    Given I have a product with id 1 with promotional price not lower than price
    When I send an update product request
    Then the response status should be 400
    And the user should receive an error message "promotional_price_must_be_lower_than_regular_price"
