# URL Shortener Microservices

A distributed URL shortener built with Go microservice architecture (gRPC, Redis, PostgreSQL, Kafka).

This project was built following system design principles from this excellent YouTube video: <a href="https://www.youtube.com/watch?v=iUU4O1sWtJA&t=2815s" target="_blank">System Design Interview: URL Shortener (like bit.ly)</a>

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