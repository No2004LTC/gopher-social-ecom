-- Migration: add_cover_url_to_users
ALTER TABLE users ADD COLUMN cover_url VARCHAR(255) DEFAULT '';