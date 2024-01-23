FROM golang:1.21 as builder

WORKDIR /app

# Copy the go.mod and go.sum files first and download the dependencies.
# This is done separately from copying the entire source code to leverage Docker cache
# and avoid re-downloading dependencies if they haven't changed.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application's source code.
COPY . .

# Build the application. This assumes you have a main package at the root of your project.
# Adjust the path to the main package if it's located elsewhere.
RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/user-server/main ./cmd/user-server
RUN CGO_ENABLED=0 GOOS=linux go build -o ./build/migrator/main ./cmd/migrator

# Define environment variables for PostgreSQL and Redis connections.
# These values can be overridden when running the container.
ENV ENV="dev"\
    DATABASE_DSN="postgres://postgres:password@postgres:5432/userdb" \
    REDIS_URL="redis://redis"\
    GRPC_PORT=44044\
    GRPC_TIMEOUT=1h

# Expose the port your application listens on.
EXPOSE 44044

# Run the application.
CMD ["./build/migrator/main", "--postgres-url=postgres://postgres:password@postgres:5432/userdb?sslmode=disable", "--migration-path=./migrations"]
ENTRYPOINT ["./build/user-server/main"]


