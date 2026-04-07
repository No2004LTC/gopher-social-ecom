CREATE TABLE notifications (
                               id SERIAL PRIMARY KEY,
                               user_id INT NOT NULL,          -- Người nhận thông báo
                               actor_id INT NOT NULL,         -- Người gây ra hành động (người like, follow...)
                               type VARCHAR(50) NOT NULL,      -- 'LIKE', 'FOLLOW', 'COMMENT', 'NEW_POST'
                               entity_id INT NOT NULL,        -- ID của Post, Comment... liên quan
                               message TEXT NOT NULL,
                               is_read BOOLEAN DEFAULT FALSE,
                               created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

                               FOREIGN KEY (user_id) REFERENCES users(id),
                               FOREIGN KEY (actor_id) REFERENCES users(id)
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);