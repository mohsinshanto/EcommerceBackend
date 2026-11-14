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
