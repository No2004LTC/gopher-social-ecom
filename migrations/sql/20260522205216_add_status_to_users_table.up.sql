-- Migration: add_status_to_users_table
ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active';
