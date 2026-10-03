Feature: Order notification outbox and cook admin phone
  As a kitchen operator
  I configure cook alerts and audit SMS outbox rows for orders
  So that cooks and customers get lifecycle SMS without blocking order actions

  Background:
    Given the API is ready
    And the tenant header is "1"
    And the cook admin phone is cleared
    And a food category exists via the API with name "vegetarian"
    And a published daily menu exists with a simple item priced 12

  # REQNOTIF001S01, REQNOTIF001S02
  Scenario: Set and clear cook admin phone
    When I set the cook admin phone to "555-0199"
    Then the response status code should be 200
    And the cook admin phone should be "555-0199"
    When I clear the cook admin phone
    Then the response status code should be 200
    And the cook admin phone should be empty

  # REQNOTIF002S02
  Scenario: Create skips cook notification when phone unset
    When I create an order for the published menu for customer "Asha" phone "555-0100" expected in 3 hours
    Then the response status code should be 201
    When I list notifications for the created order
    Then the response status code should be 200
    And the order should have 0 notifications

  # REQNOTIF002S01, REQNOTIF005S01
  Scenario: Create enqueues cook notification when phone configured
    Given the cook admin phone is "555-0199"
    When I create an order for the published menu for customer "Asha" phone "555-0100" expected in 3 hours
    Then the response status code should be 201
    When I list notifications for the created order
    Then the response status code should be 200
    And the order should have a notification for event "order.created" to "555-0199"
    And the order should have 1 notification

  # REQNOTIF002S03, REQNOTIF002S05, REQNOTIF002S06
  Scenario: Accept ready pickup enqueue customer SMS; preparing does not
    Given the cook admin phone is "555-0199"
    When I create an order for the published menu for customer "Asha" phone "555-0100" expected in 3 hours
    Then the response status code should be 201
    When I accept the order
    Then the response status code should be 200
    When I list notifications for the created order
    Then the order should have a notification for event "order.accepted" to "555-0100"
    And I remember the notification count for the created order
    When I start preparing the order
    Then the response status code should be 200
    When I list notifications for the created order
    Then the order notification count should be unchanged
    And the order should not have a notification for event "order.in_progress"
    When I mark the order as READY
    Then the response status code should be 200
    When I mark the order as PICKEDUP
    Then the response status code should be 200
    When I list notifications for the created order
    Then the order should have a notification for event "order.ready" to "555-0100"
    And the order should have a notification for event "order.picked_up" to "555-0100"
    And the order should have 4 notifications

  # REQNOTIF002S04
  Scenario: Decline enqueues customer notification including refuse reason
    Given the cook admin phone is "555-0199"
    When I create an order for the published menu for customer "Asha" phone "555-0100" expected in 3 hours
    Then the response status code should be 201
    When I refuse the order with reason "Out of rice"
    Then the response status code should be 200
    When I list notifications for the created order
    Then the order should have a notification for event "order.declined" to "555-0100"
    And the "order.declined" notification body should contain "Out of rice"

  # REQNOTIF003S01, REQNOTIF004S01
  Scenario: Worker delivers pending outbox via local SMS provider
    Given the cook admin phone is "555-0199"
    When I create an order for the published menu for customer "Asha" phone "555-0100" expected in 3 hours
    Then the response status code should be 201
    When I list notifications for the created order
    Then the order should have a notification for event "order.created" to "555-0199"
    And I wait until the "order.created" notification status is "delivered"

  # REQNOTIF005S01
  Scenario: Unknown order notifications are not found
    When I list notifications for order "does-not-exist"
    Then the response status code should be 404
