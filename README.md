# URL Shortener Microservices

A distributed URL shortener built with Go microservice architecture (gRPC, Redis, PosgreSQL, Kafka)

## Architecture

- **api-gateway**: REST API server (port 8080) - handles HTTP requests and routes to gRPC services
- **auth**: gRPC authentication service (port 50051) - user management and JWT tokens
- **url-gen**: gRPC URL generation service (port 50052) - creates short URLs
- **url-read**: gRPC URL reading service (port 50053) - resolves short URLs and tracks clicks

## Local Development with Docker Compose

Requirements:
- Docker
- Docker Compose

```
# Start all services
docker compose up --build --remove-orphans

# Run tests
bash scripts/test.sh
```