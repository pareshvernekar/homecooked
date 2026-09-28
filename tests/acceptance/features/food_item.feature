Feature: Food Item API interactions
  As an API client
  I interact with food item endpoints over HTTP
  So that menu items can be managed through the API

  Background:
    Given the API is ready
    And the tenant header is "1"
    And a food category exists via the API with name "vegetarian"

  Scenario: Create a food item
    When I create a food item named "Paneer Tikka"
    Then the response status code should be 201
    And the response success flag should be true

  Scenario: List food items
    Given a food item exists via the API named "Dal Tadka"
    When I send a GET request to "/api/v1/food-items"
    Then the response status code should be 200
    And the response success flag should be true

  Scenario: Update a food item
    Given a food item exists via the API named "Butter Naan"
    When I send a PUT request to the created food item with JSON:
      """
      {"name":"Butter Naan Soft"}
      """
    Then the response status code should be 200
    And the response success flag should be true

  Scenario: Delete a food item
    Given a food item exists via the API named "Samosa"
    When I send a DELETE request to the created food item
    Then the response status code should be 204
