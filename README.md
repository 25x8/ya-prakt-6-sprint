# Gophermart Loyalty System

This is a loyalty system for the fictional Gophermart online store. It allows users to register, upload order numbers, and receive loyalty points for their orders.

## Features

- User registration and authentication
- Order upload and processing
- Balance management
- Withdrawal of loyalty points

## Requirements

- Go 1.19 or later
- PostgreSQL database
- Accrual system service (provided)

## Configuration

The application can be configured using environment variables or command-line flags:

- Server address: `RUN_ADDRESS` or `-a` flag (default: `:8080`)
- Database URI: `DATABASE_URI` or `-d` flag (required)
- Accrual system address: `ACCRUAL_SYSTEM_ADDRESS` or `-r` flag (required)

## Running the application

### 1. Start the accrual system

```bash
# For Linux
./cmd/accrual/accrual_linux_amd64

# For macOS (Intel)
./cmd/accrual/accrual_darwin_amd64

# For macOS (Apple Silicon)
./cmd/accrual/accrual_darwin_arm64

# For Windows
./cmd/accrual/accrual_windows_amd64
```

The accrual system will be available at `http://localhost:8080` by default.

### 2. Start the Gophermart service

```bash
go run cmd/gophermart/main.go -d "postgres://username:password@localhost:5432/gophermart?sslmode=disable" -r "http://localhost:8080"
```

Or using environment variables:

```bash
DATABASE_URI="postgres://username:password@localhost:5432/gophermart?sslmode=disable" \
ACCRUAL_SYSTEM_ADDRESS="http://localhost:8080" \
go run cmd/gophermart/main.go
```

## API Endpoints

### Authentication

- `POST /api/user/register` - Register a new user
- `POST /api/user/login` - Login with existing credentials

### Orders

- `POST /api/user/orders` - Upload a new order number
- `GET /api/user/orders` - Get a list of uploaded orders

### Balance

- `GET /api/user/balance` - Get current balance
- `POST /api/user/balance/withdraw` - Withdraw points
- `GET /api/user/withdrawals` - Get withdrawal history

## Development

### Building from source

```bash
go build -o gophermart cmd/gophermart/main.go
```

### Running tests

```bash
go test ./...
```
