# UMS User Service

## Overview
This service is part of the University Clubs Management application, focusing on user management and authentication.
It handles operations such as user registration, authentication, role management, and more.

## Features
- User Registration
- User Authentication
- Role-based Access Control
- Additional features include user profile management...

## Technologies Used
- Go
- gRPC
- PostgreSQL
- Docker
- Redis
- [Taskfile](https://taskfile.dev/)

## Getting Started
### Prerequisites
- Go version 1.21.4
- PostgreSQL 16
- Docker 4.26.1


### Installation
Clone the repository:
   ```bash
   git clone https://github.com/ARUMANDESU/uniclubs-user-service.git
   cd uniclubs-user-service
   go mod download
   ```
   
### Database Setup and Migrations
Before running the service, you need to set up the database and apply the necessary migrations.

1. **Database Setup**:
   - Ensure PostgreSQL is installed and running.
   - Create a new database for the service, for example, `uniclubs_user_service`.

2. **Applying Migrations**:
   - Install [migration tool](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
   - Run the migration command to set up the schema:
     ```bash
     migrate -path ./migrations -database postgresql://username:password@localhost:5432/uniclubs_user_service up
     ```
   - Replace `username` and `password` with your database credentials.

### Configuration
The Uniclubs User Service requires a configuration file to specify various settings like database connections,
and service-specific parameters. 
Depending on your environment (development, test, or production), different configurations may be needed.

#### Setting Up Configuration
```dotenv
# Example configuration snippet
ENV=dev
# Database Configuration
DATABASE_DSN=postgresql://postgres:password@172.17.0.1:5432/uniclubs_user_service
REDIS_URL=redis://:@localhost:6379
# Server Configuration
GRPC_PORT=44044
GRPC_TIMEOUT=1h
# RabbitMQ Configuration
RABBITMQ_USER=user
RABBITMQ_PASSWORD=password
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
# JWT Configuration
ACCESS_TOKEN_DURATION=15m
ACCESS_TOKEN_SECRET=hart_secret_key
REFRESH_TOKEN_SECRET=very_hard_secret_key
```

### Running the Service
After setting up the database and configuring the service, you can run it as follows:
```bash
go run cmd/user-server/main.go
```

Or use the provided Taskfile to run the service:
```bash
task run:enviroment
 ```
#or
```bash
task env
 ```

