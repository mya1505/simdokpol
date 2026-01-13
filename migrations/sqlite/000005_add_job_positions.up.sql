CREATE TABLE `job_positions` (
    `id` integer PRIMARY KEY AUTOINCREMENT,
    `nama` text NOT NULL UNIQUE,
    `is_active` boolean NOT NULL DEFAULT 1,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime
);
CREATE INDEX `idx_job_positions_deleted_at` ON `job_positions`(`deleted_at`);
