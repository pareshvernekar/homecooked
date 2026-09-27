Feature: Food Category API interactions
  As an API client
  I interact with food category endpoints over HTTP
  So that category resources can be managed through the API

  Background:
    Given the API is ready
    And the tenant header is "1"

  Scenario: Create a food category
    When I create a food category named "desserts" with description "Sweet treats"
    Then the response status code should be 201
    And the response success flag should be true

  Scenario: List food categories
    Given a food category exists via the API with name "appetizer"
    When I send a GET request to "/api/v1/categories"
    Then the response status code should be 200
    And the response success flag should be true

  Scenario: Update a food category
    Given a food category exists via the API with name "sides"
    When I update the created category description to "Updated side dishes"
    Then the response status code should be 200
    And the response success flag should be true

  Scenario: Delete a food category
    Given a food category exists via the API with name "beverage"
    When I send a DELETE request to the created category
    Then the response status code should be 204
