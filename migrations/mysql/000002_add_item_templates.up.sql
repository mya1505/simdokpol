-- Migrasi MySQL: Membuat tabel item_templates
CREATE TABLE `item_templates` (
    `id` integer AUTO_INCREMENT PRIMARY KEY,
    `nama_barang` varchar(255) NOT NULL UNIQUE,
    `fields_config` text,
    `is_active` boolean NOT NULL DEFAULT true,
    `urutan` integer NOT NULL DEFAULT 0,
    `created_at` datetime(3),
    `updated_at` datetime(3),
    `deleted_at` datetime(3),
    INDEX `idx_item_templates_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Masukkan data awal (sintaks INSERT sama)
INSERT INTO `item_templates` (nama_barang, fields_config, urutan)
VALUES
    ('KTP', '[{"label":"NIK","type":"text","data_label":"NIK","regex":"^[0-9]{16}$","placeholder":"16 Digit NIK","required_length":16,"min_length":16,"max_length":16,"is_numeric":true,"is_uppercase":false,"is_titlecase":false}]', 1),
    ('SIM', '[{"label":"Golongan SIM","type":"select","data_label":"Gol","options":["A","B I","B II","C","D"],"is_numeric":false,"is_uppercase":false,"is_titlecase":false},{"label":"Nomor SIM","type":"text","data_label":"No. SIM","regex":"^[0-9]{12,14}$","placeholder":"12-14 Digit No. SIM","min_length":12,"max_length":14,"is_numeric":true,"is_uppercase":false,"is_titlecase":false}]', 2),
    ('STNK', '[{"label":"Nomor Polisi","type":"text","data_label":"No. Pol","regex":"^[A-Z0-9 ]{1,10}$","placeholder":"Contoh: DD 1234 AB","max_length":10,"is_numeric":false,"is_uppercase":true,"is_titlecase":false},{"label":"Nomor Rangka","type":"text","data_label":"No. Rangka","regex":"^[A-Z0-9]{17}$","placeholder":"17 Digit No. Rangka (VIN)","required_length":17,"min_length":17,"max_length":17,"is_numeric":false,"is_uppercase":true,"is_titlecase":false},{"label":"Nomor Mesin","type":"text","data_label":"No. Mesin","regex":"^[A-Z0-9]{1,15}$","placeholder":"Hingga 15 digit (huruf & angka)","max_length":15,"is_numeric":false,"is_uppercase":true,"is_titlecase":false}]', 3),
    ('BPKB', '[{"label":"Nomor BPKB","type":"text","data_label":"No. BPKB","regex":"^[A-Z0-9]{9}$","placeholder":"9 Digit No. BPKB (huruf & angka)","required_length":9,"min_length":9,"max_length":9,"is_numeric":false,"is_uppercase":true,"is_titlecase":false},{"label":"Atas Nama","type":"text","data_label":"a.n.","placeholder":"Nama di BPKB","is_numeric":false,"is_uppercase":false,"is_titlecase":true}]', 4),
    ('IJAZAH', '[{"label":"Tingkat Ijazah","type":"select","data_label":"Tingkat","options":["SD","SMP","SMA/SMK","D3","S1","S2","S3"],"is_numeric":false,"is_uppercase":false,"is_titlecase":false},{"label":"Nomor Ijazah","type":"text","data_label":"No. Ijazah","regex":"^[A-Z0-9\\/-]{1,50}$","placeholder":"No. Ijazah (termasuk / dan -)","max_length":50,"is_numeric":false,"is_uppercase":true,"is_titlecase":false}]', 5),
    ('ATM', '[{"label":"Nama Bank","type":"select","data_label":"Bank","options":["BRI","BCA","Mandiri","BNI","BTN","Lainnya"],"is_numeric":false,"is_uppercase":false,"is_titlecase":false},{"label":"Nomor Rekening","type":"text","data_label":"No. Rek","regex":"^[0-9]{1,20}$","placeholder":"Hingga 20 digit angka","max_length":20,"is_numeric":true,"is_uppercase":false,"is_titlecase":false}]', 6),
    ('LAINNYA', '[]', 99);