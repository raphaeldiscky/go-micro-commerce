# Example API calls for testing the microservices through Traefik

# Health checks
echo "=== Health Checks ==="
curl -s http://localhost/api/v1/products/health | jq
curl -s http://localhost/api/v1/sellers/health | jq

# Product Service API calls
echo -e "\n=== Product Service ==="

# Create a product
echo "Creating a product..."
curl -X POST http://localhost/api/v1/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gaming Laptop",
    "description": "High-performance gaming laptop with RTX 4080",
    "price": 1999.99,
    "category": "Electronics",
    "seller_id": "550e8400-e29b-41d4-a716-446655440001"
  }' | jq

# Get all products
echo -e "\nGetting all products..."
curl -s http://localhost/api/v1/products | jq

# Seller Service API calls
echo -e "\n=== Seller Service ==="

# Create a seller
echo "Creating a seller..."
curl -X POST http://localhost/api/v1/sellers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "TechStore Inc",
    "email": "contact@techstore.com",
    "phone": "+1-555-0123",
    "address": "123 Tech Street, Silicon Valley, CA"
  }' | jq

# Get all sellers
echo -e "\nGetting all sellers..."
curl -s http://localhost/api/v1/sellers | jq

# Test Traefik routing and middleware
echo -e "\n=== Traefik Features Test ==="

# Test CORS headers
echo "Testing CORS headers..."
curl -I -X OPTIONS http://localhost/api/v1/products \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST"

# Test rate limiting (make multiple requests)
echo -e "\nTesting rate limiting..."
for i in {1..5}; do
  echo "Request $i:"
  curl -s -w "Status: %{http_code}, Time: %{time_total}s\n" \
    -o /dev/null http://localhost/api/v1/products/health
done

echo -e "\n=== Monitoring Endpoints ==="
echo "Traefik API: http://localhost:8080/api/rawdata"
echo "Prometheus: http://localhost:9091"
echo "Grafana: http://localhost:3001"
echo "Kafka UI: http://localhost:8081"
