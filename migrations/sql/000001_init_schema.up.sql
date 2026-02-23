-- Đây là file "Up" - dùng để tạo hoặc nâng cấp cấu trúc DB
CREATE TABLE "users" (
                         "id" BIGSERIAL PRIMARY KEY,           -- ID tự tăng, kiểu lớn cho quy mô mạng xã hội
                         "username" varchar UNIQUE NOT NULL,    -- Username không được trùng
                         "email" varchar UNIQUE NOT NULL,       -- Email không được trùng
                         "password_hash" varchar NOT NULL,      -- Lưu mật khẩu đã mã hóa
                         "avatar_url" varchar DEFAULT '',       -- Ảnh đại diện (mặc định trống)
                         "created_at" timestamptz NOT NULL DEFAULT (now()) -- Tự động lưu thời gian tạo (có múi giờ)
);

-- Tạo Index để tìm kiếm nhanh hơn theo username và email
CREATE INDEX ON "users" ("username");
CREATE INDEX ON "users" ("email");