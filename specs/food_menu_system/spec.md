# Food Menu System Specification

## Overview
The Food Menu System is a REST API that allows administrators to create and manage food catalogs, weekly menus, and catering menus. **This system supports a multi-tenant architecture**, where each tenant operates in isolation with their own data.

## Core Features
1. Admin can create a catalog of food items for a specific tenant
2. Admin can create weekly menus for daily tiffin from items in the catalog for a specific tenant
3. Admin can create catering menus from items in the catalog for a specific tenant
4. Admin can specify size and price for menu items based on weekly vs. catering menus for a specific tenant
5. Users can view menus for both catering and weekly options for their tenant
6. Users can place orders based on weekly and catering menus for their tenant
7. Admin can notify users when their orders are ready for a specific tenant
8. Admin can send menus to one or more users for a specific tenant

## User Roles
- **Admin**: Can create and manage food items, menus, and orders for a specific tenant. Can send notifications for a specific tenant.
- **User**: Can view menus, place orders, and receive notifications for their tenant.

## Tenant Model
All API requests must include a tenant ID in the request headers:
- **X-Tenant-ID**: The unique identifier for the tenant (string)

All entities are scoped to a specific tenant via the tenant ID field.

### Domain Model

### Entities
- **Tenant**: Represents a client organization that uses the system.
- **FoodCatalog**: Represents a collection of food items for a specific tenant.
- **FoodItem**: Represents individual food items in a catalog for a specific tenant.
- **WeeklyMenu**: Represents a weekly menu for daily tiffin for a specific tenant.
- **CateringMenu**: Represents a menu for catering events for a specific tenant.
- **MenuItem**: Represents an item in a menu with size and price for a specific tenant.
- **Order**: Represents a user's order for a specific tenant.
- **Notification**: Represents a notification sent to users for a specific tenant.

### Relationships
- A **Tenant** can have multiple **FoodCatalog**, **WeeklyMenu**, **CateringMenu**, **Order**, and **Notification** instances.
- All entities are scoped to a specific **Tenant** via the tenant ID field.
- A **FoodCatalog** contains multiple **FoodItem** instances for a specific tenant.
- A **FoodItem** can be included in multiple **WeeklyMenu** and **CateringMenu** instances via **MenuItem** entries for a specific tenant.
- A **WeeklyMenu** is composed of **MenuItem** instances, each referencing a **FoodItem** with specific size, price, and sequence for a specific tenant.
- A **CateringMenu** is composed of **MenuItem** instances, each referencing a **FoodItem** with specific size, price, and sequence for a specific tenant.
- Each **MenuItem** in a **WeeklyMenu** or **CateringMenu** has a unique sequence number to define the order of items for a specific tenant.
- A **User** can place multiple **Order** instances for a specific tenant.
- An **Order** can include multiple **MenuItem** instances from either a **WeeklyMenu** or **CateringMenu** for a specific tenant.
- An **Order** can have one or more **Notification** instances for a specific tenant.

## Data Models

### WeeklyMenu
A **WeeklyMenu** is a collection of **MenuItem** entries that define the daily tiffin menu for a specific period for a specific tenant.

```json
{
  "id": "string",
  "tenantId": "string",  // Unique identifier for the tenant
  "name": "string",
  "description": "string",
  "startDate": "datetime",
  "endDate": "datetime",
  "menuItems": [
    {
      "id": "string",
      "foodItemId": "string",
      "size": "string",
      "price": "number",
      "sequence": "number"
    }
  ],
  "createdAt": "datetime",
  "updatedAt": "datetime"
}
```

### CateringMenu
A **CateringMenu** is a collection of **MenuItem** entries for catering events for a specific tenant.

```json
{
  "id": "string",
  "tenantId": "string",  // Unique identifier for the tenant
  "name": "string",
  "description": "string",
  "menuItems": [
    {
      "id": "string",
      "foodItemId": "string",
      "size": "string",
      "price": "number",
      "sequence": "number"
    }
  ],
  "createdAt": "datetime",
  "updatedAt": "datetime"
}
```

### MenuItem
A **MenuItem** represents an individual item in a **WeeklyMenu** or **CateringMenu** for a specific tenant.

```json
{
  "id": "string",
  "tenantId": "string",  // Unique identifier for the tenant
  "menuId": "string",
  "menuType": "string" (weekly or catering),
  "foodItemId": "string",
  "size": "string",
  "price": "number",
  "sequence": "number",
  "createdAt": "datetime",
  "updatedAt": "datetime"
}
```

### FoodCatalog
```json
{
  "id": "string",
  "tenantId": "string",  // Unique identifier for the tenant
  "name": "string",
  "description": "string",
  "createdAt": "datetime",
  "updatedAt": "datetime",
  "foodItems": [
    {
      "id": "string"
    }
  ]
}
```

### FoodItem
```json
{
  "id": "string",
  "tenantId": "string",  // Unique identifier for the tenant
  "name": "string",
  "description": "string",
  "price": "number",
  "categoryId": "string",
  "createdAt": "datetime",
  "updatedAt": "datetime"
}
```

### Database Schema

#### Tables

**tenant**
```sql
CREATE TABLE tenant (
  id VARCHAR(50) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL
);
```

**menu_item**
```sql
CREATE TABLE menu_item (
  id VARCHAR(50) PRIMARY KEY,
  tenant_id VARCHAR(50) NOT NULL,
  food_item_id VARCHAR(50) NOT NULL,
  menu_id VARCHAR(50) NOT NULL,
  menu_type ENUM('weekly', 'catering') NOT NULL,
  description VARCHAR(100),
  size VARCHAR(50) NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  sequence INT NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) ON DELETE CASCADE,
  FOREIGN KEY (food_item_id) REFERENCES food_item(id) ON DELETE CASCADE,
  FOREIGN KEY (menu_id) REFERENCES weekly_menu(id) ON DELETE CASCADE,
  FOREIGN KEY (menu_id) REFERENCES catering_menu(id) ON DELETE CASCADE,
  UNIQUE KEY (tenant_id, menu_id, sequence)  -- Ensures unique sequence within each menu for each tenant
);
```

**weekly_menu**
```sql
CREATE TABLE weekly_menu (
  id VARCHAR(50) PRIMARY KEY,
  tenant_id VARCHAR(50) NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  start_date DATETIME NOT NULL,
  end_date DATETIME NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) ON DELETE CASCADE
);
```

**catering_menu**
```sql
CREATE TABLE catering_menu (
  id VARCHAR(50) PRIMARY KEY,
  tenant_id VARCHAR(50) NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  event_date DATETIME NOT NULL,
  event_location VARCHAR(200) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) ON DELETE CASCADE
);
```

**food_catalog**
```sql
CREATE TABLE food_catalog (
  id VARCHAR(50) PRIMARY KEY,
  tenant_id VARCHAR(50) NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) ON DELETE CASCADE
);
```

**food_item**
```sql
CREATE TABLE food_item (
  id VARCHAR(50) PRIMARY KEY,
  tenant_id VARCHAR(50) NOT NULL,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  category_id VARCHAR(50) NOT NULL,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) ON DELETE CASCADE
);
```

**food_category**
```sql
CREATE TABLE food_category (
  id VARCHAR(50) PRIMARY KEY,
  tenant_id VARCHAR(50) NOT NULL,
  name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  created_at DATETIME NOT NULL,
  updated_at DATETIME NOT NULL,
  FOREIGN KEY (tenant_id) REFERENCES tenant(id) ON DELETE CASCADE
);
```

-- Link food_item to categories
```sql
ALTER TABLE food_item
  ADD CONSTRAINT fk_food_item_category
  FOREIGN KEY (category_id) REFERENCES food_category(id) ON DELETE RESTRICT;
```

### Relationships

1. **MenuItems to WeeklyMenus**:
   - Each `menu_item` in a `weekly_menu` is uniquely identified by the combination of `menu_id` (weekly_menu.id) and `sequence`.
   - A `weekly_menu` can contain multiple `menu_items`.
   - A `menu_item` can belong to only one `weekly_menu`.

2. **MenuItems to CateringMenus**:
   - Each `menu_item` in a `catering_menu` is uniquely identified by the combination of `menu_id` (catering_menu.id) and `sequence`.
   - A `catering_menu` can contain multiple `menu_items`.
   - A `menu_item` can belong to only one `catering_menu`.

3. **Shared MenuItems**:
   - The same `food_item` can appear in multiple menus (both weekly and catering) with different configurations (size, price, sequence).
   - The `menu_type` field ensures that each `menu_item` is correctly categorized and associated with the appropriate menu type.

### MenuItem
```json
{
  "id": "string",
  "tenantId": "string",
  "menuId": "string",
  "menuType": "string" (weekly or catering),
  "foodItemId": "string",
  "size": "string",
  "price": "number",
  "sequence": "number",
  "createdAt": "datetime",
  "updatedAt": "datetime"
}
```

### API Contract

#### Menu Items Endpoints

**POST /api/v1/menus/{menuType}/menu-items**
- Add a menu item to a specific menu type (weekly or catering)
- Request Body:
```json
{
  "foodItemId": "string",
  "size": "string",
  "price": "number",
  "sequence": "number"
}
```

**GET /api/v1/menus/{menuType}/{menuId}/menu-items**
- List all menu items for a specific menu
- Returns:
```json
[
  {
    "id": "string",
    "foodItemId": "string",
    "size": "string",
    "price": "number",
    "sequence": "number"
  }
]
```

**PUT /api/v1/menus/{menuType}/{menuId}/menu-items/{menuItemId}**
- Update a specific menu item
- Request Body:
```json
{
  "size": "string",
  "price": "number",
  "sequence": "number"
}
```

**DELETE /api/v1/menus/{menuType}/{menuId}/menu-items/{menuItemId}**
- Delete a specific menu item

### Order
```json
{
  "id": "string",
  "userId": "string",
  "menuType": "string" (weekly or catering),
  "menuId": "string",
  "status": "string" (pending, preparing, ready, delivered),
  "orderDate": "datetime",
  "deliveryDate": "datetime",
  "totalPrice": "number",
  "items": [
    {
      "menuItemId": "string",
      "quantity": "number"
    }
  ],
  "createdAt": "datetime",
  "updatedAt": "datetime"
}
```

### Notification
```json
{
  "id": "string",
  "userId": "string",
  "orderId": "string",
  "type": "string" (order_placed, order_preparing, order_ready, order_delivered),
  "message": "string",
  "isRead": "boolean",
  "createdAt": "datetime"
}
```

## Test Cases
### Food Items
1. **Create a Food Item**: Verify that a new food item can be created successfully.
2. **List Food Items**: Verify that all food items can be listed.
3. **Get a Specific Food Item**: Verify that a specific food item can be retrieved.
4. **Update a Food Item**: Verify that a food item can be updated.
5. **Delete a Food Item**: Verify that a food item can be deleted.

### Weekly Menus
1. **Create a Weekly Menu**: Verify that a new weekly menu can be created.
2. **List Weekly Menus**: Verify that all weekly menus can be listed.
3. **Get a Specific Weekly Menu**: Verify that a specific weekly menu can be retrieved.
4. **Add Menu Items to Weekly Menu**: Verify that menu items can be added to a weekly menu.
5. **Update a Weekly Menu**: Verify that a weekly menu can be updated.
6. **Delete a Weekly Menu**: Verify that a weekly menu can be deleted.

### Catering Menus
1. **Create a Catering Menu**: Verify that a new catering menu can be created.
2. **List Catering Menus**: Verify that all catering menus can be listed.
3. **Get a Specific Catering Menu**: Verify that a specific catering menu can be retrieved.
4. **Add Menu Items to Catering Menu**: Verify that menu items can be added to a catering menu.
5. **Update a Catering Menu**: Verify that a catering menu can be updated.
6. **Delete a Catering Menu**: Verify that a catering menu can be deleted.

### Menu Items
1. **Add Menu Item to Weekly Menu**: Verify that a menu item can be added to a weekly menu.
2. **Add Menu Item to Catering Menu**: Verify that a menu item can be added to a catering menu.
3. **List Menu Items in Weekly Menu**: Verify that menu items in a weekly menu can be listed.
4. **List Menu Items in Catering Menu**: Verify that menu items in a catering menu can be listed.
5. **Update Menu Item in Weekly Menu**: Verify that a menu item in a weekly menu can be updated.
6. **Update Menu Item in Catering Menu**: Verify that a menu item in a catering menu can be updated.
7. **Delete Menu Item from Weekly Menu**: Verify that a menu item can be deleted from a weekly menu.
8. **Delete Menu Item from Catering Menu**: Verify that a menu item can be deleted from a catering menu.

### Orders
1. **Place an Order**: Verify that a new order can be placed.
2. **List Orders**: Verify that all orders can be listed.
3. **Get a Specific Order**: Verify that a specific order can be retrieved.
4. **List Orders for a User**: Verify that orders for a specific user can be listed.

### Notifications
1. **Send a Notification**: Verify that a notification can be sent.
2. **List Notifications**: Verify that all notifications can be listed.
3. **Get a Specific Notification**: Verify that a specific notification can be retrieved.
4. **List Notifications for a User**: Verify that notifications for a specific user can be listed.

### FoodCatalog
1. **Create a Food Catalog**: Verify that a new food catalog can be created successfully.
2. **List Food Catalogs**: Verify that all food catalogs can be listed.
3. **Get a Specific Food Catalog**: Verify that a specific food catalog can be retrieved.
4. **Update a Food Catalog**: Verify that a food catalog can be updated.
5. **Delete a Food Catalog**: Verify that a food catalog can be deleted.

## Security
- **Authentication**: JWT (JSON Web Tokens) for API access.
- **Authorization**: Role-based access control (RBAC) for different user roles.
- **Data Protection**: Encrypt sensitive data at rest and in transit.

## Error Handling
- **Standard Error Responses**: Consistent error response format.
- **HTTP Status Codes**: Appropriate HTTP status codes for different scenarios.
- **Validation Errors**: Detailed error messages for validation failures.

## Deployment
- **Containerization**: Docker for easy deployment.
- **Orchestration**: Kubernetes for managing containerized applications.
- **CI/CD**: Automated build and deployment pipelines.

## Monitoring and Logging
- **Logging**: Comprehensive logging for debugging and auditing.
- **Monitoring**: Tools for monitoring application performance and health.
- **Alerts**: Alerts for critical issues and anomalies.

## Performance
- **Scalability**: Design for horizontal scalability.
- **Load Testing**: Regular load testing to ensure performance under load.
- **Caching**: Implement caching for frequently accessed data.

## Documentation
- **API Documentation**: Swagger/OpenAPI for API documentation.
- **User Guides**: Guides for users and administrators.
- **Developer Documentation**: Detailed documentation for developers.

## Compliance
- **Data Privacy**: Compliance with data privacy regulations.
- **Security Standards**: Compliance with security standards and best practices.

## Future Enhancements
- **Mobile App**: Development of a mobile application for easier access.
- **Integration**: Integration with third-party services and platforms.
- **AI Recommendations**: AI-driven menu recommendations for users.

## Conclusion
The Food Menu System is designed to provide a comprehensive solution for managing food items, menus, and orders. It ensures a seamless experience for both administrators and users, with robust features for creating, updating, and managing menus and orders.

--- 

### Test-Driven Development (TDD) Plan

1. **Write Tests**: Write unit tests for all API endpoints and business logic.
2. **Implement Minimal Functionality**: Implement the minimal functionality required to pass the tests.
3. **Refactor**: Refactor the code to improve readability and maintainability.
4. **Add More Tests**: Write additional tests to cover edge cases and new features.
5. **Iterate**: Repeat the process until all requirements are met.

--- 

### Example Test Cases for TDD

#### Food Items
```go
testFuncCreateFoodItem := func(t *testing.T) {
    // Setup
    foodItem := FoodItem{
        Name:        "Chicken Biryani",
        Description: "Spicy biryani with chicken",
        Price:       250.00,
          CategoryId:  "category-uuid-1",
    }
    
    // Execute
    response, err := http.Post("/api/v1/food-items", "application/json", foodItem)
    if err != nil {
        t.Fatalf("Error creating food item: %v", err)
    }
    
    // Assert
    if response.StatusCode != http.StatusCreated {
        t.Errorf("Expected status code 201, got %d", response.StatusCode)
    }
    
    // Verify response body
    var createdFoodItem FoodItem
    err = json.Unmarshal(response.Body, &createdFoodItem)
    if err != nil {
        t.Fatalf("Error parsing response: %v", err)
    }
    if createdFoodItem.Name != foodItem.Name {
        t.Errorf("Expected name %s, got %s", foodItem.Name, createdFoodItem.Name)
    }
}
```

--- 

### Example Implementation for TDD

#### Create Food Item
```go
handlerCreateFoodItem := func(w http.ResponseWriter, r *http.Request) {
    // Parse request body
    var foodItem FoodItem
    if err := json.NewDecoder(r.Body).Decode(&foodItem); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Validate input
    if foodItem.Name == "" {
        http.Error(w, "Name is required", http.StatusBadRequest)
        return
    }
    
    // Create food item
    createdFoodItem, err := foodItemRepository.Create(foodItem)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(createdFoodItem)
}
```




#### Order Table
```sql
CREATE TABLE order (
  id VARCHAR(50) PRIMARY KEY,
  user_id VARCHAR(50) NOT NULL,
  menu_type VARCHAR(10) NOT NULL, -- weekly or catering
  menu_id VARCHAR(50) NOT NULL,
  status VARCHAR(20) NOT NULL, -- pending, preparing, ready, delivered
  order_date TIMESTAMP WITH TIME ZONE NOT NULL,
  delivery_date TIMESTAMP WITH TIME ZONE,
  total_price DECIMAL(10, 2) NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Order Item Table
```sql
CREATE TABLE order_item (
  id VARCHAR(50) PRIMARY KEY,
  order_id VARCHAR(50) REFERENCES order(id),
  menu_item_id VARCHAR(50) REFERENCES menu_item(id),
  quantity INTEGER NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Notification Table
```sql
CREATE TABLE notification (
  id VARCHAR(50) PRIMARY KEY,
  user_id VARCHAR(50) NOT NULL,
  order_id VARCHAR(50),
  type VARCHAR(20) NOT NULL, -- order_placed, order_preparing, order_ready, order_delivered
  message TEXT NOT NULL,
  is_read BOOLEAN DEFAULT false,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```