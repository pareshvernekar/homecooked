Feature: Tenant menu and size-unit API
  As a kitchen operator
  I manage draft menus, publish them for customers, and use size units
  So that sellable menus are available through the API

  Background:
    Given the API is ready
    And the tenant header is "1"
    And a food category exists via the API with name "vegetarian"

  # REQMENU001, REQMENU003, REQMENU007, REQMENU008, REQITEM001, REQITEM005
  Scenario: Create a draft daily menu, add a combo item, publish and unpublish
    Given a food item exists via the API named "Menu Rice"
    And another food item exists via the API named "Menu Dal"
    When I create a draft daily menu named "Acceptance Daily" with category "Mains"
    Then the response status code should be 201
    And the response success flag should be true
    When I add a combo menu item named "Acceptance Thali" priced 5 and 7
    Then the response status code should be 201
    And the response success flag should be true
    When I publish the created menu
    Then the response status code should be 200
    And the response success flag should be true
    When I get the created menu tree
    Then the response status code should be 200
    And the menu status should be "published"
    And the first menu item default_total should be 12
    When I try to add a simple item to the published menu
    Then the response status code should be 400
    When I unpublish the created menu
    Then the response status code should be 200
    And the response success flag should be true
    When I get the created menu tree
    Then the menu status should be "draft"

  # REQMENU002
  Scenario: List menus defaults to published
    Given a food item exists via the API named "List Draft Dish"
    And I create a draft daily menu named "Draft Hidden Menu" with category "Mains"
    And I add a simple menu item named "Draft Dish" priced 9.5
    When I send a GET request to "/api/v1/menus"
    Then the response status code should be 200
    And the response success flag should be true
    And the created menu should not appear in the list
    When I send a GET request to "/api/v1/menus?status=draft"
    Then the response status code should be 200
    And the created menu should appear in the list

  # REQSIZE001, REQSIZE002
  Scenario: List system size units and create a custom unit
    When I send a GET request to "/api/v1/size-units"
    Then the response status code should be 200
    And the response success flag should be true
    And the size unit list should include code "serving"
    And the size unit list should include code "tray"
    When I create a custom size unit with display name "Acceptance Pan"
    Then the response status code should be 201
    And the response success flag should be true
    And the created size unit is_system should be false
