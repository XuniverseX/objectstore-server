DROP TABLE IF EXISTS `tbl_user_file`;
CREATE TABLE `tbl_user_file`(
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `user_name` varchar(64)  NOT NULL,
    `file_hash` varchar(64)  NOT NULL DEFAULT '' COMMENT '文件hash',
    `file_size` bigint(20) DEFAULT '0' COMMENT '文件大小',
    `file_name` varchar(256) NOT NULL DEFAULT '' COMMENT '文件名',
    `upload_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '上传时间',
    `last_update` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
    `status` int(11)      NOT NULL DEFAULT '0' COMMENT '文件状态(0正常/1已删除/2禁用)',
    PRIMARY KEY (`id`),
#     UNIQUE KEY `idx_user_file` (`user_name`, `file_hash`),
    KEY `idx_status` (`status`),
    KEY `idx_user_id` (`user_name`)
)ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;