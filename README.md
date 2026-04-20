# ConnectVN - GopherSocial-Commerce Backend

**ConnectVN** là một hệ thống Social Network kết hợp E-commerce được xây dựng bằng ngôn ngữ **Golang**, tập trung vào khả năng mở rộng, hiệu năng cao và trải nghiệm người dùng thời gian thực. Dự án tuân thủ nghiêm ngặt nguyên lý **Clean Architecture** để đảm bảo code dễ bảo trì và kiểm thử.

---

## Tính năng nổi bật

- **Authentication & Security**:
    - Xác thực qua JWT.
    - Hệ thống mã OTP gửi qua Gmail cho chức năng quên mật khẩu, lưu trữ trong **Redis** với cơ chế tự hủy (TTL).
- **User Management**:
    - Quản lý hồ sơ cá nhân (Bio, Avatar, Cover Image).
    - Tích hợp **MinIO** (S3 compatible) để lưu trữ hình ảnh hiệu quả.
- **Social Interactions**:
    - Hệ thống Follow/Unfollow.
    - Tương tác bài viết: Like, Comment (Lồng nhau).
- **Real-time Engine**:
    - Chat thời gian thực qua **WebSockets**.
    - Theo dõi trạng thái Online/Offline của bạn bè thông qua **Redis Hash Map**.
    - Thông báo (Notifications) tức thì.
- **High-Performance Feed**:
    - Tối ưu hóa truy vấn bảng tin (Newsfeed).
    - Thiết kế sẵn kiến trúc **Push Model (Fan-out on Write)** sử dụng **Apache Kafka** để xử lý dữ liệu lớn.

---

## Kiến trúc dự án (Clean Architecture)

Dự án được chia làm 4 lớp chính:

1.  **Domain (Entities & Interfaces)**: Chứa các định nghĩa model và giao tiếp giữa các tầng.
2.  **Usecase (Business Logic)**: Xử lý nghiệp vụ chính của hệ thống (Được phân tách thành `AuthUsecase` và `UserUsecase`).
3.  **Repository (Data Access)**: Làm việc với Database (PostgreSQL) và Cache (Redis).
4.  **Delivery (Handlers & Middlewares)**: Tiếp nhận các request từ HTTP/WebSocket.

```text
internal/
├── domain/             # Interfaces & Models
├── usecase/            # Business Logic (Auth, User, Chat, Post...)
├── repository/         # Data Access (Postgres, Redis)
└── delivery/
    ├── http/           # HTTP Handlers & Middlewares
    └── ws/             # WebSocket Hub & Handler
pkg/                    # Shared packages (Mail, Storage, Auth...)
cmd/api/main.go         # Entry point