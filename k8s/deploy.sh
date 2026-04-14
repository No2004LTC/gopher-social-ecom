#!/bin/bash

# 1. Định nghĩa các biến (Nếu chạy local thì dùng mặc định, nếu CI thì lấy từ env)
export SERVICE_NAME=${SERVICE_NAME:-"gopher-be"}
export IMAGE_NAME=${IMAGE_NAME:-"your-docker-username/gopher-be"}
export IMAGE_TAG=${IMAGE_TAG:-"latest"}

echo "🚀 Đang triển khai dịch vụ: $SERVICE_NAME"

# 2. Xử lý file YAML (Render)
# Dùng envsubst để thay thế toàn bộ $SERVICE_NAME, $IMAGE_NAME... vào file main.yaml
# Sau đó đẩy thẳng vào kubectl
envsubst < k8s/main.yaml | kubectl apply -f -

# 3. Kiểm tra trạng thái
echo "⏳ Chờ $SERVICE_NAME sẵn sàng..."
kubectl rollout status deployment/$SERVICE_NAME --timeout=60s

echo "✅ Triển khai thành công!"