-- Migration: create_messages_table
CREATE TABLE messages (
                          id BIGSERIAL PRIMARY KEY,
                          from_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
                          to_user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
                          content TEXT NOT NULL,
                          created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index để truy vấn lịch sử chat nhanh hơn
CREATE INDEX idx_messages_conversation ON messages(from_user_id, to_user_id);