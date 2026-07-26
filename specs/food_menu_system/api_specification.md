# Food Menu System - REST API Specification

## Overview
This document defines the complete REST API specification for the Food Menu System. The system supports multi-tenancy, where each tenant operates in isolation with their own data.

**Base URL**: `/api/v1`

**Authentication**: JWT Bearer Token (required for all authenticated endpoints)

**Tenant Context**: All requests must include `X-Tenant-ID` header for admin operations. For user operations, the tenant ID is derived from the authenticated user's profile.

---

## Table of Contents
1. [Food Items API](#food-items-api)
2. [Categories API](#categories-api)
3. [Weekly Menus API](#weekly-menus-api)
4. [Catering Menus API](#catering-menus-api)
5. [Menu Items API](#menu-items-api)
6. [Orders API](#orders-api)
7. [Notifications API](#notifications-api)

---

## Food Items API

### Create Food Item
**Endpoint**: `POST /api/v1/food-items`  
**Role**: Admin only  
**Headers**: 
- `Authorization`: Bearer <token>
- `X-Tenant-ID`: <tenant-id>

**Request Body**:
```json
{
  "name": "Butter Chicken",
  "description": "Creamy tomato-based chicken curry",
  "category": "Non-Vegetarian"
}
```

**Response (201 Created)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Butter Chicken",
  "description": "Creamy tomato-based chicken curry",
  "category": "Non-Vegetarian",
  "createdAt": "2026-05-08T10:30:00Z",
  "updatedAt": "2026-05-08T10:30:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid input data
- `401 Unauthorized`: Missing or invalid authentication token
- `403 Forbidden`: Insufficient permissions
- `409 Conflict`: Food item already exists for this tenant

---

### Get Food Item by ID
**Endpoint**: `GET /api/v1/food-items/{id}`  
**Role**: Admin, User (own items)  
**Path Parameters**:
- `id`: UUID of the food item

**Response (200 OK)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Butter Chicken",
  "description": "Creamy tomato-based chicken curry",
  "category": "Non-Vegetarian",
  "createdAt": "2026-05-08T10:30:00Z",
  "updatedAt": "2026-05-08T10:30:00Z"
}
```

---

### List Food Items
**Endpoint**: `GET /api/v1/food-items`  
**Role**: Admin, User (own items)  
**Query Parameters**:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)
- `category`: Filter by category (optional)
- `search`: Search in name/description (optional)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "uuid-1234-5678",
      "name": "Butter Chicken",
      "category": "Non-Vegetarian"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 45,
    "totalPages": 5
  }
}
```

---

### Update Food Item
**Endpoint**: `PUT /api/v1/food-items/{id}`  
**Role**: Admin only  
**Path Parameters**:
- `id`: UUID of the food item

**Request Body**:
```json
{
  "name": "Grilled Chicken",
  "description": "Herb-marinated grilled chicken",
  "category": "Non-Vegetarian"
}
```

**Response (200 OK)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Grilled Chicken",
  "description": "Herb-marinated grilled chicken",
  "category": "Non-Vegetarian",
  "updatedAt": "2026-05-08T11:00:00Z"
}
```

---

### Delete Food Item
**Endpoint**: `DELETE /api/v1/food-items/{id}`  
**Role**: Admin only  
**Path Parameters**:
- `id`: UUID of the food item

**Response (204 No Content)**: Success

**Error Responses**:
- `404 Not Found`: Food item not found
- `409 Conflict`: Food item is referenced by menus

---

## Categories API

### Create Category
**Endpoint**: `POST /api/v1/categories`  
**Role**: Admin only  
**Headers**: 
- `Authorization`: Bearer <token>
- `X-Tenant-ID`: <tenant-id>

**Request Body**:
```json
{
  "name": "Vegetarian",
  "description": "Plant-based meals"
}
```

**Response (201 Created)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Vegetarian",
  "description": "Plant-based meals",
  "createdAt": "2026-05-08T10:30:00Z",
  "updatedAt": "2026-05-08T10:30:00Z"
}
```

---

### Get Category by ID
**Endpoint**: `GET /api/v1/categories/{id}`  
**Role**: Admin, User (own categories)

**Response (200 OK)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Vegetarian",
  "description": "Plant-based meals",
  "createdAt": "2026-05-08T10:30:00Z",
  "updatedAt": "2026-05-08T10:30:00Z"
}
```

---

### List Categories
**Endpoint**: `GET /api/v1/categories`  
**Role**: Admin, User (own categories)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "uuid-1234-5678",
      "name": "Vegetarian",
      "description": "Plant-based meals"
    }
  ],
  "meta": {
    "total": 5,
    "totalCount": 5
  }
}
```

---

### Update Category
**Endpoint**: `PUT /api/v1/categories/{id}`  
**Role**: Admin only

**Request Body**:
```json
{
  "name": "Veg",
  "description": "Vegetable-based meals"
}
```

**Response (200 OK)**: Success

---

### Delete Category
**Endpoint**: `DELETE /api/v1/categories/{id}`  
**Role**: Admin only

**Response (204 No Content)**: Success

**Error Responses**:
- `409 Conflict`: Category has associated food items

---

## Weekly Menus API

### Create Weekly Menu
**Endpoint**: `POST /api/v1/weekly-menus`  
**Role**: Admin only  
**Headers**: 
- `Authorization`: Bearer <token>
- `X-Tenant-ID`: <tenant-id>

**Request Body**:
```json
{
  "name": "Summer Special Week",
  "description": "Fresh seasonal dishes for the week",
  "startDate": "2026-05-11T00:00:00Z",
  "endDate": "2026-05-17T23:59:59Z"
}
```

**Response (201 Created)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Summer Special Week",
  "description": "Fresh seasonal dishes for the week",
  "startDate": "2026-05-11T00:00:00Z",
  "endDate": "2026-05-17T23:59:59Z",
  "menuItems": [],
  "createdAt": "2026-05-08T10:30:00Z",
  "updatedAt": "2026-05-08T10:30:00Z"
}
```

---

### Get Weekly Menu by ID
**Endpoint**: `GET /api/v1/weekly-menus/{id}`  
**Role**: Admin, User (own menus)

**Response (200 OK)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Summer Special Week",
  "description": "Fresh seasonal dishes for the week",
  "startDate": "2026-05-11T00:00:00Z",
  "endDate": "2026-05-17T23:59:59Z",
  "menuItems": [
    {
      "id": "uuid-menuitem-1",
      "foodItemId": "uuid-food-1",
      "size": "Regular",
      "price": 12.99,
      "sequence": 1
    }
  ]
}
```

---

### List Weekly Menus
**Endpoint**: `GET /api/v1/weekly-menus`  
**Role**: Admin, User (own menus)  
**Query Parameters**:
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)
- `startDate`: Filter by start date (optional)
- `endDate`: Filter by end date (optional)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "uuid-1234-5678",
      "name": "Summer Special Week",
      "startDate": "2026-05-11T00:00:00Z",
      "endDate": "2026-05-17T23:59:59Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 20,
    "totalPages": 2
  }
}
```

---

### Update Weekly Menu
**Endpoint**: `PUT /api/v1/weekly-menus/{id}`  
**Role**: Admin only

**Request Body**:
```json
{
  "name": "Updated Summer Special",
  "description": "Updated description",
  "startDate": "2026-05-11T00:00:00Z",
  "endDate": "2026-05-17T23:59:59Z"
}
```

**Response (200 OK)**: Success

---

### Delete Weekly Menu
**Endpoint**: `DELETE /api/v1/weekly-menus/{id}`  
**Role**: Admin only

**Response (204 No Content)**: Success

**Error Responses**:
- `404 Not Found`: Weekly menu not found
- `409 Conflict`: Weekly menu has associated orders

---

## Catering Menus API

### Create Catering Menu
**Endpoint**: `POST /api/v1/catering-menus`  
**Role**: Admin only

**Request Body**:
```json
{
  "name": "Corporate Event Menu",
  "description": "Premium catering for corporate event",
  "eventDate": "2026-05-15T12:00:00Z",
  "eventLocation": "Conference Center, Room 101"
}
```

**Response (201 Created)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Corporate Event Menu",
  "description": "Premium catering for corporate event",
  "eventDate": "2026-05-15T12:00:00Z",
  "eventLocation": "Conference Center, Room 101",
  "menuItems": [],
  "createdAt": "2026-05-08T10:30:00Z",
  "updatedAt": "2026-05-08T10:30:00Z"
}
```

---

### Get Catering Menu by ID
**Endpoint**: `GET /api/v1/catering-menus/{id}`  
**Role**: Admin, User (own menus)

**Response (200 OK)**:
```json
{
  "id": "uuid-1234-5678",
  "tenantId": "tenant-uuid-1234",
  "name": "Corporate Event Menu",
  "description": "Premium catering for corporate event",
  "eventDate": "2026-05-15T12:00:00Z",
  "eventLocation": "Conference Center, Room 101",
  "menuItems": [
    {
      "id": "uuid-menuitem-1",
      "foodItemId": "uuid-food-1",
      "size": "Large",
      "price": 25.99,
      "sequence": 1
    }
  ]
}
```

---

### List Catering Menus
**Endpoint**: `GET /api/v1/catering-menus`  
**Role**: Admin, User (own menus)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "uuid-1234-5678",
      "name": "Corporate Event Menu",
      "eventDate": "2026-05-15T12:00:00Z"
    }
  ],
  "meta": {
    "total": 5,
    "totalCount": 5
  }
}
```

---

### Update Catering Menu
**Endpoint**: `PUT /api/v1/catering-menus/{id}`  
**Role**: Admin only

**Request Body**:
```json
{
  "name": "Updated Corporate Event",
  "description": "Updated description",
  "eventDate": "2026-05-20T12:00:00Z",
  "eventLocation": "Conference Center, Room 202"
}
```

**Response (200 OK)**: Success

---

### Delete Catering Menu
**Endpoint**: `DELETE /api/v1/catering-menus/{id}`  
**Role**: Admin only

**Response (204 No Content)**: Success

---

## Menu Items API

### Add Menu Item to Menu
**Endpoint**: `POST /api/v1/menus/{menuType}/{menuId}/menu-items`  
**Role**: Admin only  
**Path Parameters**:
- `menuType`: `weekly` or `catering`
- `menuId`: UUID of the menu

**Request Body**:
```json
{
  "foodItemId": "uuid-food-1234",
  "size": "Regular",
  "price": 12.99,
  "sequence": 1
}
```

**Response (201 Created)**:
```json
{
  "id": "uuid-menuitem-1234",
  "tenantId": "tenant-uuid-1234",
  "menuId": "uuid-1234-5678",
  "menuType": "weekly",
  "foodItemId": "uuid-food-1234",
  "size": "Regular",
  "price": 12.99,
  "sequence": 1,
  "createdAt": "2026-05-08T10:30:00Z",
  "updatedAt": "2026-05-08T10:30:00Z"
}
```

---

### List Menu Items in Menu
**Endpoint**: `GET /api/v1/menus/{menuType}/{menuId}/menu-items`  
**Role**: Admin, User (own menus)  
**Path Parameters**:
- `menuType`: `weekly` or `catering`
- `menuId`: UUID of the menu

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "uuid-menuitem-1234",
      "foodItemId": "uuid-food-1234",
      "size": "Regular",
      "price": 12.99,
      "sequence": 1
    }
  ]
}
```

---

### Update Menu Item
**Endpoint**: `PUT /api/v1/menus/{menuType}/{menuId}/menu-items/{menuItemId}`  
**Role**: Admin only

**Request Body**:
```json
{
  "size": "Large",
  "price": 15.99,
  "sequence": 2
}
```

**Response (200 OK)**: Success

---

### Delete Menu Item
**Endpoint**: `DELETE /api/v1/menus/{menuType}/{menuId}/menu-items/{menuItemId}`  
**Role**: Admin only

**Response (204 No Content)**: Success

**Error Responses**:
- `404 Not Found`: Menu item not found
- `409 Conflict`: Menu item is referenced by orders

---

## Orders API

### Create Order
**Endpoint**: `POST /api/v1/orders`  
**Role**: User only  
**Headers**: 
- `Authorization`: Bearer <token>
- `X-Tenant-ID`: <tenant-id> (derived from user)

**Request Body**:
```json
{
  "menuType": "weekly",
  "menuId": "uuid-1234-5678",
  "items": [
    {
      "menuItemId": "uuid-menuitem-1",
      "quantity": 2
    },
    {
      "menuItemId": "uuid-menuitem-2",
      "quantity": 1
    }
  ]
}
```

**Response (201 Created)**:
```json
{
  "id": "uuid-order-1234",
  "tenantId": "tenant-uuid-1234",
  "userId": "uuid-user-5678",
  "menuType": "weekly",
  "menuId": "uuid-1234-5678",
  "status": "pending",
  "orderDate": "2026-05-08T10:30:00Z",
  "deliveryDate": "2026-05-11T12:00:00Z",
  "totalPrice": 38.97,
  "items": [
    {
      "menuItemId": "uuid-menuitem-1",
      "quantity": 2,
      "price": 12.99
    }
  ],
  "createdAt": "2026-05-08T10:30:00Z",
  "updatedAt": "2026-05-08T10:30:00Z"
}
```

---

### Get Order by ID
**Endpoint**: `GET /api/v1/orders/{id}`  
**Role**: User (own orders), Admin (all orders)

**Response (200 OK)**:
```json
{
  "id": "uuid-order-1234",
  "tenantId": "tenant-uuid-1234",
  "userId": "uuid-user-5678",
  "menuType": "weekly",
  "menuId": "uuid-1234-5678",
  "status": "ready",
  "orderDate": "2026-05-08T10:30:00Z",
  "deliveryDate": "2026-05-11T12:00:00Z",
  "totalPrice": 38.97,
  "items": [
    {
      "menuItemId": "uuid-menuitem-1",
      "quantity": 2,
      "price": 12.99
    }
  ]
}
```

---

### List Orders
**Endpoint**: `GET /api/v1/orders`  
**Role**: User (own orders), Admin (all orders)  
**Query Parameters**:
- `status`: Filter by status (pending, preparing, ready, delivered) (optional)
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "uuid-order-1234",
      "menuType": "weekly",
      "status": "ready",
      "orderDate": "2026-05-08T10:30:00Z",
      "deliveryDate": "2026-05-11T12:00:00Z",
      "totalPrice": 38.97
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 25,
    "totalPages": 3
  }
}
```

---

### Update Order Status
**Endpoint**: `PATCH /api/v1/orders/{id}/status`  
**Role**: Admin only

**Request Body**:
```json
{
  "status": "preparing"
}
```

**Response (200 OK)**:
```json
{
  "id": "uuid-order-1234",
  "status": "preparing",
  "updatedAt": "2026-05-08T11:00:00Z"
}
```

---

### Delete Order
**Endpoint**: `DELETE /api/v1/orders/{id}`  
**Role**: Admin only

**Response (204 No Content)**: Success

**Error Responses**:
- `404 Not Found`: Order not found
- `409 Conflict`: Order cannot be deleted (referenced by notifications)

---

## Notifications API

### Create Notification
**Endpoint**: `POST /api/v1/notifications`  
**Role**: Admin only  
**Headers**: 
- `Authorization`: Bearer <token>
- `X-Tenant-ID`: <tenant-id>

**Request Body**:
```json
{
  "userId": "uuid-user-5678",
  "orderId": "uuid-order-1234",
  "type": "order_ready",
  "message": "Your order is ready for pickup!"
}
```

**Response (201 Created)**:
```json
{
  "id": "uuid-notification-1234",
  "tenantId": "tenant-uuid-1234",
  "userId": "uuid-user-5678",
  "orderId": "uuid-order-1234",
  "type": "order_ready",
  "message": "Your order is ready for pickup!",
  "isRead": false,
  "createdAt": "2026-05-08T11:00:00Z"
}
```

---

### Get Notification by ID
**Endpoint**: `GET /api/v1/notifications/{id}`  
**Role**: User (own notifications), Admin (all notifications)

**Response (200 OK)**:
```json
{
  "id": "uuid-notification-1234",
  "tenantId": "tenant-uuid-1234",
  "userId": "uuid-user-5678",
  "orderId": "uuid-order-1234",
  "type": "order_ready",
  "message": "Your order is ready for pickup!",
  "isRead": true,
  "createdAt": "2026-05-08T11:00:00Z"
}
```

---

### List Notifications
**Endpoint**: `GET /api/v1/notifications`  
**Role**: User (own notifications), Admin (all notifications)  
**Query Parameters**:
- `type`: Filter by notification type (optional)
- `isRead`: Filter by read status (true/false) (optional)
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 10, max: 100)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "uuid-notification-1234",
      "type": "order_ready",
      "message": "Your order is ready for pickup!",
      "isRead": true,
      "createdAt": "2026-05-08T11:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 15,
    "totalPages": 2
  }
}
```

---

### Mark Notification as Read
**Endpoint**: `PATCH /api/v1/notifications/{id}/read`  
**Role**: User (own notifications)

**Response (200 OK)**:
```json
{
  "id": "uuid-notification-1234",
  "isRead": true,
  "updatedAt": "2026-05-08T12:00:00Z"
}
```

---

### Delete Notification
**Endpoint**: `DELETE /api/v1/notifications/{id}`  
**Role**: User (own notifications), Admin (all notifications)

**Response (204 No Content)**: Success

---

## Error Response Format

All endpoints follow a consistent error response format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "Field name that caused the error"
    }
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `INVALID_AUTH_TOKEN` | 401 | Authentication token is missing or invalid |
| `UNAUTHORIZED` | 401 | User is not authenticated |
| `FORBIDDEN` | 403 | User does not have permission to perform this action |
| `NOT_FOUND` | 404 | Resource not found |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `CONFLICT` | 409 | Resource conflict (e.g., item already exists) |
| `INTERNAL_SERVER_ERROR` | 500 | Internal server error |

---

## API Versioning Strategy

- Current version: `/api/v1`
- Version headers: `API-Version: v1`
- Breaking changes will trigger new major versions (e.g., `/api/v2`)

---

## Authentication & Authorization

### JWT Token Structure
```json
{
  "sub": "user-id",
  "tenantId": "tenant-id",
  "role": "admin|user",
  "exp": "expiration-time",
  "iat": "issued-at"
}
```

### Role-Based Access Control

| Role | Permissions |
|------|-------------|
| `admin` | Create/update/delete food items, menus, orders, notifications |
| `user` | View menus, place orders, view own notifications |

---

## Rate Limiting

- Default limit: 100 requests per minute per tenant
- Admin endpoints: 500 requests per minute
- User endpoints: 100 requests per minute

Rate limit headers in responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: Wed, 08 May 2026 10:35:00 GMT
```

---

## Documentation Standards

This API specification follows REST best practices and is compatible with OpenAPI/Swagger documentation generation.

### Swagger/OpenAPI Integration
The API can be documented using OpenAPI 3.0 specification for interactive API testing and client code generation.

---

## Implementation Notes

1. **Tenant Isolation**: All admin operations require `X-Tenant-ID` header to ensure tenant data isolation.

2. **Menu Item Sequencing**: Menu items must maintain sequence order within each menu. The system automatically handles resequencing when items are updated/deleted.

3. **Order Validation**: Orders validate that menu items exist and are available before creation.

4. **Notification Triggers**: Notifications can be created programmatically or triggered by status changes (e.g., order status change from `pending` to `ready`).

5. **Price Calculations**: Order totals are calculated as the sum of (item price × quantity) for all items in the order.
