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
