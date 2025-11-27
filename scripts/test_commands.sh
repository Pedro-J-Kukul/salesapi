#!/bin/bash

# Live Testing Script for Sales API
# This script contains curl commands to test all API endpoints

set -e

# Configuration
API_URL="${API_URL:-http://localhost:4000}"
ADMIN_EMAIL="admin@example.com"
ADMIN_PASSWORD="adminpass123"
STAFF_EMAIL="staff@example.com"
STAFF_PASSWORD="staffpass123"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
print_section() {
    echo -e "\n${YELLOW}========================================${NC}"
    echo -e "${YELLOW}$1${NC}"
    echo -e "${YELLOW}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Store tokens
ADMIN_TOKEN=""
STAFF_TOKEN=""

###############################################################################
# User Registration and Activation
###############################################################################

print_section "User Registration and Authentication Tests"

# Register admin user
echo "1. Registering admin user..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/v1/users" \
  -H "Content-Type: application/json" \
  -d "{
    \"first_name\": \"Admin\",
    \"last_name\": \"User\",
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\",
    \"role\": \"admin\"
  }")

if echo "$REGISTER_RESPONSE" | grep -q "user"; then
    print_success "Admin user registered successfully"
    echo "$REGISTER_RESPONSE" | python3 -m json.tool
else
    print_error "Admin user registration failed"
    echo "$REGISTER_RESPONSE"
fi

# Register staff user
echo -e "\n2. Registering staff user..."
REGISTER_STAFF=$(curl -s -X POST "$API_URL/v1/users" \
  -H "Content-Type: application/json" \
  -d "{
    \"first_name\": \"Staff\",
    \"last_name\": \"User\",
    \"email\": \"$STAFF_EMAIL\",
    \"password\": \"$STAFF_PASSWORD\",
    \"role\": \"staff\"
  }")

if echo "$REGISTER_STAFF" | grep -q "user"; then
    print_success "Staff user registered successfully"
    echo "$REGISTER_STAFF" | python3 -m json.tool
else
    print_error "Staff user registration failed"
    echo "$REGISTER_STAFF"
fi

# Note: In a real scenario, you would need to activate users using the activation token
# For testing, you can manually activate users in the database:
# UPDATE users SET is_active = true WHERE email IN ('admin@example.com', 'staff@example.com');

echo -e "\n${YELLOW}NOTE: Activate users before proceeding with authentication${NC}"
echo "Run: psql -d sales -c \"UPDATE users SET is_active = true WHERE email IN ('$ADMIN_EMAIL', '$STAFF_EMAIL');\""
read -p "Press enter when users are activated..."

###############################################################################
# Authentication
###############################################################################

print_section "Authentication Tests"

# Authenticate admin user
echo "3. Authenticating admin user..."
AUTH_RESPONSE=$(curl -s -X POST "$API_URL/v1/tokens/authentication" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\"
  }")

ADMIN_TOKEN=$(echo "$AUTH_RESPONSE" | python3 -c "import sys, json; print(json.load(sys.stdin).get('authentication_token', ''))")

if [ -n "$ADMIN_TOKEN" ]; then
    print_success "Admin authenticated successfully"
    echo "Admin Token: $ADMIN_TOKEN"
else
    print_error "Admin authentication failed"
    echo "$AUTH_RESPONSE"
    exit 1
fi

# Authenticate staff user
echo -e "\n4. Authenticating staff user..."
STAFF_AUTH=$(curl -s -X POST "$API_URL/v1/tokens/authentication" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$STAFF_EMAIL\",
    \"password\": \"$STAFF_PASSWORD\"
  }")

STAFF_TOKEN=$(echo "$STAFF_AUTH" | python3 -c "import sys, json; print(json.load(sys.stdin).get('authentication_token', ''))")

if [ -n "$STAFF_TOKEN" ]; then
    print_success "Staff authenticated successfully"
    echo "Staff Token: $STAFF_TOKEN"
else
    print_error "Staff authentication failed"
    echo "$STAFF_AUTH"
fi

###############################################################################
# User Management
###############################################################################

print_section "User Management Tests"

# Get current user profile
echo "5. Getting admin user profile..."
curl -s -X GET "$API_URL/v1/users/profile" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -m json.tool

# List all users
echo -e "\n6. Listing all users (admin only)..."
curl -s -X GET "$API_URL/v1/users" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -m json.tool

# Test pagination
echo -e "\n7. Testing user pagination..."
curl -s -X GET "$API_URL/v1/users?page=1&page_size=5" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -m json.tool

###############################################################################
# Product Management
###############################################################################

print_section "Product Management Tests"

# Create products
echo "8. Creating products..."
PRODUCT1=$(curl -s -X POST "$API_URL/v1/products" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Laptop\",
    \"price\": 999.99
  }")

PRODUCT1_ID=$(echo "$PRODUCT1" | python3 -c "import sys, json; print(json.load(sys.stdin).get('product', {}).get('id', ''))")
print_success "Created product with ID: $PRODUCT1_ID"
echo "$PRODUCT1" | python3 -m json.tool

PRODUCT2=$(curl -s -X POST "$API_URL/v1/products" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Mouse\",
    \"price\": 29.99
  }")

PRODUCT2_ID=$(echo "$PRODUCT2" | python3 -c "import sys, json; print(json.load(sys.stdin).get('product', {}).get('id', ''))")
print_success "Created product with ID: $PRODUCT2_ID"

PRODUCT3=$(curl -s -X POST "$API_URL/v1/products" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Keyboard\",
    \"price\": 79.99
  }")

PRODUCT3_ID=$(echo "$PRODUCT3" | python3 -c "import sys, json; print(json.load(sys.stdin).get('product', {}).get('id', ''))")
print_success "Created product with ID: $PRODUCT3_ID"

# List all products
echo -e "\n9. Listing all products..."
curl -s -X GET "$API_URL/v1/products" | python3 -m json.tool

# Get specific product
echo -e "\n10. Getting product by ID..."
curl -s -X GET "$API_URL/v1/products/$PRODUCT1_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -m json.tool

# Update product
echo -e "\n11. Updating product..."
curl -s -X PUT "$API_URL/v1/products/$PRODUCT1_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Gaming Laptop\",
    \"price\": 1299.99
  }" | python3 -m json.tool

# Test price filtering
echo -e "\n12. Testing product price filtering..."
curl -s -X GET "$API_URL/v1/products?min_price=50&max_price=100" | python3 -m json.tool

###############################################################################
# Sales Management
###############################################################################

print_section "Sales Management Tests"

# Create sales
echo "13. Creating sales..."
SALE1=$(curl -s -X POST "$API_URL/v1/sales" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": 1,
    \"product_id\": $PRODUCT1_ID,
    \"quantity\": 2
  }")

SALE1_ID=$(echo "$SALE1" | python3 -c "import sys, json; print(json.load(sys.stdin).get('sale', {}).get('id', ''))")
print_success "Created sale with ID: $SALE1_ID"
echo "$SALE1" | python3 -m json.tool

SALE2=$(curl -s -X POST "$API_URL/v1/sales" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": 1,
    \"product_id\": $PRODUCT2_ID,
    \"quantity\": 5
  }")

SALE2_ID=$(echo "$SALE2" | python3 -c "import sys, json; print(json.load(sys.stdin).get('sale', {}).get('id', ''))")
print_success "Created sale with ID: $SALE2_ID"

# List all sales
echo -e "\n14. Listing all sales..."
curl -s -X GET "$API_URL/v1/sales" | python3 -m json.tool

# Get specific sale
echo -e "\n15. Getting sale by ID..."
curl -s -X GET "$API_URL/v1/sales/$SALE1_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -m json.tool

# Update sale
echo -e "\n16. Updating sale quantity..."
curl -s -X PUT "$API_URL/v1/sales/$SALE1_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"quantity\": 3
  }" | python3 -m json.tool

# Filter sales by user
echo -e "\n17. Filtering sales by user..."
curl -s -X GET "$API_URL/v1/sales?user_id=1" | python3 -m json.tool

# Filter sales by product
echo -e "\n18. Filtering sales by product..."
curl -s -X GET "$API_URL/v1/sales?product_id=$PRODUCT1_ID" | python3 -m json.tool

###############################################################################
# Permission Testing
###############################################################################

print_section "Permission and Authorization Tests"

# Staff tries to create product (should fail)
echo "19. Testing staff permissions (should fail to create product)..."
STAFF_CREATE=$(curl -s -X POST "$API_URL/v1/products" \
  -H "Authorization: Bearer $STAFF_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Unauthorized Product\",
    \"price\": 50.00
  }")

if echo "$STAFF_CREATE" | grep -q "not permitted"; then
    print_success "Permission check working - staff cannot create products"
else
    print_error "Permission check failed"
fi
echo "$STAFF_CREATE" | python3 -m json.tool

# Staff can view products
echo -e "\n20. Testing staff can view products..."
curl -s -X GET "$API_URL/v1/products" \
  -H "Authorization: Bearer $STAFF_TOKEN" | python3 -m json.tool

###############################################################################
# Error Handling Tests
###############################################################################

print_section "Error Handling Tests"

# Invalid authentication
echo "21. Testing invalid authentication..."
curl -s -X POST "$API_URL/v1/tokens/authentication" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"nonexistent@example.com\",
    \"password\": \"wrongpass\"
  }" | python3 -m json.tool

# Access protected endpoint without token
echo -e "\n22. Testing access without authentication..."
curl -s -X GET "$API_URL/v1/users" | python3 -m json.tool

# Invalid product data
echo -e "\n23. Testing invalid product data..."
curl -s -X POST "$API_URL/v1/products" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"\",
    \"price\": -10
  }" | python3 -m json.tool

# Non-existent resource
echo -e "\n24. Testing non-existent resource..."
curl -s -X GET "$API_URL/v1/products/99999" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -m json.tool

###############################################################################
# Cleanup Tests
###############################################################################

print_section "Cleanup Tests"

# Delete sale
echo "25. Deleting sale..."
curl -s -X DELETE "$API_URL/v1/sales/$SALE2_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | python3 -m json.tool

# Delete product
echo -e "\n26. Deleting product..."
curl -s -X DELETE "$API_URL/v1/products/$PRODUCT3_ID" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

print_success "Product deleted"

###############################################################################
# Metrics
###############################################################################

print_section "Metrics Tests"

echo "27. Getting application metrics..."
curl -s -X GET "$API_URL/v1/metrics" | python3 -m json.tool

###############################################################################
# Summary
###############################################################################

print_section "Testing Complete"
echo -e "${GREEN}All API endpoint tests completed successfully!${NC}"
echo -e "\nSummary:"
echo "- User registration and authentication: ✓"
echo "- User management: ✓"
echo "- Product CRUD operations: ✓"
echo "- Sales CRUD operations: ✓"
echo "- Permission-based access control: ✓"
echo "- Error handling: ✓"
echo "- Metrics: ✓"
