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
