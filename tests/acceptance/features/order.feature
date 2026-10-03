Feature: Order intake against published daily menus
  As a kitchen operator
  I take advance orders, record payments, and mark pickup
  So that open tickets use live prices and settled tickets stay frozen

  Background:
    Given the API is ready
    And the tenant header is "1"
    And a food category exists via the API with name "vegetarian"
    And a published daily menu exists with a simple item priced 12

  # REQORDER001, REQORDER004, REQORDER005, REQOLINE001, REQOLINE003,
  # REQPAY001, REQPAY002, REQITEM006S03, REQLIFE001S01, REQLIFE004
  Scenario: Create order, add line, live price moves, pay, then freeze at PICKEDUP
    When I create an order for the published menu for customer "Asha" phone "555-0100" expected in 3 hours
    Then the response status code should be 201
    And the response success flag should be true
    And the order status should be "RECEIVED"
    When I add a line for the menu item with quantity 2
    Then the response status code should be 201
    And the order line unit_price should be 12
    And the order line extended_amount should be 24
    When I get the created order
    Then the order charged_total should be 24
    And the order paid_amount should be 0
    And the order payment_received should be false
    When I update the published size option price to 15
    And I get the created order
    Then the order charged_total should be 30
    When I record a cash payment of 30 on the order
    Then the response status code should be 201
    When I get the created order
    Then the order payment_received should be true
    And the order balance should be 0
    When I accept the order
    Then the response status code should be 200
    And the order status should be "ACCEPTED"
    When I start preparing the order
    Then the response status code should be 200
    And the order status should be "IN_PROGRESS"
    When I mark the order as READY
    Then the response status code should be 200
    And the order status should be "READY"
    When I mark the order as PICKEDUP
    Then the response status code should be 200
    And the order status should be "PICKEDUP"
    When I update the published size option price to 99
    And I get the created order
    Then the order charged_total should be 30
