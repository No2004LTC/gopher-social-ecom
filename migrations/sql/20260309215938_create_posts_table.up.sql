-- 1. Tạo bảng posts
CREATE TABLE IF NOT EXISTS posts (
                                     id BIGSERIAL PRIMARY KEY,
                                     user_id BIGINT NOT NULL,
                                     content TEXT,
                                     image_url TEXT,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Ràng buộc khóa ngoại: Nếu User bị xóa, các bài Post liên quan cũng tự động xóa (Cascade)
                                     CONSTRAINT fk_posts_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 2. Tạo Index cho user_id để tăng tốc độ truy vấn khi lấy Newsfeed theo User
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);

-- 3. Tạo Index cho created_at vì Newsfeed luôn sắp xếp theo thời gian mới nhất
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);