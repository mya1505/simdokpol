-- Migrasi 000002: Menambahkan tabel lisensi
CREATE TABLE "licenses" (
    "key" TEXT PRIMARY KEY,
    "status" TEXT NOT NULL,
    "activated_at" DATETIME,
    "activated_by_id" INTEGER,
    "notes" TEXT
);

-- Masukkan key 'status' default agar kita bisa cek
INSERT INTO "configurations" ("key", "value") 
VALUES ('license_status', 'UNLICENSED') 
ON CONFLICT("key") DO NOTHING;