# ConnectVN - GopherSocial Backend

ConnectVN là một hệ thống Social Network được xây dựng bằng ngôn ngữ Golang, tập trung vào khả năng mở rộng, hiệu năng cao và trải nghiệm người dùng thời gian thực. Dự án tuân thủ nghiêm ngặt nguyên lý Clean Architecture để đảm bảo code dễ bảo trì và kiểm thử.

---

## Tech Stack (Công nghệ sử dụng)

- Ngôn ngữ: Golang (Go 1.21+)
- Web Framework / Router: Gin HTTP Framework
- Database (RDBMS): PostgreSQL
- In-Memory Data / Cache: Redis (Xử lý OTP, Rate Limiting, quản lý trạng thái Online/Offline bằng Hash Map)
- Message Broker: Apache Kafka (Xử lý hàng đợi bất đồng bộ cho các sự kiện tương tác và cơ chế Fan-out)
- Object Storage: MinIO (S3 Compatible - Lưu trữ Avatar, Cover, Media)
- Real-time Engine: Gorilla WebSocket kết hợp Redis Pub/Sub (Mô hình phân tán Server-to-Client)
- Infrastructure: Docker & Docker Compose

---

## Tính năng nổi bật

- Authentication & Security:
  - Xác thực qua JWT.
  - Hệ thống mã OTP gửi qua Gmail cho chức năng quên mật khẩu, lưu trữ trong Redis với cơ chế tự hủy (TTL).

- User Management:
  - Quản lý hồ sơ cá nhân (Bio, Avatar, Cover Image).
  - Upload và lưu trữ hình ảnh hiệu suất cao qua MinIO.

- Social Interactions:
  - Hệ thống Follow/Unfollow.
  - Tương tác bài viết: Like, Comment (Lồng nhau) - Xử lý bất đồng bộ qua Kafka để tối ưu API.

- Real-time Engine:
  - Chat 1-1 thời gian thực qua WebSockets.
  - Theo dõi trạng thái Online/Offline của bạn bè theo thời gian thực.
  - Thông báo (Notifications) tức thì đẩy trực tiếp về Client.

- High-Performance Feed:
  - Tối ưu hóa truy vấn bảng tin (Newsfeed).
  - Thiết kế sẵn kiến trúc Push Model (Fan-out on Write) sử dụng Apache Kafka để xử lý phân phối bài viết cho lượng dữ liệu lớn.

