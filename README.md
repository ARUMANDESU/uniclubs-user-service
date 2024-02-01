# UMS User Service

## Overview
This service is part of the Univercity Club Management application, focusing on user management and authentication. It handles operations such as user registration, authentication, role management, and more.

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
The Uniclubs User Service requires a configuration file to specify various settings like database connections, and service-specific parameters. Depending on your environment (development, test, or production), different configurations may be needed.

#### Configuration Files
- `dev.yaml`: Contains configuration for the development environment.
- `test.yaml`: Used for the test environment.
- `local.yaml`: Configuration for local development.

#### Setting Up Configuration
1. Choose the appropriate configuration file based on your environment.
2. Update the file with your specific settings, such as database connection strings, port numbers, and any third-party service credentials.
3. Ensure the application has access to this configuration file at runtime, either by placing it in the expected directory or setting an environment variable to its path.

#### Example Configuration
Here's an example of what the configuration file might look like (refer to `dev.yaml`, `test.yaml`, or `local.yaml` for full details):

```yaml
# Example configuration snippet
env: "local"
database_dsn: "postgresql://<user>:<password>@localhost:5432/<database_name>"
redis_url: "redis://:@localhost:6379"
grpc:
  port: 44044
  timeout: 1h
```

### Running the Service
After setting up the database and configuring the service, you can run it as follows:
  ```bash
  go run cmd/user-server/main.go --config=<path to the config file>
  ```

