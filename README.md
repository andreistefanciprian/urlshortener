# URL Shortener Microservices

A distributed URL shortener built with Go microservice architecture (gRPC, Redis, PostgreSQL, Kafka).

This project was built following system design principles from this excellent YouTube video: [System Design Interview: URL Shortener (like bit.ly)](https://www.youtube.com/watch?v=iUU4O1sWtJA&t=2815s)

## Architecture

- **api-gateway**: REST API server (port 8080) - handles HTTP requests and routes to gRPC services
- **auth**: gRPC authentication service (port 50051) - user management and JWT tokens
- **url-gen**: gRPC URL generation service (port 50052) - creates short URLs
- **url-read**: gRPC URL reading service (port 50053) - resolves short URLs and tracks clicks

## API Endpoints

The API Gateway exposes two main REST endpoints:

##### POST /create
Creates a new short URL from a long URL.

```json
# Request
{
  "longUrl": "https://example.com/very/long/url",
  "expiresIn": 7
}

# Response
{
  "shortUrl": "l.it/gt0PFD8",
  "expiration": "2025-10-10T12:00:00Z"
}
```

##### GET /{shortCode}
Retrieves and redirects to the original long URL.

```
# Request
GET /gt0PFD8

# Response
302 Found: Redirects to original URL
404 Not Found: Short code doesn't exist  
410 Gone: Short URL has expired
```

## GetLongURL Architecture Flow

Cache-first architecture for optimal performance:

![GetLongURL Flow](get_long_url_flow.png)

**Flow:** Client → API Gateway → URL-Read Service (Redis Cache → PostgreSQL if cache miss) → 302 Redirect

**Benefits:** Fast lookups via Redis cache, graceful fallback to database, scalable read throughput

## Local Development with Docker Compose

Requirements:
- Docker
- Docker Compose

```
# Start all services
make up

# Add domain mapping to resolve l.it to localhost (required for testing)
# Edit /etc/hosts file: sudo vim /etc/hosts
# Add this line: 127.0.0.1 l.it
# This allows generated URLs like http://l.it/TRDZip6 to work locally

# Run tests
bash scripts/test.sh

# Use curl to create short URL
curl -s -X POST -H "Content-Type: application/json" -d '{"longUrl": "https://protobuf.dev/getting-started/gotutorial/", "expiresIn": 1}' http://l.it/create

# Use curl to get Long URL
curl -I http://l.it/TRDZip6

# Tear down services
make down
```