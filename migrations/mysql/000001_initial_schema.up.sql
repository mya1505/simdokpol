-- +migrate Up

CREATE TABLE `users` (
    `id` integer AUTO_INCREMENT PRIMARY KEY,
    `nama_lengkap` varchar(255) NOT NULL,
    `nrp` varchar(20) NOT NULL UNIQUE,
    `kata_sandi` varchar(255) NOT NULL,
    `pangkat` varchar(100),
    `peran` varchar(50) NOT NULL DEFAULT 'OPERATOR',
    `jabatan` varchar(100),
    `regu` varchar(10),
    `created_at` datetime(3),
    `updated_at` datetime(3),
    `deleted_at` datetime(3),
    INDEX `idx_users_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `residents` (
    `id` integer AUTO_INCREMENT PRIMARY KEY,
    `nik` varchar(16) NOT NULL UNIQUE,
    `nama_lengkap` varchar(255) NOT NULL,
    `tempat_lahir` varchar(100) NOT NULL,
    `tanggal_lahir` datetime(3) NOT NULL,
    `jenis_kelamin` varchar(20) NOT NULL,
    `agama` varchar(50) NOT NULL,
    `pekerjaan` varchar(100) NOT NULL,
    `alamat` text NOT NULL,
    `created_at` datetime(3),
    `updated_at` datetime(3),
    `deleted_at` datetime(3),
    INDEX `idx_residents_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `lost_documents` (
    `id` integer AUTO_INCREMENT PRIMARY KEY,
    `nomor_surat` varchar(255) NOT NULL UNIQUE,
    `tanggal_laporan` datetime(3) NOT NULL,
    `status` varchar(50) NOT NULL DEFAULT 'DITERBITKAN',
    `lokasi_hilang` text,
    `resident_id` integer NOT NULL,
    `petugas_pelapor_id` integer NOT NULL,
    `pejabat_persetuju_id` integer,
    `operator_id` integer NOT NULL,
    `last_updated_by_id` integer,
    `tanggal_persetujuan` datetime(3),
    `created_at` datetime(3),
    `updated_at` datetime(3),
    `deleted_at` datetime(3),
    FOREIGN KEY (`pejabat_persetuju_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`operator_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`last_updated_by_id`) REFERENCES `users`(`id`),
    FOREIGN KEY (`resident_id`) REFERENCES `residents`(`id`),
    FOREIGN KEY (`petugas_pelapor_id`) REFERENCES `users`(`id`),
    INDEX `idx_lost_documents_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `lost_items` (
    `id` integer AUTO_INCREMENT PRIMARY KEY,
    `lost_document_id` integer NOT NULL,
    `nama_barang` varchar(255) NOT NULL,
    `deskripsi` text,
    FOREIGN KEY (`lost_document_id`) REFERENCES `lost_documents`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `audit_logs` (
    `id` integer AUTO_INCREMENT PRIMARY KEY,
    `user_id` integer NOT NULL,
    `aksi` varchar(255) NOT NULL,
    `detail` text,
    `timestamp` datetime(3) NOT NULL,
    FOREIGN KEY (`user_id`) REFERENCES `users`(`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `configurations` (
    `key` varchar(255) PRIMARY KEY,
    `value` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;