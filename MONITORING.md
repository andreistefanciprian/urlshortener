# Monitoring & Observability

This document covers the monitoring and observability features of the URL Shortener microservices.

## Prometheus Metrics

The API Gateway includes built-in Prometheus metrics for comprehensive observability and monitoring.

### Metrics Endpoint

- **URL**: `http://localhost:9090/metrics`
- **Port**: 9090 (separate from API traffic on port 80/8080)
- **Format**: Prometheus exposition format
- **Security**: Isolated on separate port for monitoring-only access

### Available Metrics

#### `create_short_url_total`
**Type**: Counter  
**Description**: Tracks short URL creation requests by outcome status

**Labels**:
- `success` - URL successfully created and stored
- `error` - Internal server errors (gRPC failures, database issues)
- `invalid_url` - Invalid URL format or unsupported schemes
- `invalid_expiration` - Invalid expiration time (negative values)

#### `get_long_url_total`
**Type**: Counter  
**Description**: Tracks URL resolution requests by outcome status

**Labels**:
- `success` - URL successfully resolved and user redirected
- `error` - Internal server errors (gRPC failures, database issues)
- `not_found` - Short code doesn't exist in database
- `expired` - Short URL has passed its expiration time

### Example Metrics Output

```prometheus
# HELP create_short_url_total Total number of short URL creation requests
# TYPE create_short_url_total counter
create_short_url_total{status="success"} 42
create_short_url_total{status="invalid_url"} 3
create_short_url_total{status="error"} 1

# HELP get_long_url_total Total number of long URL retrieval requests  
# TYPE get_long_url_total counter
get_long_url_total{status="success"} 156
get_long_url_total{status="not_found"} 8
get_long_url_total{status="expired"} 2
get_long_url_total{status="error"} 1
```

## Monitoring Setup

### Local Development

```bash
# Filter for URL shortener specific metrics
curl -s http://localhost:9090/metrics | grep url
```

### Grafana Dashboards

#### Key Performance Indicators (KPIs)

1. **Request Rate**: Requests per second for both create and resolve operations
2. **Success Rate**: Percentage of successful operations vs errors
3. **Error Rate**: Rate of different error types (4xx vs 5xx)
4. **Expiration Tracking**: Rate of expired URL access attempts

#### Sample Queries

```promql
# Request rate (requests per second)
rate(create_short_url_total[5m])
rate(get_long_url_total[5m])

# Success rate percentage
(rate(create_short_url_total{status="success"}[5m]) / rate(create_short_url_total[5m])) * 100
(rate(get_long_url_total{status="success"}[5m]) / rate(get_long_url_total[5m])) * 100

# Error rate
rate(create_short_url_total{status="error"}[5m])
rate(get_long_url_total{status="error"}[5m])

# Not found rate (indicating potential scanning/guessing)
rate(get_long_url_total{status="not_found"}[5m])
```

## Alerting

### Recommended Alerts

#### High Error Rate
```yaml
- alert: HighErrorRate
  expr: rate(create_short_url_total{status="error"}[5m]) > 0.1
  for: 2m
  labels:
    severity: warning
  annotations:
    summary: "High error rate in URL creation"
```

#### Service Availability
```yaml
- alert: ServiceDown
  expr: up{job="url-shortener"} == 0
  for: 1m
  labels:
    severity: critical
  annotations:
    summary: "URL Shortener service is down"
```

#### High Not Found Rate (Potential Attack)
```yaml
- alert: HighNotFoundRate
  expr: rate(get_long_url_total{status="not_found"}[5m]) > 1.0
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High rate of not found requests - potential scanning"
```

## Logging

The application uses structured logging with different log levels:

### Log Levels
- **INFO**: Normal operations, successful requests
- **WARN**: Unusual but non-critical situations
- **ERROR**: Error conditions that need attention
- **FATAL**: Critical errors that cause service shutdown

### Environment Variables
```bash
LOG_LEVEL=info  # debug, info, warn, error, fatal
```

### Sample Log Entries

```json
{
  "level": "info",
  "msg": "Successfully processed CreateShortURL request for https://example.com",
  "time": "2025-10-08T00:45:15Z"
}

{
  "level": "error", 
  "msg": "gRPC call to GetLongURL failed: connection refused",
  "time": "2025-10-08T00:45:16Z"
}
```