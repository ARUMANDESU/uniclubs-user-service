FROM golang:1.22 as builder

WORKDIR /app

# Copy the go.mod and go.sum files first and download the dependencies.
# This is done separately from copying the entire source code to leverage Docker cache
# and avoid re-downloading dependencies if they haven't changed.
COPY go.mod go.sum ./
RUN go mod download
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
RUN apt-get update && apt-get install -y libvips-dev
# Copy the rest of the application's source code.
COPY . .

# Build the application. This assumes you have a main package at the root of your project.
# Adjust the path to the main package if it's located elsewhere.
RUN CGO_ENABLED=1 GOOS=linux go build -o ./build/user-server/main ./cmd/user-server

# Define environment variables for PostgreSQL and Redis connections.
# These values can be overridden when running the container.
ENV ENV="dev"
ENV DATABASE_DSN="postgres://postgres:password@postgres:5432/userdb"
ENV REDIS_URL="redis://redis"
ENV GRPC_PORT=44044
ENV GRPC_TIMEOUT=1h
ENV RABBITMQ_USER="admin"
ENV RABBITMQ_PASSWORD="admin"
ENV RABBITMQ_HOST="localhost"
ENV RABBITMQ_PORT="5672"
ENV ACCESS_TOKEN_DURATION="15m"
ENV ACCESS_TOKEN_SECRET="some_hard_secret"
ENV REFRESH_TOKEN_SECRET="more_harder_secret"
ENV AWS_REGION="us-east-1"
ENV AWS_ACCESS_KEY_ID="access_key_id"
ENV AWS_SECRET_ACCESS_KEY="secret_access_key"
ENV AWS_S3_BUCKET="bucket_name"

# Expose the port your application listens on.
EXPOSE 44044

# Run the application.
ENTRYPOINT ["./build/user-server/main"]
CMD ["migrate", "-path", "./migrations", "-database", "$DATABASE_DSN?sslmode=disable", "up"]