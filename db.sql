DROP TABLE IF EXISTS `t_rule`;
CREATE TABLE `t_rule`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT "主键，报警规则ID",
  `expr` varchar(1023) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT "监控指标",
  `op` varchar(31) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT "比较符号",
  `value` varchar(31) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT "报警阈值",
  `for` varchar(1023) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT "持续时间",
  `overdue_value` tinyint(4) NOT NULL DEFAULT '30' COMMENT "告警失效时间(单位：分钟)",
  `summary` varchar(1023) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT "报警标题",
  `description` varchar(1023) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '' COMMENT "报警描述",
  `source_id` bigint(20) NOT NULL COMMENT "数据源ID",
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  COMMENT "创建时间",
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT "最近更新时间",
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact COMMENT "报警规则表";


DROP TABLE IF EXISTS `t_source`;
CREATE TABLE `t_source`  (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `name` varchar(1023) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `url` varchar(1023) CHARACTER SET utf8 COLLATE utf8_general_ci NOT NULL DEFAULT '',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP  COMMENT "创建时间",
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT "最近更新时间",
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Compact;
