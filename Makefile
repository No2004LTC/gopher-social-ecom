# 1. Load biến môi trường từ .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# 2. Định nghĩa biến
API_ENTRY = cmd/api/main.go
MIGRATE_ENTRY = cmd/migrate/main.go
MIGRATION_DIR = migrations/sql

# Infrastructure
#Bật container
up:
	docker compose up -d

#Tắt container
down:
	docker compose down

# import thêm các thư viện cần vào go.mod
tidy:
	go mod tidy

#Chạy
run: tidy
	go run $(API_ENTRY)

# 3. Migration (Sửa lại để truyền tham số vào code Go)
# Chạy toàn bộ các file .up.sql chưa thực thi
migrate-up: tidy
	go run $(MIGRATE_ENTRY) up

# Tắt 1 bảng gần nhất ( file .down.sql gần nhất)
migrate-down: tidy
	go run $(MIGRATE_ENTRY) down

# Xóa sạch tất cả các bảng đã tạo (chạy tất cả file .down.sql)
migrate-drop: tidy
	go run $(MIGRATE_ENTRY) drop

# Dọn sạch và tạo lại từ đầu chỉ với 1 lệnh
db-reset:
	go run $(MIGRATE_ENTRY) drop
	go run $(MIGRATE_ENTRY) up

# Tạo file migration mới
new-migration:
	@read -p "Nhập tên migration: " desc; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	mkdir -p $(MIGRATION_DIR); \
	touch $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql; \
	touch $(MIGRATION_DIR)/$${timestamp}_$${desc}.down.sql; \
	echo "-- Migration: $${desc}" >> $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql; \
	echo "✅ Đã tạo: $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql"

test: tidy
	go test -v -cover ./...

.PHONY: up down tidy run db-reset migrate-up migrate-down migrate-drop migrate-force new-migration test