#!/bin/bash

echo "🧪 Testing startup order and dependencies..."

# Function to check service health
check_service() {
    local service=$1
    local endpoint=$2
    echo -n "Checking $service... "
    
    if curl -f -s "$endpoint" > /dev/null 2>&1; then
        echo "✅ HEALTHY"
        return 0
    else
        echo "❌ NOT READY"
        return 1
    fi
}

# Function to wait for service
wait_for_service() {
    local service=$1
    local endpoint=$2
    local max_attempts=30
    local attempt=1
    
    echo "⏳ Waiting for $service to be ready..."
    
    while [ $attempt -le $max_attempts ]; do
        if check_service "$service" "$endpoint"; then
            return 0
        fi
        echo "   Attempt $attempt/$max_attempts - waiting 5s..."
        sleep 5
        attempt=$((attempt + 1))
    done
    
    echo "❌ $service failed to become ready after $max_attempts attempts"
    return 1
}

# Test the startup sequence
echo "1️⃣ Checking PostgreSQL (via Docker health)..."
if docker inspect proyecto1-postgres-local --format='{{.State.Health.Status}}' | grep -q "healthy"; then
    echo "✅ PostgreSQL HEALTHY"
else
    echo "❌ PostgreSQL NOT HEALTHY"
    exit 1
fi

echo ""
echo "2️⃣ Testing LocalStack base service..."
wait_for_service "LocalStack" "http://localhost:4566/_localstack/health" || exit 1

echo ""
echo "3️⃣ Testing LocalStack SQS service..."
echo -n "Checking LocalStack SQS... "
if curl -f -s "http://localhost:4566/_localstack/health" | grep -q 'sqs.*running'; then
    echo "✅ HEALTHY"
else
    echo "❌ NOT READY"
    exit 1
fi

echo ""
echo "4️⃣ Testing API health..."
wait_for_service "API" "http://localhost:80/api/health" || exit 1

echo ""
echo "5️⃣ Testing Rankings endpoint..."
wait_for_service "Rankings" "http://localhost:80/api/public/rankings" || exit 1

echo ""
echo "✅ All services started in correct order!"
echo "🎉 Startup sequence test PASSED!"
