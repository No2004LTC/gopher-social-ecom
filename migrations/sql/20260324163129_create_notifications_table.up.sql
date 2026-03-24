-- Migration: create_notifications_table
CREATE TABLE notifications (
                               id SERIAL PRIMARY KEY,
                               user_id INT NOT NULL,
                               actor_id INT NOT NULL,
                               type VARCHAR(50),
                               entity_id INT,
                               is_read BOOLEAN DEFAULT FALSE,
                               created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);