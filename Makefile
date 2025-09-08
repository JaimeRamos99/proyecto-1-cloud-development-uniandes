# Proyecto_1 Makefile

.PHONY: help local clean build logs health test-api nginx-logs api-logs db-logs localstack-logs rebuild-api stop localstack-health localstack-test storage-test

# Default target
help:
	@echo "Available commands:"
	@echo "  local           - Start full local environment (nginx + API + PostgreSQL + LocalStack)"
	@echo "  build           - Build the Go application"
	@echo "  rebuild-api     - Rebuild and restart only the API container"
	@echo "  logs            - Show logs from all containers"
	@echo "  nginx-logs      - Show logs from nginx container only"
	@echo "  api-logs        - Show logs from API container only"
	@echo "  db-logs         - Show logs from PostgreSQL container only"
	@echo "  localstack-logs - Show logs from LocalStack container only"
	@echo "  health          - Check health status of all services"
	@echo "  localstack-health - Check LocalStack services health (S3, SQS)"
	@echo "  localstack-test - Test LocalStack setup (create/list S3 buckets, SQS queues)"
	@echo "  storage-test    - Test file storage functionality with ObjectStorage"
	@echo "  test-api        - Test API endpoints with sample requests"
	@echo "  stop            - Stop all containers (keeps volumes)"
	@echo "  clean           - NUCLEAR: Stop and remove EVERYTHING (containers, volumes, networks, images)"

# Local: Full local environment with nginx proxy and LocalStack
local:
	@echo "🚀 Starting full local environment with nginx proxy and LocalStack..."
	docker-compose -f docker-compose.local.yml up -d
	@echo ""
	@echo "✅ Services are starting up..."
	@echo "🌐 nginx proxy:      http://localhost:80"
	@echo "🔗 API endpoints:    http://localhost:80/api"
	@echo "🗄️  PostgreSQL:      localhost:5432"
	@echo "☁️  LocalStack:       http://localhost:4566"
	@echo "📋 Health check:     http://localhost:80/nginx-health"
	@echo "📊 LocalStack UI:    http://localhost:4566/_localstack/cockpit (Pro) or use LocalStack Desktop"
	@echo ""
	@echo "LocalStack Services (auto-created):"
	@echo "  📁 S3 Bucket:       proyecto1-videos"
	@echo "  📨 SQS Queue:       proyecto1-video-processing"
	@echo "  💀 SQS DLQ:         proyecto1-video-processing-dlq"
	@echo "  🔍 Health:          http://localhost:4566/_localstack/health"
	@echo ""
	@echo "Available API endpoints:"
	@echo "  POST http://localhost:80/api/auth/signup"
	@echo "  POST http://localhost:80/api/auth/login"
	@echo "  POST http://localhost:80/api/auth/logout"
	@echo "  POST http://localhost:80/api/videos/upload (requires JWT)"
	@echo ""
	@echo "📝 Next steps:"
	@echo "  1. Wait ~60s for all services to initialize"
	@echo "  2. Run 'make health' to verify all services"
	@echo "  3. Run 'make localstack-test' to test S3/SQS"
	@echo "  4. Run 'make test-api' to test API endpoints"

# Build the Go application
build:
	@echo "🔨 Building Go application..."
	cd backend && go mod tidy
	cd backend && go build -o bin/api cmd/api/main.go
	@echo "✅ Binary created at backend/bin/api"

# Rebuild and restart only the API container
rebuild-api:
	@echo "🔄 Rebuilding API container..."
	docker-compose -f docker-compose.local.yml build api
	docker-compose -f docker-compose.local.yml up -d api
	@echo "✅ API container rebuilt and restarted"
	@echo "⏳ Waiting 10s for container to stabilize..."
	@sleep 10
	@make api-logs

# Show all logs
logs:
	docker-compose -f docker-compose.local.yml logs -f

# Show nginx logs only
nginx-logs:
	@echo "📋 nginx proxy logs:"
	docker logs -f proyecto1-nginx-local

# Show API logs only
api-logs:
	@echo "🚀 API logs:"
	docker logs -f proyecto1-api-local

# Show database logs only
db-logs:
	@echo "🗄️ PostgreSQL logs:"
	docker logs -f proyecto1-postgres-local

# Show LocalStack logs only
localstack-logs:
	@echo "☁️ LocalStack logs:"
	docker logs -f proyecto1-localstack-local

# Check health status of all services
health:
	@echo "🏥 Checking service health..."
	@echo ""
	@echo "🔍 nginx health check:"
	@curl -f http://localhost:80/nginx-health 2>/dev/null && echo " ✅ nginx OK" || echo " ❌ nginx FAIL"
	@echo ""
	@echo "🔍 PostgreSQL connection:"
	@docker exec proyecto1-postgres-local pg_isready -U postgres -d proyecto_1 >/dev/null 2>&1 && echo " ✅ PostgreSQL OK" || echo " ❌ PostgreSQL FAIL"
	@echo ""
	@echo "🔍 LocalStack health check:"
	@curl -s http://localhost:4566/_localstack/health >/dev/null 2>&1 && echo " ✅ LocalStack OK" || echo " ❌ LocalStack FAIL"
	@echo ""
	@echo "🔍 LocalStack S3 service:"
	@curl -s http://localhost:4566/_localstack/health | grep -q '"s3": "available"' && echo " ✅ S3 Service OK" || echo " ❌ S3 Service FAIL"
	@echo ""
	@echo "🔍 LocalStack SQS service:"
	@curl -s http://localhost:4566/_localstack/health | grep -q '"sqs": "available"' && echo " ✅ SQS Service OK" || echo " ❌ SQS Service FAIL"
	@echo ""
	@echo "🔍 API container status:"
	@docker ps --filter name=proyecto1-api-local --format "table {{.Names}}\t{{.Status}}" | tail -n +2 | while read line; do echo " $$line"; done
	@echo ""
	@echo "🔍 All containers:"
	@docker-compose -f docker-compose.local.yml ps

# Check LocalStack specific services in detail
localstack-health:
	@echo "☁️ Detailed LocalStack health check..."
	@echo ""
	@echo "🔍 LocalStack general health:"
	@curl -s http://localhost:4566/_localstack/health | python3 -m json.tool 2>/dev/null || echo " ❌ LocalStack not responding"
	@echo ""
	@echo "📁 S3 Buckets:"
	@docker exec proyecto1-localstack-local awslocal s3 ls 2>/dev/null || echo " ❌ Could not list S3 buckets"
	@echo ""
	@echo "📨 SQS Queues:"
	@docker exec proyecto1-localstack-local awslocal sqs list-queues 2>/dev/null || echo " ❌ Could not list SQS queues"
	@echo ""
	@echo "🔍 S3 Bucket contents:"
	@docker exec proyecto1-localstack-local awslocal s3 ls s3://proyecto1-videos/ --recursive 2>/dev/null || echo " ❌ Could not list bucket contents"

# Test LocalStack setup by creating and testing resources
localstack-test:
	@echo "🧪 Testing LocalStack setup..."
	@echo ""
	@echo "1️⃣ Testing S3 functionality:"
	@echo "   📁 Creating test file..."
	@echo "test content from makefile - $$(date)" | docker exec -i proyecto1-localstack-local awslocal s3 cp - s3://proyecto1-videos/test/makefile-test.txt 2>/dev/null && echo "   ✅ S3 upload OK" || echo "   ❌ S3 upload FAIL"
	@echo "   📋 Listing S3 contents:"
	@docker exec proyecto1-localstack-local awslocal s3 ls s3://proyecto1-videos/ --recursive 2>/dev/null | head -10
	@echo ""
	@echo "2️⃣ Testing SQS functionality:"
	@echo "   📨 Sending test message..."
	@docker exec proyecto1-localstack-local awslocal sqs send-message \
		--queue-url http://localhost:4566/000000000000/proyecto1-video-processing \
		--message-body '{"test":"makefile-message","timestamp":"'$$(date)'","action":"test"}' >/dev/null 2>&1 && echo "   ✅ SQS send OK" || echo "   ❌ SQS send FAIL"
	@echo "   📊 Checking queue attributes:"
	@docker exec proyecto1-localstack-local awslocal sqs get-queue-attributes \
		--queue-url http://localhost:4566/000000000000/proyecto1-video-processing \
		--attribute-names ApproximateNumberOfMessages 2>/dev/null | grep -o '"ApproximateNumberOfMessages":"[^"]*"' || echo "   ❌ SQS attributes FAIL"
	@echo ""
	@echo "3️⃣ Testing presigned URLs:"
	@docker exec proyecto1-localstack-local awslocal s3 presign s3://proyecto1-videos/test/makefile-test.txt --expires-in 300 2>/dev/null && echo "   ✅ Presigned URL OK" || echo "   ❌ Presigned URL FAIL"
	@echo ""
	@echo "4️⃣ Cleaning up test resources:"
	@docker exec proyecto1-localstack-local awslocal s3 rm s3://proyecto1-videos/test/makefile-test.txt 2>/dev/null && echo "   ✅ Cleanup OK" || echo "   ❌ Cleanup FAIL"

# Test ObjectStorage functionality (requires Go code to be compiled)
storage-test:
	@echo "🧪 Testing ObjectStorage functionality..."
	@echo ""
	@echo "ℹ️  This test requires your Go application to have ObjectStorage integration"
	@echo "   If not implemented yet, this will show how to use it:"
	@echo ""
	@echo "📝 Example Go code to test ObjectStorage:"
	@echo '   cfg := config.Load()'
	@echo '   s3Config := &providers.S3Config{'
	@echo '       AccessKeyID: "test",'
	@echo '       SecretAccessKey: "test",'
	@echo '       Region: "us-east-1",'
	@echo '       BucketName: "proyecto1-videos",'
	@echo '       EndpointURL: "http://localhost:4566",'
	@echo '   }'
	@echo '   provider, _ := providers.NewS3Provider(s3Config)'
	@echo '   manager := ObjectStorage.NewFileStorageManager(provider)'
	@echo '   err := manager.UploadFile([]byte("test"), "test-file.txt")'

# Test API endpoints with sample requests
test-api:
	@echo "🧪 Testing API endpoints..."
	@echo ""
	@echo "1️⃣ Testing nginx health:"
	@curl -s http://localhost:80/nginx-health && echo " ✅" || echo " ❌"
	@echo ""
	@echo "2️⃣ Testing API signup endpoint (should return validation error):"
	@curl -s -X POST http://localhost:80/api/auth/signup \
		-H "Content-Type: application/json" \
		-d '{}' | grep -q error && echo " ✅ Endpoint accessible" || echo " ❌ Endpoint not responding"
	@echo ""
	@echo "3️⃣ Testing video upload endpoint without auth (should return 401):"
	@curl -s -X POST http://localhost:80/api/videos/upload | grep -q "missing\|unauthorized\|invalid" && echo " ✅ Auth protection working" || echo " ❌ Auth not working"
	@echo ""
	@echo "4️⃣ Testing with LocalStack integration:"
	@echo "   (This requires ObjectStorage to be integrated in your API)"
	@echo ""
	@echo "For full API testing, use the examples in nginx/README.md"

# Stop all containers but keep volumes
stop:
	@echo "🛑 Stopping all containers..."
	docker-compose -f docker-compose.local.yml stop
	@echo "✅ All containers stopped (volumes preserved)"
	@echo ""
	@echo "💾 Persistent data preserved:"
	@echo "   🗄️  PostgreSQL data"
	@echo "   ☁️  LocalStack data"

# NUCLEAR CLEANUP - Destroys absolutely everything
clean:
	@echo "🚨 NUCLEAR CLEANUP - DESTROYING EVERYTHING..."
	@echo "⚠️  This will delete ALL data including:"
	@echo "   🗄️  PostgreSQL databases"
	@echo "   ☁️  LocalStack S3 buckets and SQS queues"
	@echo "   📦 Docker images and networks"
	@echo ""
	@read -p "Are you sure? Type 'yes' to continue: " confirm && [ "$$confirm" = "yes" ] || (echo "Cancelled." && exit 1)
	@echo ""
	@echo "Stopping and removing all project containers..."
	docker-compose -f docker-compose.local.yml down -v --remove-orphans
	@echo "Removing ALL project images..."
	docker-compose -f docker-compose.local.yml down --rmi all
	@echo "Pruning ALL unused Docker resources..."
	docker system prune -a -f --volumes
	@echo "Removing any leftover proyecto1 containers..."
	docker ps -aq --filter "name=proyecto1" | xargs -r docker rm -f
	@echo "Removing any leftover proyecto1 images..."
	docker images --filter "reference=*proyecto1*" -q | xargs -r docker rmi -f
	@echo "Removing any leftover proyecto1 volumes..."
	docker volume ls --filter "name=proyecto1" -q | xargs -r docker volume rm
	@echo "Removing any leftover proyecto_1 volumes..."
	docker volume ls --filter "name=proyecto_1" -q | xargs -r docker volume rm
	@echo "Removing any leftover localstack volumes..."
	docker volume ls --filter "name=localstack" -q | xargs -r docker volume rm
	@echo "Removing any leftover proyecto networks..."
	docker network ls --filter "name=proyecto" -q | xargs -r docker network rm
	@echo ""
	@echo "💥 NUCLEAR CLEANUP COMPLETE - Everything obliterated!"
	@echo "🚀 Run 'make local' to start fresh"