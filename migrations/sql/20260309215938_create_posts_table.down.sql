-- Xóa Index trước (tốt cho hiệu năng khi xóa bảng lớn)
DROP INDEX IF EXISTS idx_posts_created_at;
DROP INDEX IF EXISTS idx_posts_user_id;

-- Xóa bảng posts
DROP TABLE IF EXISTS posts;