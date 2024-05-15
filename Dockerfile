FROM golang:1.22 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy the rest of the application's source code.
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o ./build/user-server/main ./cmd/user-server

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

EXPOSE 44044

# Run the application.
ENTRYPOINT ["./build/user-server/main"]