# Food Offerings Catalog Specification

## Overview
The Food Offerings Catalog is a core component of the OpenSpec framework that allows administrators to create, manage, and categorize all food offerings within the system. Food items are categorized into two primary groups: **Meat** and **Non-Meat**.

---

## Catalog Structure

### 1. Food Offering Categories
Food offerings are divided into two main categories:

- **Meat**: Items containing meat products
- **Non-Meat**: Items that are vegetarian, vegan, or contain no meat products

Each food offering must be assigned to one of these categories during creation.

---

## Specifications

### 1. Food Offering Specification (`food_offerings_spec.yaml`)
```yaml
id: food_offerings_catalog
name: Food Offerings Catalog
description: Catalog of all food offerings categorized as meat and non-meat
version: 1.0
operations:
  create_offering:
    description: Create a new food offering
    inputs:
      - name: offering_name
        type: string
        required: true
      - name: category
        type: string
        enum: ["meat", "non_meat"]
        required: true
      - name: description
        type: string
        required: false
      - name: price
        type: number
        required: true
    outputs:
      - name: offering_id
        type: string
    error_codes:
      - 400: Invalid input
      - 409: Offering already exists

  list_offerings:
    description: List all food offerings
    inputs: []
    outputs:
      - name: offerings
        type: array
        items:
          type: object
          properties:
            offering_id:
              type: string
            name:
              type: string
            category:
              type: string
            description:
              type: string
            price:
              type: number
    error_codes:
      - 500: Internal server error

  get_offering:
    description: Get details of a specific food offering
    inputs:
      - name: offering_id
        type: string
        required: true
    outputs:
      - name: offering
        type: object
        properties:
          offering_id:
            type: string
          name:
            type: string
          category:
            type: string
          description:
            type: string
          price:
            type: number
    error_codes:
      - 404: Offering not found
      - 500: Internal server error
```

---

### 2. Admin Task Specification (`admin_food_offerings.yaml`)
```yaml
id: admin_food_offerings
name: Admin Food Offerings Management
description: Admin tasks for managing food offerings
version: 1.0
operations:
  create_offering:
    description: Create a new food offering
    inputs:
      - name: offering_data
        type: object
        required: true
        properties:
          name:
            type: string
          category:
            type: string
            enum: ["meat", "non_meat"]
          description:
            type: string
          price:
            type: number
    outputs:
      - name: success
        type: boolean
      - name: message
        type: string
    error_codes:
      - 400: Invalid input data
      - 409: Offering already exists

  update_offering:
    description: Update an existing food offering
    inputs:
      - name: offering_id
        type: string
        required: true
      - name: update_data
        type: object
        required: true
        properties:
          name:
            type: string
          description:
            type: string
          price:
            type: number
    outputs:
      - name: success
        type: boolean
      - name: message
        type: string
    error_codes:
      - 404: Offering not found
      - 400: Invalid input data

  delete_offering:
    description: Delete a food offering
    inputs:
      - name: offering_id
        type: string
        required: true
    outputs:
      - name: success
        type: boolean
      - name: message
        type: string
    error_codes:
      - 404: Offering not found
```

---

## Workflow Examples

### 1. Creating a New Food Offering
**Command for Meat Offering:**
```bash
openspec task run admin_food_offerings.create_offering \
  --inputs '{
    "offering_data": {
      "name": "Grilled Chicken",
      "category": "meat",
      "description": "Juicy grilled chicken breast with herbs",
      "price": 12.99
    }
  }'
```

**Command for Non-Meat Offering:**
```bash
openspec task run admin_food_offerings.create_offering \
  --inputs '{
    "offering_data": {
      "name": "Caprese Salad",
      "category": "non_meat",
      "description": "Fresh mozzarella, tomatoes, and basil",
      "price": 8.99
    }
  }'
```

---

### 2. Listing All Food Offerings
**Command:**
```bash
openspec task run food_offerings_catalog.list_offerings
```

---

### 3. Getting Details of a Specific Offering
**Command:**
```bash
openspec task run food_offerings_catalog.get_offering \
  --inputs '{"offering_id": "offering_123"}'
```

---

## Implementation Details

### Directory Structure
```
openspec/
├── specs/
│   ├── food_offerings/
│   │   └── food_offerings_spec.yaml
│   └── tasks/
│       └── admin_food_offerings.yaml
└── changes/
    └── implementations/
        └── food_offerings/
            ├── meat/
            └── non_meat/
```

---

### Example Implementation Files
- **Meat Offering Implementation**: `openspec/changes/implementations/food_offerings/meat/grilled_chicken_v1.py`
- **Non-Meat Offering Implementation**: `openspec/changes/implementations/food_offerings/non_meat/caprese_salad_v1.py`

---

## Monitoring and Logging

### Metrics
- **Offering Creation Rate**: Number of new offerings created per time period
- **Category Distribution**: Percentage of offerings in meat vs. non-meat categories
- **Update/Delete Operations**: Frequency of updates and deletions

### Audit Trail
- Version history of food offerings
- Execution logs for all admin operations
- Change tracking for category assignments