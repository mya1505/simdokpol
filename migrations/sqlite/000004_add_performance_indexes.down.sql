DROP INDEX IF EXISTS `idx_lost_documents_tanggal_laporan`;
DROP INDEX IF EXISTS `idx_lost_documents_operator_id`;
DROP INDEX IF EXISTS `idx_lost_documents_resident_id`;
DROP INDEX IF EXISTS `idx_lost_documents_petugas_pelapor_id`;
DROP INDEX IF EXISTS `idx_lost_documents_pejabat_persetuju_id`;
DROP INDEX IF EXISTS `idx_lost_documents_last_updated_by_id`;

DROP INDEX IF EXISTS `idx_lost_items_lost_document_id`;

DROP INDEX IF EXISTS `idx_audit_logs_timestamp`;
DROP INDEX IF EXISTS `idx_audit_logs_user_id`;
