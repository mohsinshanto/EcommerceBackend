# E-commerce Backend

A RESTful e-commerce backend built with **Go**, **Gin**, and **MySQL**. The project focuses on scalable product discovery, nested category management, authentication, cart management, and order processing.

## Tech Stack

- **Language:** Go
- **Web Framework:** Gin
- **ORM:** GORM
- **Database:** MySQL
- **Caching:** Redis
- **Search Engine:** Typesense
- **Authentication:** JWT
- **Payment Gateway:** SSLCommerz

## Features
## API Endpoints

Base URL: `http://localhost:8080/api`

### Authentication

| Method | Endpoint | Description | Access |
|---|---|---|---|
| POST | `/register` | Register a new user | Public |
| POST | `/login` | Authenticate user and obtain JWT | Public |

### Products

| Method | Endpoint | Description | Access |
|---|---|---|---|
| GET | `/products` | Get products with pagination, search, and filters | Public |
| GET | `/products/count` | Get total product count | Public |
| GET | `/products/:id` | Get product details | Public |
| POST | `/products` | Create a product | Admin |
| DELETE | `/products/:id` | Delete a product | Admin |

### Categories

| Method | Endpoint | Description | Access |
|---|---|---|---|
| GET | `/categories` | Get hierarchical category tree | Public |
| POST | `/categories` | Create a category | Admin |
| PUT | `/categories/:id` | Update a category | Admin |
| DELETE | `/categories/:id` | Delete a category | Admin |

### Cart

| Method | Endpoint | Description | Access |
|---|---|---|---|
| POST | `/cart` | Add product to cart | Authenticated |
| GET | `/cart` | Get user's cart | Authenticated |
| PUT | `/cart/:id` | Update cart quantity | Authenticated |
| DELETE | `/cart/:id` | Remove item from cart | Authenticated |

### Orders

| Method | Endpoint | Description | Access |
|---|---|---|---|
| POST | `/order` | Create an order | Authenticated |
| GET | `/orders` | Get user's orders | Authenticated |
| PATCH | `/orders/:id/archive` | Archive an order | Authenticated |
| GET | `/admin/orders` | Get all orders | Admin |

### SSLCommerz Payment Callbacks

| Method | Endpoint | Description | Access |
|---|---|---|---|
| POST | `/payments/sslcommerz/success` | Handle successful payment callback | Gateway |
| GET | `/payments/sslcommerz/success` | Handle successful payment callback | Gateway |
| POST | `/payments/sslcommerz/fail` | Handle failed payment callback | Gateway |
| GET | `/payments/sslcommerz/fail` | Handle failed payment callback | Gateway |
| POST | `/payments/sslcommerz/cancel` | Handle cancelled payment callback | Gateway |
| GET | `/payments/sslcommerz/cancel` | Handle cancelled payment callback | Gateway |
| POST | `/payments/sslcommerz/ipn` | Handle Instant Payment Notification | Gateway |

**Note:** Protected endpoints require a valid JWT. Admin endpoints require both authentication and admin authorization.
## Products API

### 1. Get All Products

Retrieve a paginated list of products with optional search, category filtering, price filtering, and sorting.

**Endpoint:** `GET /api/products`

**Query Parameters:**

| Parameter | Type | Description |
|---|---|---|
| `page` | int | Page number (default: 1) |
| `limit` | int | Number of products per page (default: 8) |
| `search` | string | Search products by keyword |
| `category` | uint | Filter by category ID (includes descendant categories) |
| `min_price` | float | Minimum product price |
| `max_price` | float | Maximum product price |
| `sort` | string | Sort products by price |

**Supported Sorting Values:**

- `price_asc` — Sort by price, lowest first.
- `price_desc` — Sort by price, highest first.

**Example Requests:**

```http
GET /api/products
GET /api/products?page=2&limit=8
GET /api/products?search=laptop
GET /api/products?category=2
GET /api/products?min_price=1000&max_price=50000
GET /api/products?sort=price_asc
GET /api/products?search=laptop&category=2&sort=price_desc
```

**Success Response (200 OK):**

```json
{
  "message": "Products loaded",
  "data": {
    "products": [
      {
        "id": 510033,
        "name": "Smart Ultrabook 489",
        "description": "Product description",
        "price": 75000,
        "stock": 20,
        "image_url": "https://example.com/image.jpg",
        "category": "Laptop",
        "category_id": 2,
        "category_info": {
          "id": 2,
          "name": "Laptop"
        }
      }
    ],
    "count": 510020,
    "search_count": 120,
    "category_counts": {
      "2": 120,
      "6": 80
    },
    "page": 1,
    "limit": 8,
    "has_next": true,
    "has_prev": false
  }
}
```

**Response Fields:**

| Field | Type | Description |
|---|---|---|
| `products` | array | List of product objects |
| `count` | int | Total active product count |
| `search_count` | int or null | Number of matching search results |
| `category_counts` | object | Product counts grouped by category ID |
| `page` | int | Current page |
| `limit` | int | Products per page |
| `has_next` | boolean | Whether another page exists |
| `has_prev` | boolean | Whether a previous page exists |

### 2. Get Product by ID

Retrieve a single product using its ID.

**Endpoint:** `GET /api/products/:id`

**Example Request:**

```http
GET /api/products/510033
```

**Success Response (200 OK):**

Returns a single product object with the same fields described in the Product Object Fields section.

**Error Responses:**

- `400 Bad Request` — Invalid product ID.
- `404 Not Found` — Product not found.

### 3. Get Product Count

Retrieve the total number of active products.

**Endpoint:** `GET /api/products/count`

### 4. Create Product (Admin Only)

Create a new product.

**Endpoint:** `POST /api/products`

**Authentication:** Required (Admin)

**Request Body:**

```json
{
  "name": "Wireless Headphones",
  "description": "Bluetooth wireless headphones",
  "price": 2500,
  "stock": 50,
  "image_url": "https://example.com/headphones.jpg",
  "category": "Audio"
}
```

All request fields are required.

### 5. Delete Product (Admin Only)

Delete a product by its ID.

**Endpoint:** `DELETE /api/products/:id`

**Authentication:** Required (Admin)

**Example Request:**

```http
DELETE /api/products/510033
```
## Categories API

The Categories API manages product categories using a hierarchical (parent-child) structure. It supports creating, retrieving, updating, and deleting categories.

### Base URL

```text
/api/categories
```

### Authentication

- `GET /api/categories` is publicly accessible.
- `POST`, `PUT`, and `DELETE` require JWT authentication and Admin privileges.

### API Endpoints

| Method | Endpoint | Access | Description |
|---|---|---|---|
| GET | `/api/categories` | Public | Retrieve all categories as a nested tree |
| POST | `/api/categories` | Admin | Create a new category |
| PUT | `/api/categories/:id` | Admin | Update an existing category |
| DELETE | `/api/categories/:id` | Admin | Delete a category |

---

### 1. Get Category Tree

**Endpoint:** `GET /api/categories`

**Access:** Public

Retrieves all categories and organizes them into a hierarchical tree using parent-child relationships.

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Categories loaded successfully",
  "data": [
    {
      "id": 1,
      "name": "Electronics",
      "parent_id": null,
      "children": [
        {
          "id": 2,
          "name": "Laptop",
          "parent_id": 1,
          "children": [
            {
              "id": 3,
              "name": "Dell",
              "parent_id": 2,
              "children": []
            },
            {
              "id": 4,
              "name": "HP",
              "parent_id": 2,
              "children": []
            }
          ]
        },
        {
          "id": 6,
          "name": "Mobile",
          "parent_id": 1,
          "children": []
        }
      ]
    }
  ]
}
```

**Error Response (500 Internal Server Error):**

```json
{
  "success": false,
  "error": "Failed to load categories"
}
```

---

### 2. Create Category

**Endpoint:** `POST /api/categories`

**Access:** Admin

Creates a new category. A category can be created as a root category or as a child of an existing category.

**Request Headers:**

```text
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request Body:**

```json
{
  "name": "Gaming Laptop",
  "parent_id": 2
}
```

To create a root category, set `parent_id` to `null` or omit it.

**Request Fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | Yes | Category name |
| `parent_id` | uint or null | No | ID of the parent category |

**Success Response (201 Created):**

```json
{
  "success": true,
  "message": "Category created successfully",
  "data": {
    "id": 9,
    "name": "Gaming Laptop",
    "parent_id": 2,
    "parent": null,
    "children": null
  }
}
```

**Error Response (400 Bad Request):**

Returned when the request body is invalid or the required `name` field is missing.

```json
{
  "success": false,
  "error": "error details"
}
```

**Error Response (500 Internal Server Error):**

Returned when category creation fails, including when the specified parent category does not exist.

```json
{
  "success": false,
  "error": "error details"
}
```

---

### 3. Update Category

**Endpoint:** `PUT /api/categories/:id`

**Access:** Admin

Updates the name and parent category of an existing category.

**Request Headers:**

```text
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

**Request Body:**

```json
{
  "name": "Gaming Laptops",
  "parent_id": 2
}
```

To make the category a root category, set `parent_id` to `null`.

**Request Fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `name` | string | Yes | Updated category name |
| `parent_id` | uint or null | No | ID of the new parent category |

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Category updated successfully"
}
```

**Error Responses:**

| Status Code | Condition |
|---|---|
| 400 Bad Request | Invalid category ID, invalid request body, or circular category reference |
| 404 Not Found | Category does not exist |
| 500 Internal Server Error | Existing circular reference, database error, or another internal error |

Example error response:

```json
{
  "success": false,
  "error": "error details"
}
```

---

### 4. Delete Category

**Endpoint:** `DELETE /api/categories/:id`

**Access:** Admin

Deletes an existing category.

A category cannot be deleted if it has child categories. Child categories must be deleted first.

**Request Headers:**

```text
Authorization: Bearer <JWT_TOKEN>
```

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Category deleted successfully"
}
```

**Error Responses:**

| Status Code | Condition |
|---|---|
| 400 Bad Request | Invalid category ID |
| 404 Not Found | Category does not exist |
| 409 Conflict | Category has child categories |
| 500 Internal Server Error | Database error or another internal error |

Example error response:

```json
{
  "success": false,
  "error": "error details"
}
```

---

### Category Validation Rules

- Category name is required when creating or updating a category.
- Category name has a maximum database column size of 100 characters.
- A parent category must exist before it can be assigned.
- A category cannot be assigned as its own parent.
- Circular parent-child relationships are not allowed.
- A category with child categories cannot be deleted.
- Admin privileges are required for category creation, updating, and deletion.
## Cart API

The Cart API allows authenticated users to add products to their shopping cart, retrieve cart items, update item quantities, and remove items.

### Base URL

```text
/api/cart
```

### Authentication

All Cart API endpoints require JWT authentication.

**Request Header:**

```text
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json
```

The cart is associated with the authenticated user. Users can only update or remove their own cart items.

### API Endpoints

| Method | Endpoint | Access | Description |
|---|---|---|---|
| POST | `/api/cart` | Authenticated | Add a product to the cart |
| GET | `/api/cart` | Authenticated | Retrieve the user's cart |
| PUT | `/api/cart/:id` | Authenticated | Update cart item quantity |
| DELETE | `/api/cart/:id` | Authenticated | Remove an item from the cart |

---

### 1. Add to Cart

**Endpoint:** `POST /api/cart`

**Access:** Authenticated

Adds a product to the authenticated user's cart.

If the product already exists in the cart, its quantity is increased instead of creating a duplicate cart item.

**Request Body:**

```json
{
  "product_id": 10,
  "quantity": 2
}
```

**Request Fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `product_id` | uint | Yes | ID of the product |
| `quantity` | integer | Yes | Quantity to add; minimum value is 1 |

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Item added to cart",
  "data": {
    "id": 5,
    "quantity": 2,
    "product": {
      "id": 10,
      "name": "Smartphone",
      "price": 25000,
      "image_url": "https://example.com/image.jpg"
    }
  }
}
```

**Error Responses:**

| Status Code | Condition |
|---|---|
| 400 Bad Request | Invalid request body, product not found, insufficient stock, or cart operation failed |
| 401 Unauthorized | User is not authenticated |

Example error response:

```json
{
  "success": false,
  "error": "error details"
}
```

**Business Logic:**

- Checks whether the requested product exists.
- Checks whether sufficient stock is available.
- If the product already exists in the cart, increases its quantity.
- Otherwise, creates a new cart item.
- Checks stock against the resulting total quantity when updating an existing item.

---

### 2. Get User Cart

**Endpoint:** `GET /api/cart`

**Access:** Authenticated

Retrieves all cart items belonging to the authenticated user.

The API also calculates the total cart price using:

```text
Total = Sum(Quantity × Product Price)
```

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Cart loaded",
  "data": {
    "cartItems": [
      {
        "id": 5,
        "quantity": 2,
        "product": {
          "id": 10,
          "name": "Smartphone",
          "price": 25000,
          "image_url": "https://example.com/image.jpg"
        }
      },
      {
        "id": 6,
        "quantity": 1,
        "product": {
          "id": 20,
          "name": "Wireless Headphone",
          "price": 3000,
          "image_url": "https://example.com/headphone.jpg"
        }
      }
    ],
    "Total": 53000
  }
}
```

**Note:** The current `CartResponse` struct has an incorrect JSON tag (`josn:"total"`). Therefore, Go's JSON encoder uses the exported field name `Total`.

After correcting the tag to `json:"total"`, the response key will be:

```json
"total": 53000
```

**Error Responses:**

| Status Code | Condition |
|---|---|
| 401 Unauthorized | User is not authenticated |
| 500 Internal Server Error | Failed to retrieve the cart |

Example error response:

```json
{
  "success": false,
  "error": "error details"
}
```

---

### 3. Update Cart Item Quantity

**Endpoint:** `PUT /api/cart/:id`

**Access:** Authenticated

Updates the quantity of an existing cart item.

The `id` in the URL represents the cart item ID, not the product ID.

**Request Example:**

```text
PUT /api/cart/5
```

**Request Body:**

```json
{
  "quantity": 3
}
```

**Request Fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `quantity` | integer | Yes | New quantity; minimum value is 1 |

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Cart updated"
}
```

**Error Responses:**

| Status Code | Condition |
|---|---|
| 400 Bad Request | Invalid cart item ID, invalid quantity, or cart item not found |
| 401 Unauthorized | User is not authenticated |
| 403 Forbidden | Cart item belongs to another user |
| 404 Not Found | Associated product does not exist |
| 409 Conflict | Insufficient product stock |
| 500 Internal Server Error | Failed to update the cart item or another internal error |

Example error response:

```json
{
  "success": false,
  "error": "error details"
}
```

**Business Logic:**

- Finds the cart item using its ID.
- Verifies that the cart item belongs to the authenticated user.
- Validates the requested quantity.
- Checks whether sufficient product stock is available.
- Updates the cart item quantity.

---

### 4. Remove Cart Item

**Endpoint:** `DELETE /api/cart/:id`

**Access:** Authenticated

Removes an item from the authenticated user's cart.

The `id` in the URL represents the cart item ID.

**Request Example:**

```text
DELETE /api/cart/5
```

**Success Response (200 OK):**

```json
{
  "success": true,
  "message": "Item removed from cart"
}
```

**Error Responses:**

| Status Code | Condition |
|---|---|
| 400 Bad Request | Invalid cart item ID or cart item not found |
| 401 Unauthorized | User is not authenticated |
| 403 Forbidden | Cart item belongs to another user |
| 500 Internal Server Error | Failed to remove the cart item or another internal error |

Example error response:

```json
{
  "success": false,
  "error": "error details"
}
```

**Business Logic:**

- Finds the cart item using its ID.
- Verifies that the cart item belongs to the authenticated user.
- Deletes the cart item from the database.

---

### Cart Business Rules

- Every cart operation requires authentication.
- Each cart item belongs to a specific user.
- Adding an existing product increases its quantity.
- Product stock is checked before adding or updating quantities.
- A user cannot update or remove another user's cart items.
- Cart total is calculated from product prices and quantities.
- Cart item quantity must be at least 1.
### Authentication & Authorization
- User registration and login
- JWT-based authentication
- Password hashing
- Role-based authorization for admin operations

### Product Management
- Product listing with pagination
- Product details
- Product search using Typesense
- Filtering by category and price
- Sorting products
- Category-based product counts using faceting

### Nested Category Management
- Hierarchical category structure
- Parent and child category relationships
- Recursive descendant category retrieval
- Filtering products by parent or nested categories

### Cart Management
- Add products to cart
- Update cart items
- Remove cart items

### Order & Payment Management
- Order creation and order item management
- Cash on Delivery (COD)
- SSLCommerz payment integration
- Payment success, failure, cancellation, and IPN handling

### Performance & Scalability
- Redis-based product count caching
- Typesense-powered product search and filtering
- Batch synchronization of products to Typesense

## Getting Started

### Prerequisites

Make sure you have the following installed:

- Go 1.24+
- MySQL
- Redis
- Typesense

### Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/mohsinshanto/EcommerceBackend.git
   cd EcommerceBackend
   ```

2. Create your environment file:

   ```bash
   cp .env.example .env
   ```

3. Update the `.env` file with your local database, Redis, Typesense, JWT, and payment gateway configuration.

4. Install dependencies:

   ```bash
   go mod download
   ```

5. Start the application:

   ```bash
   go run .
   ```

The backend runs on `http://localhost:8080` with the default local configuration.

## Environment Variables

See [`.env.example`](.env.example) for the required environment variable names and example values.

**Important:** Never commit your actual `.env` file or production credentials.

## Author

**Mohsin**

- GitHub: [mohsinshanto](https://github.com/mohsinshanto)
