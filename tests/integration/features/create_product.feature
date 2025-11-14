Feature: Product Creation
  As a shop owner
  I want to create products
  So that I can sell them in my shop

  Scenario: Successfully create a product with valid data
    Given I have valid product data with images
    When I send a create product request
    Then the response status should be 201
    And the product should be created successfully

  # HTTP Validations (Infrastructure layer - Contracts)
  Scenario: Create product without images
    Given I have product data without images
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "at_least_one_image_is_required"

  Scenario: Create product with empty name
    Given I have product data with empty name
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "product_name_is_required"

  Scenario: Create product with empty description
    Given I have product data with empty description
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "product_description_is_required"

  Scenario: Create product without category
    Given I have product data without category
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "category_id_is_required"

  Scenario: Create product with invalid shop_id
    Given I have product data with invalid shop_id
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "shop_id_is_required"

  Scenario: Create product with image size too large
    Given I have product data with oversized image
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "image_size_too_large_max_3mb"

  Scenario: Create product with invalid image type
    Given I have product data with invalid image type
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "invalid_image_type_only_jpeg_png_allowed"

  # Business Validations (Domain layer - Product.Validate())
  Scenario: Create product with negative price
    Given I have product data with negative price
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "product_price_must_be_positive"

  Scenario: Create product with negative stock
    Given I have product data with negative stock
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "product_stock_cannot_be_negative"

  Scenario: Create product with negative minimum stock
    Given I have product data with negative minimum stock
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "product_minimum_stock_cannot_be_negative"

  Scenario: Create product with minimum stock but no stock
    Given I have product data with minimum stock but no stock
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "minimum_stock_requires_stock"

  Scenario: Create product with minimum stock greater than stock
    Given I have product data with minimum stock greater than stock
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "product_minimum_stock_cannot_be_greater_than_stock"

  Scenario: Create product as promotional without promotional price
    Given I have product data as promotional without promotional price
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "promotional_product_requires_promotional_price"

  Scenario: Create product with promotional price not lower than regular price
    Given I have product data with promotional price not lower than price
    When I send a create product request
    Then the response status should be 400
    And the user should receive an error message "promotional_price_must_be_lower_than_regular_price"
