-- Perintah untuk menghapus semua tabel (Migrasi TURUN / Rollback - PostgreSQL)

DROP TABLE IF EXISTS "item_templates";
DROP TABLE IF EXISTS "configurations";
DROP TABLE IF EXISTS "audit_logs";
DROP TABLE IF EXISTS "lost_items";
DROP TABLE IF EXISTS "lost_documents";
DROP TABLE IF EXISTS "residents";
DROP TABLE IF EXISTS "users";