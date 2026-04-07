-- Migration: posts_update_count
ALTER TABLE posts ADD COLUMN likes_count INT DEFAULT 0;
ALTER TABLE posts ADD COLUMN comments_count INT DEFAULT 0;