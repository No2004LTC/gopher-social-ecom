# Đọc biến môi trường từ .env
include .env
export

# Định nghĩa các đường dẫn chính
API_ENTRY = cmd/api/main.go
MIGRATE_ENTRY = cmd/migrate/main.go
MIGRATION_DIR = migrations/sql

# 1. Hạ tầng (Infrastructure) - Sử dụng Docker Compose V2
up:
	docker compose up -d

down:
	docker compose down

# 2. Phát triển (Development)
tidy:
	go mod tidy

# Chạy Server API
run: tidy
	go run $(API_ENTRY)

# 3. Migration (Học tập style của công ty bạn)
# Chạy migration bằng chính code Go chúng ta vừa viết
migrate: tidy
	go run $(MIGRATE_ENTRY)

# Tạo file migration mới với timestamp chuẩn xác
new-migration:
	@read -p "Nhập tên migration (VD: create_users_table): " desc; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	touch $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql; \
	touch $(MIGRATION_DIR)/$${timestamp}_$${desc}.down.sql; \
	echo "-- Migration: $${desc}" >> $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql; \
	echo "✅ Đã tạo: $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql"

# 4. Kiểm thử (Testing)
test: tidy
	go test -v -cover ./...

.PHONY: up down tidy run migrate new-migration test