-- Migration 009: perubahan idempotency key notifikasi untuk broadcast perubahan kuota
-- Mengizinkan satu outbox event disiarkan ke banyak user, tetapi menjamin
-- satu user hanya menerima tepat satu email per outbox event.

ALTER TABLE notifications 
    DROP CONSTRAINT IF EXISTS notifications_outbox_unique;

ALTER TABLE notifications 
    DROP CONSTRAINT IF EXISTS notifications_outbox_user_unique;

ALTER TABLE notifications 
    ADD CONSTRAINT notifications_outbox_user_unique UNIQUE (outbox_id, user_id);

