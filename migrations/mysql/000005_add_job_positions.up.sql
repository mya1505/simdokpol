-- +migrate Up

CREATE TABLE `job_positions` (
    `id` integer AUTO_INCREMENT PRIMARY KEY,
    `nama` varchar(150) NOT NULL UNIQUE,
    `is_active` boolean NOT NULL DEFAULT true,
    `created_at` datetime(3),
    `updated_at` datetime(3),
    `deleted_at` datetime(3),
    INDEX `idx_job_positions_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
