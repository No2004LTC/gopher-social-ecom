-- Migration: add_is_read_to_messages
ALTER TABLE messages ADD COLUMN is_read BOOLEAN DEFAULT FALSE;