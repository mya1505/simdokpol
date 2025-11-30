-- Perintah untuk membuat semua tabel (Migrasi NAIK - PostgreSQL)

CREATE TABLE "users" (
    "id" SERIAL PRIMARY KEY,
    "nama_lengkap" VARCHAR(255) NOT NULL,
    "nrp" VARCHAR(20) NOT NULL UNIQUE,
    "kata_sandi" VARCHAR(255) NOT NULL,
    "pangkat" VARCHAR(100),
    "peran" VARCHAR(50) NOT NULL DEFAULT 'OPERATOR',
    "jabatan" VARCHAR(100),
    "regu" VARCHAR(10),
    "created_at" TIMESTAMPTZ,
    "updated_at" TIMESTAMPTZ,
    "deleted_at" TIMESTAMPTZ
);
CREATE INDEX "idx_users_deleted_at" ON "users"("deleted_at");

CREATE TABLE "residents" (
    "id" SERIAL PRIMARY KEY,
    "nik" VARCHAR(16) NOT NULL UNIQUE,
    "nama_lengkap" VARCHAR(255) NOT NULL,
    "tempat_lahir" VARCHAR(100) NOT NULL,
    "tanggal_lahir" TIMESTAMPTZ NOT NULL,
    "jenis_kelamin" VARCHAR(20) NOT NULL,
    "agama" VARCHAR(50) NOT NULL,
    "pekerjaan" VARCHAR(100) NOT NULL,
    "alamat" TEXT NOT NULL,
    "created_at" TIMESTAMPTZ,
    "updated_at" TIMESTAMPTZ,
    "deleted_at" TIMESTAMPTZ
);
CREATE INDEX "idx_residents_deleted_at" ON "residents"("deleted_at");

CREATE TABLE "lost_documents" (
    "id" SERIAL PRIMARY KEY,
    "nomor_surat" VARCHAR(255) NOT NULL UNIQUE,
    "tanggal_laporan" TIMESTAMPTZ NOT NULL,
    "status" VARCHAR(50) NOT NULL DEFAULT 'DITERBITKAN',
    "lokasi_hilang" TEXT,
    "resident_id" INTEGER NOT NULL REFERENCES "residents"("id"),
    "petugas_pelapor_id" INTEGER NOT NULL REFERENCES "users"("id"),
    "pejabat_persetuju_id" INTEGER REFERENCES "users"("id"),
    "operator_id" INTEGER NOT NULL REFERENCES "users"("id"),
    "last_updated_by_id" INTEGER REFERENCES "users"("id"),
    "tanggal_persetujuan" TIMESTAMPTZ,
    "created_at" TIMESTAMPTZ,
    "updated_at" TIMESTAMPTZ,
    "deleted_at" TIMESTAMPTZ
);
CREATE INDEX "idx_lost_documents_deleted_at" ON "lost_documents"("deleted_at");

CREATE TABLE "lost_items" (
    "id" SERIAL PRIMARY KEY,
    "lost_document_id" INTEGER NOT NULL REFERENCES "lost_documents"("id"),
    "nama_barang" VARCHAR(255) NOT NULL,
    "deskripsi" TEXT
);

CREATE TABLE "audit_logs" (
    "id" SERIAL PRIMARY KEY,
    "user_id" INTEGER NOT NULL REFERENCES "users"("id"),
    "aksi" VARCHAR(255) NOT NULL,
    "detail" TEXT,
    "timestamp" TIMESTAMPTZ NOT NULL
);

CREATE TABLE "configurations" (
    "key" VARCHAR(255) PRIMARY KEY,
    "value" TEXT,
    "updated_at" TIMESTAMPTZ
);

CREATE TABLE "item_templates" (
    "id" SERIAL PRIMARY KEY,
    "nama_barang" VARCHAR(100) NOT NULL UNIQUE,
    "fields_config" JSONB, 
    "is_active" BOOLEAN NOT NULL DEFAULT TRUE,
    "urutan" INTEGER DEFAULT 0,
    "created_at" TIMESTAMPTZ,
    "updated_at" TIMESTAMPTZ,
    "deleted_at" TIMESTAMPTZ
);
CREATE INDEX "idx_item_templates_deleted_at" ON "item_templates"("deleted_at");

-- SEEDING DATA
INSERT INTO "item_templates" ("nama_barang", "fields_config", "urutan")
VALUES
    ('KTP', '[{"label":"NIK","type":"text","data_label":"NIK","regex":"^[0-9]{16}$","placeholder":"16 Digit NIK","required_length":16,"min_length":16,"max_length":16,"is_numeric":true}]', 1),
    ('SIM', '[{"label":"Golongan SIM","type":"select","data_label":"Gol","options":["A","B I","B II","C","D"]},{"label":"Nomor SIM","type":"text","data_label":"No. SIM","regex":"^[0-9]{12,14}$","is_numeric":true}]', 2),
    ('STNK', '[{"label":"Nomor Polisi","type":"text","data_label":"No. Pol"},{"label":"Nomor Rangka","type":"text","data_label":"No. Rangka"},{"label":"Nomor Mesin","type":"text","data_label":"No. Mesin"}]', 3),
    ('BPKB', '[{"label":"Nomor BPKB","type":"text","data_label":"No. BPKB"},{"label":"Atas Nama","type":"text","data_label":"a.n."}]', 4),
    ('IJAZAH', '[{"label":"Tingkat","type":"select","data_label":"Tingkat","options":["SD","SMP","SMA","D3","S1"]},{"label":"Nomor Ijazah","type":"text","data_label":"No. Ijazah"}]', 5),
    ('ATM', '[{"label":"Nama Bank","type":"select","data_label":"Bank","options":["BRI","BCA","Mandiri"]},{"label":"Nomor Rekening","type":"text","data_label":"No. Rek"}]', 6),
    ('LAINNYA', '[]', 99)
ON CONFLICT ("nama_barang") DO NOTHING;