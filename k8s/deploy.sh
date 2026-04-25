#!/bin/bash

export SERVICE_NAME=${SERVICE_NAME:-"gopher-be"}
export IMAGE_NAME=${IMAGE_NAME:-"your-docker-username/gopher-be"}
export IMAGE_TAG=${IMAGE_TAG:-"latest"}

echo "🚀 Đang triển khai dịch vụ: $SERVICE_NAME"


envsubst < k8s/main.yaml | kubectl apply -f -

echo "⏳ Chờ $SERVICE_NAME sẵn sàng..."
kubectl rollout status deployment/$SERVICE_NAME --timeout=60s

echo "✅ Triển khai thành công!"