/*
 Navicat Premium Data Transfer


*/


SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for message_auth
-- ----------------------------
DROP TABLE IF EXISTS `message_auth`;
CREATE TABLE `message_auth`  (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
  `username` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '账号',
  `password` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '密码',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB  CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;


INSERT INTO `blog`.`message_auth` (`id`, `username`, `password`) VALUES (1, 'admin', '123456');


-- ----------------------------
-- Table structure for message_send_tasks
-- ----------------------------
DROP TABLE IF EXISTS `message_send_tasks`;
CREATE TABLE `message_send_tasks`  (
  `id` varchar(36) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `name` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '任务名称',
  `created_on` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `created_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '创建人',
  `modified_on` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `modified_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '修改人',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '消息任务表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for message_send_tasks_ins
-- ----------------------------
DROP TABLE IF EXISTS `message_send_tasks_ins`;
CREATE TABLE `message_send_tasks_ins`  (
  `id` varchar(36) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `task_id` varchar(36) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '任务id',
  `way_id` varchar(36) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '渠道id',
  `way_type` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '渠道类型',
  `content_type` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '内容类型',
  `config` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '实例配置',
  `extra` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '任务额外信息',
  `created_on` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `created_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '创建人',
  `modified_on` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `modified_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '修改人',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `task_id`(`task_id`) USING BTREE,
  INDEX `way_id`(`way_id`) USING BTREE,
  INDEX `content_type`(`content_type`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '消息任务实例表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for message_send_tasks_logs
-- ----------------------------
DROP TABLE IF EXISTS `message_send_tasks_logs`;
CREATE TABLE `message_send_tasks_logs`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `task_id` varchar(36) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '任务id',
  `log` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '日志内容',
  `status` int(11) NULL DEFAULT NULL COMMENT '任务状态',
  `created_on` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `created_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '创建人',
  `modified_on` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `modified_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '修改人',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `task_id`(`task_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '消息任务日志表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for message_send_ways
-- ----------------------------
DROP TABLE IF EXISTS `message_send_ways`;
CREATE TABLE `message_send_ways`  (
  `id` varchar(36) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL,
  `name` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '消息渠道名称',
  `type` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '类型',
  `auth` varchar(2048) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '认证信息',
  `created_on` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `created_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '创建人',
  `modified_on` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `modified_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '修改人',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `type`(`type`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '消息渠道表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for message_settings
-- ----------------------------
DROP TABLE IF EXISTS `message_settings`;
CREATE TABLE `message_settings`  (
  `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `section` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL COMMENT '分类',
  `key` varchar(50) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '键',
  `value` text CHARACTER SET utf8 COLLATE utf8_general_ci NULL COMMENT '值',
  `created_on` timestamp NULL DEFAULT NULL COMMENT '创建时间',
  `created_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '创建人',
  `modified_on` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  `modified_by` varchar(100) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT '' COMMENT '修改人',
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `key`(`key`) USING BTREE,
  INDEX `section`(`section`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8 COLLATE = utf8_general_ci COMMENT = '设置表' ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
