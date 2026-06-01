-- Migration: seed_admin_user
INSERT INTO users (username, email, password_hash, avatar_url, created_at, updated_at)
VALUES (
         'lethanhcong20052004',
         'lethanhcong20052004@gmail.com',
         '$argon2id$v=19$m=65536,t=3,p=2$4vP+uunad+cHjXdx4wHZOQ$19J2BMRwpQlv2X4xZKL29JUdNeBCTEbxmi3ngXNNJpg',
         '',
         CURRENT_TIMESTAMP,
         CURRENT_TIMESTAMP
       ) ON CONFLICT (email) DO NOTHING;
