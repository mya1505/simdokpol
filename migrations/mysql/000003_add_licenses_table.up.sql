CREATE TABLE `licenses` (
    `key` VARCHAR(255) PRIMARY KEY,
    `status` VARCHAR(50) NOT NULL,
    `activated_at` DATETIME(3),
    `activated_by_id` INTEGER,
    `notes` TEXT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

INSERT INTO `configurations` (`key`, `value`) 
VALUES ('license_status', 'UNLICENSED') 
ON DUPLICATE KEY UPDATE `key`=`key`;