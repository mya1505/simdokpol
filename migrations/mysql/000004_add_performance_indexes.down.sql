-- +migrate Down

DROP INDEX `idx_lost_documents_tanggal_laporan` ON `lost_documents`;
DROP INDEX `idx_lost_documents_operator_id` ON `lost_documents`;
DROP INDEX `idx_lost_documents_resident_id` ON `lost_documents`;
DROP INDEX `idx_lost_documents_petugas_pelapor_id` ON `lost_documents`;
DROP INDEX `idx_lost_documents_pejabat_persetuju_id` ON `lost_documents`;
DROP INDEX `idx_lost_documents_last_updated_by_id` ON `lost_documents`;

DROP INDEX `idx_lost_items_lost_document_id` ON `lost_items`;

DROP INDEX `idx_audit_logs_timestamp` ON `audit_logs`;
DROP INDEX `idx_audit_logs_user_id` ON `audit_logs`;
