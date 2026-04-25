# Load biến môi trường từ .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# entry points cho API và Migration
API_ENTRY = cmd/api/main.go
WORKER_ENTRY = cmd/worker/main.go
MIGRATE_ENTRY = cmd/migrate/main.go
MIGRATION_DIR = migrations/sql

# Biến cho Kubernetes
export SERVICE_NAME := gopher-be
export IMAGE_NAME   := your-docker-username/gopher-be
export IMAGE_TAG    := latest

.PHONY: up down tidy run run-worker run-all db-reset migrate-up migrate-down migrate-drop migrate-force new-migration test k8s-up k8s-down test-env

# --- Docker Compose ---
# Run container
up:
	docker compose up -d

#Tat container
down: ## Tắt container Infra
	docker compose down

# --- Go Development ---
# Kiem tra và cập nhật dependencies
tidy:
	go mod tidy

# Chạy API
run: tidy
	go run $(API_ENTRY)

run-worker: tidy
	go run $(WORKER_ENTRY)

test: tidy
	go test -v -cover ./...

# --- Migration ---
# Tao schema db
migrate-up: tidy
	go run $(MIGRATE_ENTRY) up

# Rollback schema db(-1)
migrate-down: tidy
	go run $(MIGRATE_ENTRY) down

# Drop schema db(all)
migrate-drop: tidy
	go run $(MIGRATE_ENTRY) drop

# Reset schema db (drop + up)
db-reset:
	go run $(MIGRATE_ENTRY) drop
	go run $(MIGRATE_ENTRY) up

# Tao file migration mới theo timestamp
new-migration:
	@read -p "Nhập tên migration: " desc; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	mkdir -p $(MIGRATION_DIR); \
	touch $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql; \
	touch $(MIGRATION_DIR)/$${timestamp}_$${desc}.down.sql; \
	echo "-- Migration: $${desc}" >> $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql; \
	echo "✅ Đã tạo: $(MIGRATION_DIR)/$${timestamp}_$${desc}.up.sql"

# --- Kubernetes  ---
k8s-up:
	@echo "🛠 1. Kiểm tra cụm K8s Local (Kind)..."
	@kind get clusters | grep -q "gopher-social" || kind create cluster --name gopher-social
	@echo "📦 2. Build và Load Image..."
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	kind load docker-image $(IMAGE_NAME):$(IMAGE_TAG) --name gopher-social

	@echo "⚙️ 3. Triển khai ConfigMap (Giải quyết lỗi CreateContainerConfigError)..."
	if [ -f k8s/config.yaml ]; then \
		envsubst < k8s/config.yaml | kubectl apply -f -; \
	else \
		echo "⚠️ Cảnh báo: Không tìm thấy k8s/config.yaml"; \
	fi

	@echo "🚀 4. Triển khai App (Deployment & Service)..."
	envsubst < k8s/main.yaml | kubectl apply -f -

	@echo "⏳ 5. Đợi hệ thống khởi động..."
	kubectl wait --for=condition=available deployment/$(SERVICE_NAME) --timeout=60s
	@echo "✅ Chúc mừng! Hệ thống đã lên xanh lè."

k8s-down: ## Xóa bỏ môi trường K8s
	@echo "🗑 Đang dọn dẹp K8s..."
	envsubst < k8s/main.yaml | kubectl delete -f - || true
	envsubst < k8s/config.yaml | kubectl delete -f - || true
	kind delete cluster --name gopher-social


test-env:
	@echo "Service name là: $$SERVICE_NAME"
	envsubst < k8s/main.yaml | grep "name:" 