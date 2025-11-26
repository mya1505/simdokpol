-- +migrate Down

DROP TABLE IF EXISTS `configurations`;
DROP TABLE IF EXISTS `audit_logs`;
DROP TABLE IF EXISTS `lost_items`;
DROP TABLE IF EXISTS `lost_documents`;
DROP TABLE IF EXISTS `residents`;
DROP TABLE IF EXISTS `users`;