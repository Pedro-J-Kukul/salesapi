# 🚀 Sales API

A modern, production-ready RESTful API for managing sales, products, and users with role-based authentication, AI-powered chatbot assistant, and comprehensive CRUD operations.

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-316192?style=flat&logo=postgresql)](https://www.postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

[📺 Watch the demo video on YouTube](https://youtu.be/eFz3BC-CA6k)

---

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Quick Start](#-quick-start)
- [API Documentation](#-api-documentation)
- [Docker Deployment](#-docker-deployment)
- [Development](#-development)
- [Testing](#-testing)
- [Project Structure](#-project-structure)

---

## ✨ Features

### Core Functionality
- ✅ **User Management** - Registration, activation, authentication with JWT tokens
- ✅ **Product Management** - Full CRUD operations for products
- ✅ **Sales Tracking** - Create, update, and analyze sales data
- ✅ **AI Chatbot Assistant** - GitHub AI-powered sales assistant for business insights

### Security & Performance
- 🔐 **Role-Based Access Control** - Admin, Cashier, and Guest roles with granular permissions
- 🚦 **Rate Limiting** - Configurable request throttling
- 🛡️ **Authentication** - Secure token-based authentication
- 📧 **Email Notifications** - User activation and notification system
- 🔄 **CORS Support** - Configurable cross-origin resource sharing

### Developer Experience
- 🐳 **Docker Ready** - Full containerization with Docker Compose
- 📊 **Metrics & Monitoring** - Built-in /v1/metrics endpoint
- 🧪 **Comprehensive Tests** - Unit and validation tests
- 📝 **Clean Architecture** - Modular, maintainable codebase
- 🔧 **Easy Configuration** - Environment-based configuration

---

## 🛠 Tech Stack

- **Language**: Go 1.23+
- **Database**: PostgreSQL 16
- **Migration**: golang-migrate
- **Router**: httprouter
- **Containerization**: Docker & Docker Compose
- **AI Integration**: GitHub Models API

---

## 🚀 Quick Start

### Prerequisites

- Go 1.23 or higher
- Docker & Docker Compose (for containerized deployment)
- PostgreSQL 16 (if running locally without Docker)
- golang-migrate CLI tool

### Option 1: Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/yourusername/salesapi.git
cd salesapi

# Copy environment file
cp .env.example .env

# Edit .env with your configuration
nano .env

# Start the application with Docker Compose
make docker/up

# Run migrations
make docker/migrate/up

# View logs
make docker/logs
```

The API will be available at `http://localhost:4000`

### Option 2: Local Development

```bash
# Clone the repository
git clone https://github.com/yourusername/salesapi.git
cd salesapi

# Install dependencies
go mod download

# Copy environment file
cp .env.example .envrc

# Edit .envrc with your configuration
nano .envrc

# Load environment variables (using direnv or source)
source .envrc

# Run migrations
make migrate/up

# Start the application
make run
```

---

## 📚 API Documentation

### Base URL
```
http://localhost:4000/v1
```

### Authentication

Most endpoints require authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-token>
```

### API Endpoints

#### 🔐 Authentication

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/v1/users` | POST | Register new user | ❌ |
| `/v1/users/activate` | PUT | Activate user account | ❌ |
| `/v1/tokens/authentication` | POST | Login and get token | ❌ |
| `/v1/tokens/authentication` | DELETE | Logout | ✅ |

#### 👤 Users

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/v1/users/profile` | GET | Get current user info | Authenticated |
| `/v1/user` | GET | List all users | `users:view` |
| `/v1/user/:id` | GET | Get user by ID | `users:view` |
| `/v1/user/:id` | PUT | Update user | `users:update` |
| `/v1/user/:id` | DELETE | Delete user | `users:delete` |

#### 📦 Products

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/v1/products` | GET | List all products | `product:view` |
| `/v1/products/:id` | GET | Get product by ID | `product:view` |
| `/v1/products` | POST | Create product | `product:create` |
| `/v1/products/:id` | PUT | Update product | `product:update` |
| `/v1/products/:id` | DELETE | Delete product | `product:delete` |

#### 💰 Sales

| Endpoint | Method | Description | Permission |
|----------|--------|-------------|------------|
| `/v1/sales` | GET | List all sales | `sale:view` |
| `/v1/sales/:id` | GET | Get sale by ID | `sale:view` |
| `/v1/sales` | POST | Create sale | `sale:create` |
| `/v1/sales/:id` | PUT | Update sale | `sale:update` |
| `/v1/sales/:id` | DELETE | Delete sale | `sale:delete` |

#### 🤖 AI Chatbot

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/v1/chatbot` | POST | Query sales assistant | ✅ |

#### 📊 Monitoring

| Endpoint | Method | Description | Auth Required |
|----------|--------|-------------|---------------|
| `/v1/metrics` | GET | Application metrics | ❌ |

### Example Requests

#### Register a User

```bash
curl -X POST http://localhost:4000/v1/users \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "role": "cashier"
  }'
```

#### Activate User Account

```bash
curl -X PUT http://localhost:4000/v1/users/activate \
  -H "Content-Type: application/json" \
  -d '{
    "token": "YOUR_ACTIVATION_TOKEN"
  }'
```

#### Login

```bash
curl -X POST http://localhost:4000/v1/tokens/authentication \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

#### Create a Product

```bash
curl -X POST http://localhost:4000/v1/products \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop",
    "price": 999.99
  }'
```

#### Record a Sale

```bash
curl -X POST http://localhost:4000/v1/sales \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "product_id": 1,
    "quantity": 2
  }'
```

#### Query AI Chatbot

```bash
curl -X POST http://localhost:4000/v1/chatbot \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "What are our top selling products this month?"
  }'
```

---

## 🐳 Docker Deployment

### Available Make Commands

```bash
# Docker Operations
make docker/build          # Build Docker image
make docker/up             # Start containers
make docker/down           # Stop containers
make docker/logs           # View all logs
make docker/logs/api       # View API logs
make docker/logs/db        # View database logs
make docker/restart        # Restart containers
make docker/rebuild        # Rebuild and restart
make docker/ps             # List containers
make docker/exec/api       # Shell into API container
make docker/exec/db        # Connect to PostgreSQL
make docker/migrate/up     # Run migrations in Docker
make docker/clean          # Remove all Docker resources
```

### Docker Compose Services

- **postgres** - PostgreSQL 16 database on port 5432
- **api** - Sales API application on port 4000
- **migrate** - Migration tool (run manually)

---

## 💻 Development

### Local Setup

```bash
# Install dependencies
go mod download

# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run the application
make run

# Run with custom configuration
go run ./cmd/api -port=8080 -env=production
```

### Database Migrations

```bash
# Create a new migration
make migrate/create name=add_users_table

# Apply migrations
make migrate/up

# Rollback last migration
make migrate/down

# Check migration version
make migrate/version

# Fix dirty migration state
make migrate/fix
```

### Available Make Commands

```bash
make help              # Show all available commands
make run               # Run application locally
make build             # Build binary
make test              # Run tests
make test/cover        # Run tests with coverage
make audit             # Code quality checks
make vendor            # Vendor dependencies
```

---

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test/cover

# Run specific test
go test -v ./cmd/api/... -run TestUserValidation

# Run tests in Docker
docker-compose exec api go test ./...
```

### Test Coverage

The project includes comprehensive tests for:
- ✅ User validation (email, password, roles)
- ✅ Product validation (name, price)
- ✅ Sales validation (quantities, IDs)
- ✅ Chatbot message validation
- ✅ URL parameter parsing
- ✅ JSON payload validation

---

## 📁 Project Structure

```
salesapi/
├── cmd/
│   └── api/              # Application entry point
│       ├── main.go
│       ├── routes.go
│       ├── handlers/
│       └── *_test.go
├── internal/
│   ├── data/             # Data models and database logic
│   │   ├── users.go
│   │   ├── products.go
│   │   ├── sales.go
│   │   ├── chatbot.go
│   │   └── models.go
│   ├── validator/        # Input validation
│   └── mailer/           # Email functionality
├── migrations/           # Database migrations
├── docker-compose.yml    # Docker orchestration
├── Dockerfile           # Container image definition
├── Makefile             # Build automation
├── .env.example         # Environment template
└── README.md            # This file
```

---

## 🔒 Security

- Passwords are hashed using bcrypt
- JWT tokens for authentication
- Rate limiting to prevent abuse
- CORS configuration for web security
- Non-root Docker containers
- Environment-based secrets management

---

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

## 👏 Acknowledgments

- Built with [Go](https://golang.org)
- Uses [PostgreSQL](https://www.postgresql.org)
- AI powered by [GitHub Models](https://github.com/marketplace/models)
- Containerized with [Docker](https://www.docker.com)
