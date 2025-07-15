#!/bin/bash

# Test script to verify Swagger endpoints are working
echo "Testing Flight Ticket Service Swagger endpoints..."

# Start the server in background
echo "Starting server..."
./server &
SERVER_PID=$!

# Wait for server to start
sleep 3

echo "Testing endpoints..."

# Test root endpoint
echo "1. Testing root endpoint..."
curl -s http://localhost:8080/ | jq .

# Test health endpoint
echo -e "\n2. Testing health endpoint..."
curl -s http://localhost:8080/health | jq .

# Test Swagger JSON endpoint
echo -e "\n3. Testing Swagger JSON endpoint..."
curl -s http://localhost:8080/swagger/doc.json | jq '.info.title'

# Test Swagger UI endpoint (just check if it returns HTML)
echo -e "\n4. Testing Swagger UI endpoint..."
SWAGGER_RESPONSE=$(curl -s http://localhost:8080/swagger/)
if [[ $SWAGGER_RESPONSE == *"Swagger UI"* ]]; then
    echo "✅ Swagger UI is accessible"
else
    echo "❌ Swagger UI not accessible"
fi

# Clean up
echo -e "\nStopping server..."
kill $SERVER_PID

echo "✅ All tests completed!"
echo "Access Swagger UI at: http://localhost:8080/swagger/"
