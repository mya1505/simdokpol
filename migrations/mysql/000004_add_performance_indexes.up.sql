-- +migrate Up

CREATE INDEX `idx_lost_documents_tanggal_laporan` ON `lost_documents`(`tanggal_laporan`);
CREATE INDEX `idx_lost_documents_operator_id` ON `lost_documents`(`operator_id`);
CREATE INDEX `idx_lost_documents_resident_id` ON `lost_documents`(`resident_id`);
CREATE INDEX `idx_lost_documents_petugas_pelapor_id` ON `lost_documents`(`petugas_pelapor_id`);
CREATE INDEX `idx_lost_documents_pejabat_persetuju_id` ON `lost_documents`(`pejabat_persetuju_id`);
CREATE INDEX `idx_lost_documents_last_updated_by_id` ON `lost_documents`(`last_updated_by_id`);

CREATE INDEX `idx_lost_items_lost_document_id` ON `lost_items`(`lost_document_id`);

CREATE INDEX `idx_audit_logs_timestamp` ON `audit_logs`(`timestamp`);
CREATE INDEX `idx_audit_logs_user_id` ON `audit_logs`(`user_id`);
