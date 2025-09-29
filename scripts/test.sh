#!/bin/bash
# Test script for the URL shortening service

test_create_short_url() {
    echo "========================================="
    echo "Testing API Gateway Create Short URL"
    echo "========================================="

    echo
    echo "Test 1: Create short URL with expiration"
    curl -X POST -H "Content-Type: application/json" -d '{"longUrl": "https://google.com", "expiresIn": 10}' http://localhost:8080/create
    echo 
    
    echo
    echo  "Test 2: Create short URL without expiration"
    curl -X POST -H "Content-Type: application/json" -d '{"longUrl": "https://protobuf.dev/getting-started/gotutorial/", "expiresIn": 0}' http://localhost:8080/create
    echo
    
    echo "========================================="
    echo "Tests completed"
    echo "========================================="
}

check_db_table() {
    echo "========================================="
    echo "Checking Database Table"
    echo "========================================="
    
    # Connect to the PostgreSQL database and check the short_links table
    docker run --rm --network host -e PGPASSWORD=Auth123 postgres:15-alpine psql -h localhost -p 5432 -U url_read_user -d urls -c "SELECT * FROM short_links;"
    
    echo "========================================="
    echo "Database check completed"
    echo "========================================="
}
# Run the test functions
test_create_short_url
check_db_table