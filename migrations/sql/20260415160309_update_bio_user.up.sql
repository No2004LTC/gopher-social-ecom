-- Migration: update_bio_user
ALTER TABLE users ADD COLUMN bio TEXT DEFAULT '';