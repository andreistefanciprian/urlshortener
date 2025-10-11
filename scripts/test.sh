#!/bin/bash
# Test script for the URL shortening service

test_create_short_url() {
    echo "========================================="
    echo "Testing API Gateway Create Short URL"
    echo "========================================="

    echo
    echo "Test 1: Create short URL with expiration"
    SHORT_URL=$(curl -s -X POST -H "Content-Type: application/json" -d '{"longUrl": "https://google.com", "expiresIn": 10}' http://l.it/create | jq -r '.short_url')
    echo "Generated short URL: $SHORT_URL"
    echo 
    
    echo
    echo  "Test 2: Create short URL without expiration"
    SHORT_URL_2=$(curl -s -X POST -H "Content-Type: application/json" -d '{"longUrl": "https://protobuf.dev/getting-started/gotutorial/", "expiresIn": 0}' http://l.it/create | jq -r '.short_url')
    echo "Generated short URL: $SHORT_URL_2"
    echo
    
    echo "========================================="
    echo "Create short URL tests completed"
    echo "========================================="
}

test_get_long_url() {
    echo "========================================="
    echo "Testing API Gateway Get Long URL"
    echo "========================================="

    echo
    echo "Test 1: Get long URL from short URL: $SHORT_URL"
    curl -X GET -I $SHORT_URL
    echo 
    
    echo
    echo "Test 2: Get long URL from short URL: $SHORT_URL_2"
    curl -X GET -I $SHORT_URL_2
    echo
    
    echo
    echo "Test 3: Get long URL with empty short code (should fail)"
    curl -X GET http://l.it
    echo
    
    echo
    echo "Test 4: Get long URL with non-existent short code"
    curl -X GET http://l.it/nonexistent
    echo
    
    echo "========================================="
    echo "Get Long URL tests completed"
    echo "========================================="
}

delete_short_url() {
    echo "========================================="
    echo "Testing API Gateway Delete Short URL"
    echo "========================================="

    echo
    echo "Test 1: Create a short URL for deletion testing"
    DELETE_SHORT_URL_CODE=$(curl -s -X POST -H "Content-Type: application/json" -d '{"longUrl": "https://google.com", "expiresIn": 10}' http://l.it/create | jq -r '.short_url')
    echo "Created short URL for deletion: $DELETE_SHORT_URL_CODE"
    echo
    
    echo "Test 2: Verify the short URL works before deletion"
    curl -X GET -I $DELETE_SHORT_URL_CODE
    echo
    
    echo "Test 3: Access the short URL again to confirm it's cached"
    curl -X GET -I $DELETE_SHORT_URL_CODE
    echo
    
    echo "Test 4: Delete the short URL"
    curl -X DELETE -I $DELETE_SHORT_URL_CODE
    echo
    
    echo "Test 5: Try to access the deleted URL (should fail)"
    curl -X GET -I $DELETE_SHORT_URL_CODE
    echo
    
    echo "Test 6: Try to delete the same URL again (should fail)"
    curl -X DELETE -I $DELETE_SHORT_URL_CODE
    echo
    
    echo "========================================="
    echo "Delete Short URL tests completed"
    echo "========================================="
}

check_metrics() {
    echo "========================================="
    echo "Checking Metrics"
    echo "========================================="
    
    # Fetch and display metrics related to URL shortening
    curl -s http://localhost:9090/metrics | grep -E 'url_total|urls_total' | grep -v "#"

    echo "========================================="
    echo "Metrics check completed"
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

get_all_urls() {
    echo "========================================="
    echo "Testing API Gateway Get All URLs"
    echo "========================================="

    echo
    echo "Test: Get all URLs"
    curl -X GET http://l.it/ | jq .
    echo
    
    echo "========================================="
    echo "Get All URLs test completed"
    echo "========================================="
}

# Run the test functions
test_create_short_url
test_get_long_url
delete_short_url
get_all_urls
check_db_table
check_metrics
