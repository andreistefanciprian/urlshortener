# URL Shortener Microservices

A distributed URL shortener built with Go microservice architecture (gRPC, Redis, PostgreSQL, Kafka).

This project was built following system design principles from this excellent YouTube video: [System Design Interview: URL Shortener (like bit.ly)](https://www.youtube.com/watch?v=iUU4O1sWtJA&t=2815s)

## Architecture

- **frontend**: Web UI service (port 8090) - Bootstrap based admin interface for managing URLs with create/delete functionality
- **api-gateway**: REST API server (port 8080) - handles HTTP requests and routes to gRPC services
- **url-gen**: gRPC URL generation service (port 50052) - creates short URLs
- **url-read**: gRPC URL reading service (port 50053) - resolves short URLs and tracks clicks
- **redis**: Cache layer for fast URL lookups and reduced database load
- **postgresql**: Primary database for persistent URL storage and metadata 

## API Endpoints

The API Gateway exposes three main REST endpoints:

### POST /create
Creates a new short URL from a long URL.

**Flow:** Client → API Gateway → URL-Gen Service → PostgreSQL → Return short URL

```json
{
  "longUrl": "https://example.com/very/long/url",
  "expiresIn": 7
}

{
  "shortUrl": "l.it/gt0PFD8",
  "expiration": "2025-10-10T12:00:00Z"
}
```

### GET /{shortCode}
Retrieves and redirects to the original long URL.

**Flow:** Client → API Gateway → URL-Read Service → Redis/PostgreSQL → 302 Redirect

```
# Request
GET /gt0PFD8

# Response
302 Found: Redirects to original URL
404 Not Found: Short code doesn't exist  
410 Gone: Short URL has expired
```

### DELETE /{shortCode}
Deletes a short URL and removes it from both database and cache.

**Flow:** Client → API Gateway → URL-Gen Service → Redis/PostgreSQL → Confirmation

```
# Request
DELETE /gt0PFD8

# Response
200 OK: Short URL deleted successfully
404 Not Found: Short code doesn't exist
500 Internal Server Error: Database/cache error
```

## Frontend Web Interface

The system includes a web-based admin interface for easy URL management:

**Access**: `http://localhost:8090`

## Monitoring & Metrics

The API Gateway includes built-in Prometheus metrics for observability. See [MONITORING.md](MONITORING.md) for detailed documentation.

**Quick Access**: Metrics available at `http://localhost:9090/metrics`

## GetLongURL Architecture Flow

Cache-first architecture for optimal performance:

![GetLongURL Flow](get_long_url_flow.png)

**Flow:** Client → API Gateway → URL-Read Service (Redis Cache HIT → 302 Redirect | Cache MISS → PostgreSQL → 302 Redirect)

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

# Access the web interface
open http://localhost:8090

# Run tests
bash scripts/test.sh

# Use curl to create short URL
curl -s -X POST -H "Content-Type: application/json" -d '{"longUrl": "https://protobuf.dev/getting-started/gotutorial/", "expiresIn": 1}' http://l.it/create

# Use curl to get Long URL
curl -X GET -I http://l.it/TRDZip6

# Use curl to delete Short URL
curl -X DELETE -I http://l.it/TRDZip6

# View Prometheus metrics
curl http://localhost:9090/metrics | grep url_total

# Tear down services
make down
```