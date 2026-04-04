-- Migration: create_follows_table
CREATE TABLE IF NOT EXISTS "follows" (
                                         "follower_id" BIGINT NOT NULL,
                                         "following_id" BIGINT NOT NULL,
                                         "created_at" TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Khóa chính phức hợp: Chặn đứng việc Follow trùng lặp
                                         PRIMARY KEY ("follower_id", "following_id"),

    -- Ràng buộc với bảng users (Nếu xóa user thì tự động xóa follow liên quan)
                                         CONSTRAINT "fk_follower" FOREIGN KEY ("follower_id") REFERENCES "users"("id") ON DELETE CASCADE,
                                         CONSTRAINT "fk_following" FOREIGN KEY ("following_id") REFERENCES "users"("id") ON DELETE CASCADE
);

-- Index này cực kỳ quan trọng để đếm số Follower (ai đang follow mình) nhanh hơn
CREATE INDEX IF NOT EXISTS "idx_follows_following_id" ON "follows"("following_id");