# 🐹 Gopher Social - Backend K8s Deployment

Dự án sử dụng mô hình Hybrid: Infrastructure chạy Docker Compose và Application chạy Kubernetes (Kind).

### 🚀 Cách chạy nhanh
1. **Khởi động Infra:** `docker-compose up -d`
2. **Triển khai App vào K8s:** `./deploy.sh`

### 🛠 Yêu cầu
- Ubuntu 24.04 LTS
- Docker, Kind, Kubectl, envsubst (thường có sẵn trong Linux)