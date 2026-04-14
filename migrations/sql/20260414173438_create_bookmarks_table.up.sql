-- Migration: create_bookmarks_table
CREATE TABLE IF NOT EXISTS bookmarks (
                                         user_id BIGINT NOT NULL,
                                         post_id BIGINT NOT NULL,
                                         created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    -- Thiết lập khóa chính phức hợp (Composite Primary Key)
    -- Đảm bảo 1 user không thể lưu 1 bài viết quá 1 lần
                                         PRIMARY KEY (user_id, post_id),

    -- Khóa ngoại liên kết tới bảng users và posts
                                         CONSTRAINT fk_bookmarks_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
                                         CONSTRAINT fk_bookmarks_post FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

-- Tạo Index để truy vấn danh sách "Đã lưu" của User nhanh hơn
CREATE INDEX idx_bookmarks_user_id ON bookmarks(user_id);