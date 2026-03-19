-- Migration: like_and_comment
-- Bảng Likes: Sử dụng Composite Primary Key để tránh 1 user like 1 bài nhiều lần
CREATE TABLE likes (
                       user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
                       post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                       PRIMARY KEY (user_id, post_id)
);

-- Bảng Comments
CREATE TABLE comments (
                          id BIGSERIAL PRIMARY KEY,
                          user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
                          post_id BIGINT REFERENCES posts(id) ON DELETE CASCADE,
                          content TEXT NOT NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                          updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);