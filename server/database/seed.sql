
/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

LOCK TABLES `audit_event_records` WRITE;
/*!40000 ALTER TABLE `audit_event_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `audit_event_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `event_deliveries` WRITE;
/*!40000 ALTER TABLE `event_deliveries` DISABLE KEYS */;
/*!40000 ALTER TABLE `event_deliveries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `event_outbox` WRITE;
/*!40000 ALTER TABLE `event_outbox` DISABLE KEYS */;
/*!40000 ALTER TABLE `event_outbox` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_abnormal_bet_events` WRITE;
/*!40000 ALTER TABLE `game_abnormal_bet_events` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_abnormal_bet_events` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_abnormal_bet_notes` WRITE;
/*!40000 ALTER TABLE `game_abnormal_bet_notes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_abnormal_bet_notes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_abnormal_bet_review_histories` WRITE;
/*!40000 ALTER TABLE `game_abnormal_bet_review_histories` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_abnormal_bet_review_histories` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_activity_assets` WRITE;
/*!40000 ALTER TABLE `game_activity_assets` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_activity_assets` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_activity_events` WRITE;
/*!40000 ALTER TABLE `game_activity_events` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_activity_events` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_activity_reward_rules` WRITE;
/*!40000 ALTER TABLE `game_activity_reward_rules` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_activity_reward_rules` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_activity_rules` WRITE;
/*!40000 ALTER TABLE `game_activity_rules` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_activity_rules` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_account_roles` WRITE;
/*!40000 ALTER TABLE `game_admin_account_roles` DISABLE KEYS */;
INSERT INTO `game_admin_account_roles` VALUES (1,1,1,'2026-11-26 22:42:40.343','2026-11-26 22:42:40.343',NULL);
/*!40000 ALTER TABLE `game_admin_account_roles` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_accounts` WRITE;
/*!40000 ALTER TABLE `game_admin_accounts` DISABLE KEYS */;
INSERT INTO `game_admin_accounts` VALUES (1,0,'system-admin','System Administrator','',1,'','',0,'',NULL,1,'',NULL,0,NULL,1,'zh-CN','Asia/Kuala_Lumpur',0,1,'2026-11-26 22:42:40.342','2026-11-26 23:04:50.615',NULL);
/*!40000 ALTER TABLE `game_admin_accounts` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_departments` WRITE;
/*!40000 ALTER TABLE `game_admin_departments` DISABLE KEYS */;
INSERT INTO `game_admin_departments` VALUES (1,NULL,'ROOT','总部','Head Office',NULL,1,1,1,0,'2026-11-26 22:42:40.331','2026-11-26 22:42:40.331',NULL);
/*!40000 ALTER TABLE `game_admin_departments` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_login_logs` WRITE;
/*!40000 ALTER TABLE `game_admin_login_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_admin_login_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_menus` WRITE;
/*!40000 ALTER TABLE `game_admin_menus` DISABLE KEYS */;
INSERT INTO `game_admin_menus` VALUES (1,NULL,'DIRECTORY','system','系统管理','System Management','','','Setting','',900,1,1,0,'',1,0,'2026-11-26 22:42:40.348','2026-11-26 22:42:40.348',NULL);
INSERT INTO `game_admin_menus` VALUES (2,1,'MENU','system-admins','管理员管理','Administrator Management','/system/admins','system/SystemManagementPage','User','system:admin:list',10,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (3,1,'MENU','system-roles','角色管理','Role Management','/system/roles','system/SystemManagementPage','Avatar','system:role:list',20,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (4,1,'MENU','system-permissions','权限管理','Permission Management','/system/permissions','system/SystemManagementPage','Key','system:permission:list',30,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (5,1,'MENU','system-departments','部门管理','Department Management','/system/departments','system/SystemManagementPage','OfficeBuilding','system:department:list',40,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (6,1,'MENU','system-menus','菜单管理','Menu Management','/system/menus','system/SystemManagementPage','Menu','system:menu:list',50,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (7,1,'MENU','system-configs','系统配置','System Configuration','/system/config','system/SystemManagementPage','SetUp','system:config:list',60,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (8,1,'MENU','system-dictionaries','字典管理','Dictionary Management','/system/dictionaries','system/SystemManagementPage','Collection','system:dictionary:list',70,0,0,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:41.519','2026-11-26 22:42:41.519');
INSERT INTO `game_admin_menus` VALUES (9,1,'MENU','system-audit-logs','审计日志','Audit Logs','/system/audit-logs','system/SystemManagementPage','DocumentChecked','system:audit-log:view',80,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (10,1,'MENU','system-login-logs','登录日志','Login Logs','/system/login-logs','system/SystemManagementPage','Clock','system:login-log:view',90,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (11,1,'MENU','system-tasks','定时任务','Scheduled Tasks','/system/tasks','system/SystemManagementPage','Timer','system:task:list',100,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (12,1,'MENU','system-monitoring','系统监控','System Monitoring','/system/monitor','system/SystemManagementPage','Monitor','system:monitoring:view',110,1,1,0,'',1,0,'2026-11-26 22:42:40.350','2026-11-26 22:42:40.350',NULL);
INSERT INTO `game_admin_menus` VALUES (13,1,'MENU','country-config','国家配置','Country Configuration','/system/countries','system/SystemManagementPage','Position','system:country:list',65,1,1,0,'',1,0,'2026-11-26 22:42:40.601','2026-11-26 22:42:40.602',NULL);
INSERT INTO `game_admin_menus` VALUES (14,NULL,'DIRECTORY','operations','运营中心','Operations Center','','','Message','',500,1,1,0,'',1,0,'2026-11-26 22:42:41.818','2026-11-26 22:42:41.818',NULL);
INSERT INTO `game_admin_menus` VALUES (15,14,'MENU','asset-library','素材中心','Asset Library','/operations/assets','operations/AssetLibraryPage','FolderOpened','operation:asset:list',5,1,1,0,'',1,0,'2026-11-26 22:42:41.819','2026-11-26 22:42:41.820',NULL);
INSERT INTO `game_admin_menus` VALUES (17,NULL,'DIRECTORY','support-center','支持中心','Support Center','','','Headset','',850,1,1,0,'',1,0,'2026-11-26 22:42:47.011','2026-11-26 22:42:47.011',NULL);
INSERT INTO `game_admin_menus` VALUES (18,NULL,'DIRECTORY','ops-stability','运维中心','Operations','','','Monitor','',875,1,1,0,'',1,0,'2026-11-26 22:42:47.011','2026-11-26 22:42:47.011',NULL);
INSERT INTO `game_admin_menus` VALUES (19,17,'MENU','support-player-search','玩家查询','Player Search','/support/players','','Search','support:player:read',10,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (20,17,'MENU','support-player-timeline','玩家时间线','Player Timeline','/support/timeline','','Clock','support:timeline:read',20,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (21,17,'MENU','support-order-search','订单查询','Order Search','/support/orders','','Tickets','support:player:read',30,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (22,17,'MENU','support-payment-diagnostics','支付诊断','Payment Diagnostics','/support/payment-diagnostics','','CreditCard','monitor:payment:read',40,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (23,17,'MENU','support-game-diagnostics','游戏诊断','Game Diagnostics','/support/game-diagnostics','','Monitor','monitor:provider:read',50,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (24,17,'MENU','support-cases','支持工单','Support Cases','/support/cases','','ChatLineSquare','support:case:read',60,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (25,17,'MENU','support-exports','导出记录','Export Records','/support/exports','','Download','support:export',70,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (26,18,'MENU','payment-monitor','支付监控','Payment Monitor','/ops/payments','','CreditCard','monitor:payment:read',10,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (27,18,'MENU','provider-monitor','Provider监控','Provider Monitor','/ops/providers','','Odometer','monitor:provider:read',20,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (28,18,'MENU','maintenance-center','维护管理','Maintenance','/ops/maintenance','','Warning','maintenance:site:read',30,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 23:03:46.996',NULL);
INSERT INTO `game_admin_menus` VALUES (29,18,'MENU','release-center','发布中心','Release Center','/ops/releases','','UploadFilled','release:read',40,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (30,18,'MENU','system-health','系统健康','System Health','/ops/health','','FirstAidKit','health:system:read',50,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (31,18,'MENU','jobs-retries','任务与重试','Jobs & Retries','/ops/jobs','','Timer','jobs:read',60,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (32,18,'MENU','alert-center','告警中心','Alert Center','/ops/alerts','','Bell','alerts:read',70,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
INSERT INTO `game_admin_menus` VALUES (33,18,'MENU','backup-center','备份中心','Backup Center','/ops/backups','','FolderOpened','backup:read',80,1,1,0,'',1,0,'2026-11-26 22:42:47.012','2026-11-26 22:42:47.012',NULL);
/*!40000 ALTER TABLE `game_admin_menus` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_mfa_challenges` WRITE;
/*!40000 ALTER TABLE `game_admin_mfa_challenges` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_admin_mfa_challenges` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_mfa_recovery_codes` WRITE;
/*!40000 ALTER TABLE `game_admin_mfa_recovery_codes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_admin_mfa_recovery_codes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_operation_logs` WRITE;
/*!40000 ALTER TABLE `game_admin_operation_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_admin_operation_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_password_history` WRITE;
/*!40000 ALTER TABLE `game_admin_password_history` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_admin_password_history` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_permissions` WRITE;
/*!40000 ALTER TABLE `game_admin_permissions` DISABLE KEYS */;
INSERT INTO `game_admin_permissions` VALUES (1,'user:list','List users','List users','List users','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (2,'user:view','View user details','View user details','View user details','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (3,'user:status:update','Update user status','Update user status','Update user status','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (4,'user:restriction:update','Update user restrictions','Update user restrictions','Update user restrictions','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (5,'user:risk:update','Update user risk profile','Update user risk profile','Update user risk profile','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (6,'user:note:create','Create user internal notes','Create user internal notes','Create user internal notes','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (7,'user:login-record:view','View login records','View login records','View login records','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (8,'user:device:view','View devices','View devices','View devices','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (9,'user:device:manage','Manage devices','Manage devices','Manage devices','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (10,'user:sensitive:view','View sensitive user data','View sensitive user data','View sensitive user data','user','API','',1,1,'2026-11-26 22:42:38.948','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (11,'bet:list','List bets','List bets','List bets','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (12,'bet:view','View bet details','View bet details','View bet details','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (13,'bet:review','Mark bet for review','Mark bet for review','Mark bet for review','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (14,'bet:round:list','List game rounds','List game rounds','List game rounds','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (15,'bet:round:view','View game round details','View game round details','View game round details','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (16,'bet:rollback:list','List rollback records','List rollback records','List rollback records','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (17,'bet:rollback:view','View rollback details','View rollback details','View rollback details','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (18,'bet:rollback:review','Review rollback records','Review rollback records','Review rollback records','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (19,'bet:abnormal:list','List abnormal bets','List abnormal bets','List abnormal bets','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (20,'bet:abnormal:view','View abnormal bet details','View abnormal bet details','View abnormal bet details','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (21,'bet:abnormal:assign','Assign abnormal bet reviewer','Assign abnormal bet reviewer','Assign abnormal bet reviewer','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (22,'bet:abnormal:review','Review abnormal bets','Review abnormal bets','Review abnormal bets','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (23,'bet:abnormal:clear','Clear abnormal bets','Clear abnormal bets','Clear abnormal bets','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (24,'bet:abnormal:restrict','Restrict abnormal bet users','Restrict abnormal bet users','Restrict abnormal bet users','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (25,'bet:abnormal:note','Add abnormal bet notes','Add abnormal bet notes','Add abnormal bet notes','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (26,'bet:callback-log:list','List provider callback logs','List provider callback logs','List provider callback logs','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (27,'bet:callback-log:view','View provider callback log details','View provider callback log details','View provider callback log details','bet','API','',1,1,'2026-11-26 22:42:39.190','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (28,'fund:wallet:list','List wallet accounts','List wallet accounts','List wallet accounts','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (29,'fund:wallet:view','View wallet account details','View wallet account details','View wallet account details','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (30,'fund:wallet:freeze','Freeze wallet accounts','Freeze wallet accounts','Freeze wallet accounts','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (31,'fund:wallet:unfreeze','Unfreeze wallet accounts','Unfreeze wallet accounts','Unfreeze wallet accounts','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (32,'fund:transaction:list','List wallet transactions','List wallet transactions','List wallet transactions','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (33,'fund:transaction:view','View wallet transaction details','View wallet transaction details','View wallet transaction details','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (34,'fund:adjustment:list','List manual adjustments','List manual adjustments','List manual adjustments','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (35,'fund:adjustment:view','View manual adjustments','View manual adjustments','View manual adjustments','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (36,'fund:adjustment:create','Create manual adjustments','Create manual adjustments','Create manual adjustments','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (37,'fund:adjustment:approve','Approve manual adjustments','Approve manual adjustments','Approve manual adjustments','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (38,'fund:adjustment:reject','Reject manual adjustments','Reject manual adjustments','Reject manual adjustments','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (39,'fund:adjustment:cancel','Cancel manual adjustments','Cancel manual adjustments','Cancel manual adjustments','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (40,'fund:adjustment:large-amount','Approve large manual adjustments','Approve large manual adjustments','Approve large manual adjustments','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (41,'fund:deposit:list','List deposit orders','List deposit orders','List deposit orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (42,'fund:deposit:view','View deposit orders','View deposit orders','View deposit orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (43,'fund:deposit:review','Review deposit orders','Review deposit orders','Review deposit orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (44,'fund:deposit:confirm','Confirm reviewed deposit orders','Confirm reviewed deposit orders','Confirm reviewed deposit orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (45,'fund:deposit:cancel','Cancel deposit orders','Cancel deposit orders','Cancel deposit orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (46,'fund:withdrawal:list','List withdrawal orders','List withdrawal orders','List withdrawal orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (47,'fund:withdrawal:view','View withdrawal orders','View withdrawal orders','View withdrawal orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (48,'fund:withdrawal:risk-review','Perform withdrawal risk review','Perform withdrawal risk review','Perform withdrawal risk review','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (49,'fund:withdrawal:finance-review','Perform withdrawal finance review','Perform withdrawal finance review','Perform withdrawal finance review','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (50,'fund:withdrawal:approve','Approve withdrawal orders','Approve withdrawal orders','Approve withdrawal orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (51,'fund:withdrawal:reject','Reject withdrawal orders','Reject withdrawal orders','Reject withdrawal orders','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (52,'fund:withdrawal:retry','Retry failed withdrawal payouts','Retry failed withdrawal payouts','Retry failed withdrawal payouts','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (53,'fund:withdrawal:sensitive-view','View sensitive withdrawal information','View sensitive withdrawal information','View sensitive withdrawal information','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (54,'fund:reconciliation:list','List reconciliations','List reconciliations','List reconciliations','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (55,'fund:reconciliation:view','View reconciliation details','View reconciliation details','View reconciliation details','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (56,'fund:reconciliation:generate','Generate reconciliations','Generate reconciliations','Generate reconciliations','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (57,'fund:reconciliation:resolve','Resolve reconciliation differences','Resolve reconciliation differences','Resolve reconciliation differences','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (58,'fund:reconciliation:export','Export reconciliation differences','Export reconciliation differences','Export reconciliation differences','fund','API','',1,1,'2026-11-26 22:42:39.414','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (59,'operation:banner:list','List banners','List banners','List banners','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (60,'operation:banner:create','Create banners','Create banners','Create banners','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (61,'operation:banner:update','Update banners','Update banners','Update banners','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (62,'operation:banner:delete','Delete banners','Delete banners','Delete banners','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (63,'operation:announcement:list','List announcements','List announcements','List announcements','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (64,'operation:announcement:create','Create announcements','Create announcements','Create announcements','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (65,'operation:announcement:update','Update announcements','Update announcements','Update announcements','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (66,'operation:announcement:delete','Delete announcements','Delete announcements','Delete announcements','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (67,'operation:promotion:list','List promotions','List promotions','List promotions','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (68,'operation:promotion:create','Create promotions','Create promotions','Create promotions','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (69,'operation:promotion:update','Update promotions','Update promotions','Update promotions','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (70,'operation:promotion:status','Change promotion status','Change promotion status','Change promotion status','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (71,'operation:bonus:list','List bonuses','List bonuses','List bonuses','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (72,'operation:bonus:view','View bonuses','View bonuses','View bonuses','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (73,'operation:bonus:activate','Activate bonuses','Activate bonuses','Activate bonuses','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (74,'operation:bonus:cancel','Cancel bonuses','Cancel bonuses','Cancel bonuses','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (75,'operation:bonus:forfeit','Forfeit bonuses','Forfeit bonuses','Forfeit bonuses','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (76,'operation:vip:list','List VIP configuration','List VIP configuration','List VIP configuration','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (77,'operation:vip:update','Update VIP configuration','Update VIP configuration','Update VIP configuration','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (78,'operation:vip:adjust','Adjust user VIP','Adjust user VIP','Adjust user VIP','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (79,'operation:promo-code:list','List promo codes','List promo codes','List promo codes','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (80,'operation:promo-code:create','Create promo codes','Create promo codes','Create promo codes','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (81,'operation:promo-code:update','Update promo codes','Update promo codes','Update promo codes','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (82,'operation:promo-code:delete','Delete promo codes','Delete promo codes','Delete promo codes','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (83,'operation:notification:list','List notifications','List notifications','List notifications','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (84,'operation:notification:create','Create notifications','Create notifications','Create notifications','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (85,'operation:notification:update','Update notifications','Update notifications','Update notifications','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (86,'operation:notification:send','Send notifications','Send notifications','Send notifications','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (87,'operation:notification:cancel','Cancel notifications','Cancel notifications','Cancel notifications','operation','API','',1,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (88,'operation:popup:list','List popups','List popups','List popups','operation','API','',0,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:46.863',NULL);
INSERT INTO `game_admin_permissions` VALUES (89,'operation:popup:create','Create popups','Create popups','Create popups','operation','API','',0,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:46.863',NULL);
INSERT INTO `game_admin_permissions` VALUES (90,'operation:popup:update','Update popups','Update popups','Update popups','operation','API','',0,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:46.863',NULL);
INSERT INTO `game_admin_permissions` VALUES (91,'operation:popup:delete','Delete popups','Delete popups','Delete popups','operation','API','',0,1,'2026-11-26 22:42:39.572','2026-11-26 22:42:46.863',NULL);
INSERT INTO `game_admin_permissions` VALUES (92,'risk:dashboard:view','View risk dashboard','View risk dashboard','View risk dashboard','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (93,'risk:rule:list','List risk rules','List risk rules','List risk rules','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (94,'risk:rule:create','Create risk rules','Create risk rules','Create risk rules','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (95,'risk:rule:update','Update risk rules','Update risk rules','Update risk rules','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (96,'risk:rule:delete','Delete risk rules','Delete risk rules','Delete risk rules','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (97,'risk:rule:test','Test risk rules','Test risk rules','Test risk rules','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (98,'risk:rule:status','Update risk rule status','Update risk rule status','Update risk rule status','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (99,'risk:event:list','List risk events','List risk events','List risk events','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (100,'risk:event:view','View risk events','View risk events','View risk events','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (101,'risk:event:assign','Assign risk events','Assign risk events','Assign risk events','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (102,'risk:event:review','Review risk events','Review risk events','Review risk events','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (103,'risk:event:clear','Clear risk events','Clear risk events','Clear risk events','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (104,'risk:event:restrict','Restrict users','Restrict users','Restrict users','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (105,'risk:event:block','Block users','Block users','Block users','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (106,'risk:event:note','Add risk notes','Add risk notes','Add risk notes','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (107,'risk:review:list','List risk reviews','List risk reviews','List risk reviews','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (108,'risk:review:view','View risk reviews','View risk reviews','View risk reviews','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (109,'risk:review:claim','Claim risk reviews','Claim risk reviews','Claim risk reviews','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (110,'risk:review:assign','Assign risk reviews','Assign risk reviews','Assign risk reviews','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (111,'risk:review:decision','Decide risk reviews','Decide risk reviews','Decide risk reviews','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (112,'risk:review:escalate','Escalate risk reviews','Escalate risk reviews','Escalate risk reviews','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (113,'risk:association:list','List user associations','List user associations','List user associations','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (114,'risk:association:view','View user associations','View user associations','View user associations','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (115,'risk:blacklist:list','List blacklists','List blacklists','List blacklists','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (116,'risk:blacklist:create','Create blacklists','Create blacklists','Create blacklists','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (117,'risk:blacklist:update','Update blacklists','Update blacklists','Update blacklists','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (118,'risk:blacklist:remove','Remove blacklists','Remove blacklists','Remove blacklists','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (119,'risk:withdrawal:list','List risk withdrawals','List risk withdrawals','List risk withdrawals','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (120,'risk:withdrawal:view','View risk withdrawals','View risk withdrawals','View risk withdrawals','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (121,'risk:withdrawal:approve','Approve withdrawal risk review','Approve withdrawal risk review','Approve withdrawal risk review','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (122,'risk:withdrawal:reject','Reject withdrawal risk review','Reject withdrawal risk review','Reject withdrawal risk review','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (123,'risk:withdrawal:restrict','Restrict withdrawal','Restrict withdrawal','Restrict withdrawal','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (124,'risk:operation-log:view','View risk operation logs','View risk operation logs','View risk operation logs','risk','API','',1,1,'2026-11-26 22:42:39.705','2026-11-26 22:42:40.038',NULL);
INSERT INTO `game_admin_permissions` VALUES (125,'system:admin:list','List administrators','管理员列表','List administrators','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (126,'system:admin:view','View administrator','查看管理员','View administrator','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (127,'system:admin:create','Create administrator','创建管理员','Create administrator','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (128,'system:admin:update','Update administrator','更新管理员','Update administrator','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (129,'system:admin:status','Update administrator status','管理员状态','Update administrator status','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (130,'system:admin:reset-password','Reset administrator password','重置管理员密码','Reset administrator password','system','ACTION','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (131,'system:admin:reset-mfa','Reset administrator MFA','重置管理员 MFA','Reset administrator MFA','system','ACTION','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (132,'system:admin:force-logout','Force administrator logout','强制管理员退出','Force administrator logout','system','ACTION','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (133,'system:admin:assign-role','Assign administrator roles','分配管理员角色','Assign administrator roles','system','ACTION','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (134,'system:role:list','List roles','角色列表','List roles','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (135,'system:role:create','Create role','创建角色','Create role','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (136,'system:role:update','Update role','更新角色','Update role','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (137,'system:role:delete','Delete role','删除角色','Delete role','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (138,'system:role:assign-permission','Assign role permissions','分配角色权限','Assign role permissions','system','ACTION','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (139,'system:role:assign-menu','Assign role menus','分配角色菜单','Assign role menus','system','ACTION','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (140,'system:permission:list','List permissions','权限列表','List permissions','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (141,'system:permission:create','Create permission','创建权限','Create permission','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (142,'system:permission:update','Update permission','更新权限','Update permission','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (143,'system:permission:delete','Delete permission','删除权限','Delete permission','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (144,'system:department:list','List departments','部门列表','List departments','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (145,'system:department:create','Create department','创建部门','Create department','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (146,'system:department:update','Update department','更新部门','Update department','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (147,'system:department:delete','Delete department','删除部门','Delete department','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (148,'system:menu:list','List menus','菜单列表','List menus','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (149,'system:menu:create','Create menu','创建菜单','Create menu','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (150,'system:menu:update','Update menu','更新菜单','Update menu','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (151,'system:menu:delete','Delete menu','删除菜单','Delete menu','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (152,'system:config:list','List system configuration','系统配置列表','List system configuration','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (153,'system:config:update','Update system configuration','更新系统配置','Update system configuration','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (154,'system:config:secret-update','Update secret configuration','更新密钥配置','Update secret configuration','system','ACTION','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (155,'system:dictionary:list','List dictionaries','字典列表','List dictionaries','system','API','',0,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:41.517','2026-11-26 22:42:41.517');
INSERT INTO `game_admin_permissions` VALUES (156,'system:dictionary:create','Create dictionary','创建字典','Create dictionary','system','API','',0,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:41.517','2026-11-26 22:42:41.517');
INSERT INTO `game_admin_permissions` VALUES (157,'system:dictionary:update','Update dictionary','更新字典','Update dictionary','system','API','',0,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:41.517','2026-11-26 22:42:41.517');
INSERT INTO `game_admin_permissions` VALUES (158,'system:dictionary:delete','Delete dictionary','删除字典','Delete dictionary','system','API','',0,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:41.517','2026-11-26 22:42:41.517');
INSERT INTO `game_admin_permissions` VALUES (159,'system:audit-log:view','View audit logs','查看审计日志','View audit logs','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (160,'system:login-log:view','View administrator login logs','查看登录日志','View administrator login logs','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (161,'system:task:list','List scheduled tasks','定时任务列表','List scheduled tasks','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (162,'system:task:create','Create scheduled task','创建定时任务','Create scheduled task','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (163,'system:task:update','Update scheduled task','更新定时任务','Update scheduled task','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (164,'system:task:delete','Delete scheduled task','删除定时任务','Delete scheduled task','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (165,'system:task:run','Run scheduled task','运行定时任务','Run scheduled task','system','ACTION','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (166,'system:monitoring:view','View system monitoring','查看系统监控','View system monitoring','system','API','',1,1,'2026-11-26 22:42:40.344','2026-11-26 22:42:40.344',NULL);
INSERT INTO `game_admin_permissions` VALUES (167,'agent:list','List agents','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (168,'agent:view','View agent','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (169,'agent:create','Create agent','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (170,'agent:update','Update agent','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (171,'agent:status','Update agent status','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (172,'agent:hierarchy:view','View agent hierarchy','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (173,'agent:hierarchy:update','Update agent hierarchy','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (174,'agent:plan:assign','Assign commission plan','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (175,'agent:channel:list','List agent channels','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (176,'agent:channel:create','Create agent channel','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (177,'agent:channel:update','Update agent channel','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (178,'agent:channel:status','Update agent channel status','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (179,'agent:commission-plan:list','List commission plans','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (180,'agent:commission-plan:create','Create commission plan','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (181,'agent:commission-plan:update','Update commission plan','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (182,'agent:commission-plan:delete','Delete commission plan','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (183,'agent:commission-record:list','List commission records','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (184,'agent:commission-record:view','View commission record','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (185,'agent:commission-record:adjust','Adjust commission record','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (186,'agent:settlement:list','List agent settlements','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (187,'agent:settlement:view','View agent settlement','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (188,'agent:settlement:generate','Generate agent settlement','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (189,'agent:settlement:submit','Submit agent settlement','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (190,'agent:settlement:approve','Approve agent settlement','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (191,'agent:settlement:reject','Reject agent settlement','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (192,'agent:settlement:pay','Pay agent settlement','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (193,'agent:settlement:cancel','Cancel agent settlement','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (194,'agent:settlement:export','Export agent settlement','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (195,'agent:wallet:list','List agent wallets','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (196,'agent:wallet:view','View agent wallet','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (197,'agent:wallet:freeze','Freeze agent wallet','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (198,'agent:wallet:adjust','Adjust agent wallet','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (199,'agent:operation-log:view','View agent operation logs','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (200,'agent:attribution:update','Update user agent attribution','','','','API','',1,0,'2026-11-26 22:42:40.522','2026-11-26 22:42:40.522',NULL);
INSERT INTO `game_admin_permissions` VALUES (201,'dashboard:view','View BI dashboard','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (202,'dashboard:financial:view','View dashboard financial data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (203,'dashboard:user:view','View dashboard user data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (204,'dashboard:game:view','View dashboard game data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (205,'dashboard:provider:view','View dashboard provider data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (206,'dashboard:channel:view','View dashboard channel data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (207,'dashboard:agent:view','View dashboard agent data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (208,'dashboard:promotion:view','View dashboard promotion data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (209,'dashboard:risk:view','View dashboard risk data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (210,'dashboard:realtime:view','View dashboard realtime data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (211,'dashboard:alert:view','View dashboard alerts','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (212,'dashboard:sensitive:view','View dashboard sensitive data','','','','API','',1,0,'2026-11-26 22:42:40.563','2026-11-26 22:42:40.563',NULL);
INSERT INTO `game_admin_permissions` VALUES (213,'system:country:list','List country configuration','国家配置列表','List country configuration','system','API','',1,1,'2026-11-26 22:42:40.599','2026-11-26 22:42:40.599',NULL);
INSERT INTO `game_admin_permissions` VALUES (214,'system:country:view','View country configuration','查看国家配置','View country configuration','system','API','',1,1,'2026-11-26 22:42:40.599','2026-11-26 22:42:40.599',NULL);
INSERT INTO `game_admin_permissions` VALUES (215,'system:country:create','Create country configuration','创建国家配置','Create country configuration','system','API','',1,1,'2026-11-26 22:42:40.599','2026-11-26 22:42:40.599',NULL);
INSERT INTO `game_admin_permissions` VALUES (216,'system:country:update','Update country configuration','更新国家配置','Update country configuration','system','API','',1,1,'2026-11-26 22:42:40.599','2026-11-26 22:42:40.599',NULL);
INSERT INTO `game_admin_permissions` VALUES (217,'system:country:delete','Delete country configuration','删除国家配置','Delete country configuration','system','ACTION','',1,1,'2026-11-26 22:42:40.599','2026-11-26 22:42:40.599',NULL);
INSERT INTO `game_admin_permissions` VALUES (218,'system:country:status','Update country configuration status','国家配置状态','Update country configuration status','system','ACTION','',1,1,'2026-11-26 22:42:40.599','2026-11-26 22:42:40.599',NULL);
INSERT INTO `game_admin_permissions` VALUES (219,'system:country:branding','Update country H5 branding','配置国家 H5 品牌','Update country H5 branding','system','ACTION','',1,1,'2026-11-26 22:42:40.599','2026-11-26 22:42:40.599',NULL);
INSERT INTO `game_admin_permissions` VALUES (220,'system:country:set-default','Set default country','设置默认国家','Set default country','system','ACTION','',1,1,'2026-11-26 22:42:40.599','2026-11-26 22:42:40.599',NULL);
INSERT INTO `game_admin_permissions` VALUES (222,'fund:deposit-method:list','Deposit Method List','充值方式列表','Deposit Method List','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (223,'fund:deposit-method:create','Create Deposit Method','创建充值方式','Create Deposit Method','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (224,'fund:deposit-method:update','Update Deposit Method','更新充值方式','Update Deposit Method','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (225,'fund:deposit-method:delete','Delete Deposit Method','删除充值方式','Delete Deposit Method','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (226,'fund:deposit-channel:list','Deposit Channel List','充值渠道列表','Deposit Channel List','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (227,'fund:deposit-channel:create','Create Deposit Channel','创建充值渠道','Create Deposit Channel','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (228,'fund:deposit-channel:update','Update Deposit Channel','更新充值渠道','Update Deposit Channel','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (229,'fund:deposit-channel:delete','Delete Deposit Channel','删除充值渠道','Delete Deposit Channel','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (230,'fund:deposit-option:list','Deposit Amount Option List','充值金额档位列表','Deposit Amount Option List','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (231,'fund:deposit-option:create','Create Deposit Amount Option','创建充值金额档位','Create Deposit Amount Option','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (232,'fund:deposit-option:update','Update Deposit Amount Option','更新充值金额档位','Update Deposit Amount Option','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (233,'fund:deposit-option:delete','Delete Deposit Amount Option','删除充值金额档位','Delete Deposit Amount Option','fund','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (234,'operation:voucher:list','Voucher List','优惠券列表','Voucher List','operation','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (235,'operation:voucher:create','Create Voucher','创建优惠券','Create Voucher','operation','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (236,'operation:voucher:update','Update Voucher','更新优惠券','Update Voucher','operation','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (237,'operation:voucher:delete','Delete Voucher','删除优惠券','Delete Voucher','operation','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (238,'operation:voucher:tier:update','Configure Voucher Tiers','配置优惠券奖励档位','Configure Voucher Tiers','operation','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (239,'operation:voucher:usage:view','View Voucher Usage','查看优惠券使用记录','View Voucher Usage','operation','API','',1,1,'2026-11-26 22:42:40.691','2026-11-26 22:42:40.691',NULL);
INSERT INTO `game_admin_permissions` VALUES (240,'game:asset:upload','Upload game asset','','','','API','',1,0,'2026-11-26 22:42:40.713','2026-11-26 22:42:40.713',NULL);
INSERT INTO `game_admin_permissions` VALUES (241,'game:provider:list','List game providers','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (242,'game:provider:view','View game provider','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (243,'game:provider:update','Update game provider','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (244,'game:provider:test','Test game provider','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (245,'game:provider:sync','Synchronize game provider','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (246,'game:list','List games','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (247,'game:view','View game','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (248,'game:update','Update game','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (249,'game:status:update','Update game status','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (250,'game:category:assign','Assign game categories','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (251,'game:tag:assign','Assign game tags','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (252,'game:collection:assign','Assign game collections','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (253,'game:demo:launch','Launch game demo','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (254,'game:category:list','List game categories','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (255,'game:category:create','Create game category','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (256,'game:category:update','Update game category','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (257,'game:category:delete','Delete game category','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (258,'game:tag:list','List game tags','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (259,'game:tag:create','Create game tag','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (260,'game:tag:update','Update game tag','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (261,'game:tag:delete','Delete game tag','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (262,'game:collection:list','List game collections','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (263,'game:collection:create','Create game collection','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (264,'game:collection:update','Update game collection','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (265,'game:collection:delete','Delete game collection','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (266,'game:sync-log:view','View game synchronization logs','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (267,'game:launch-log:view','View game launch logs','','','','API','',1,0,'2026-11-26 22:42:40.827','2026-11-26 22:42:40.827',NULL);
INSERT INTO `game_admin_permissions` VALUES (268,'operation:music:list','Music List','音乐列表','Music List','operation','API','',1,1,'2026-11-26 22:42:41.141','2026-11-26 22:42:41.141',NULL);
INSERT INTO `game_admin_permissions` VALUES (269,'operation:music:create','Create Music','创建音乐','Create Music','operation','API','',1,1,'2026-11-26 22:42:41.141','2026-11-26 22:42:41.141',NULL);
INSERT INTO `game_admin_permissions` VALUES (270,'operation:music:update','Update Music','更新音乐','Update Music','operation','API','',1,1,'2026-11-26 22:42:41.141','2026-11-26 22:42:41.141',NULL);
INSERT INTO `game_admin_permissions` VALUES (271,'operation:music:delete','Delete Music','删除音乐','Delete Music','operation','API','',1,1,'2026-11-26 22:42:41.141','2026-11-26 22:42:41.141',NULL);
INSERT INTO `game_admin_permissions` VALUES (272,'operation:asset:upload','Upload Operation Asset','上传运营资源','Upload Operation Asset','operation','API','',1,1,'2026-11-26 22:42:41.141','2026-11-26 22:42:41.816',NULL);
INSERT INTO `game_admin_permissions` VALUES (273,'report:financial:view','View financial reports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (274,'report:game:view','View game reports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (275,'report:provider:view','View provider reports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (276,'report:user:view','View user reports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (277,'report:channel:view','View channel reports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (278,'report:agent:view','View agent reports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (279,'report:promotion:view','View promotion reports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (280,'report:export:create','Create report exports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (281,'report:export:list','List report exports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (282,'report:export:download','Download report exports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (283,'report:export:retry','Retry report exports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (284,'report:export:cancel','Cancel report exports','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (285,'report:sensitive:view','View sensitive report data','','','','API','',1,0,'2026-11-26 22:42:41.249','2026-11-26 22:42:41.249',NULL);
INSERT INTO `game_admin_permissions` VALUES (286,'operation:register-reward:view','View Register Reward Config','查看注册奖励配置','View Register Reward Config','operation','API','',1,1,'2026-11-26 22:42:41.325','2026-11-26 22:42:41.325',NULL);
INSERT INTO `game_admin_permissions` VALUES (287,'operation:register-reward:update','Update Register Reward Config','更新注册奖励配置','Update Register Reward Config','operation','API','',1,1,'2026-11-26 22:42:41.325','2026-11-26 22:42:41.325',NULL);
INSERT INTO `game_admin_permissions` VALUES (295,'system:site:view','View H5 site','查看站点详情','View H5 site','system','API','',1,1,'2026-11-26 22:42:41.744','2026-11-26 23:03:46.995',NULL);
INSERT INTO `game_admin_permissions` VALUES (297,'system:site:update','Update H5 site','更新站点','Update H5 site','system','API','',1,1,'2026-11-26 22:42:41.744','2026-11-26 23:03:46.995',NULL);
INSERT INTO `game_admin_permissions` VALUES (298,'system:site:status','Update H5 site status','变更站点状态','Update H5 site status','system','API','',1,1,'2026-11-26 22:42:41.744','2026-11-26 23:03:46.995',NULL);
INSERT INTO `game_admin_permissions` VALUES (300,'operation:asset:list','List Assets','素材列表','List Assets','operation','API','',1,1,'2026-11-26 22:42:41.815','2026-11-26 22:42:41.815',NULL);
INSERT INTO `game_admin_permissions` VALUES (301,'operation:asset:update','Update Asset','更新素材','Update Asset','operation','API','',1,1,'2026-11-26 22:42:41.815','2026-11-26 22:42:41.815',NULL);
INSERT INTO `game_admin_permissions` VALUES (302,'operation:asset:delete','Delete Asset','删除素材','Delete Asset','operation','API','',1,1,'2026-11-26 22:42:41.815','2026-11-26 22:42:41.815',NULL);
INSERT INTO `game_admin_permissions` VALUES (304,'system:event:list','List event outbox','查看事件中心','List event outbox','system','API','',1,1,'2026-11-26 22:42:41.864','2026-11-26 22:42:41.864',NULL);
INSERT INTO `game_admin_permissions` VALUES (305,'system:event:view','View event detail','查看事件详情','View event detail','system','API','',1,1,'2026-11-26 22:42:41.864','2026-11-26 22:42:41.864',NULL);
INSERT INTO `game_admin_permissions` VALUES (306,'system:event:dispatch','Dispatch pending events','分发事件','Dispatch pending events','system','API','',1,1,'2026-11-26 22:42:41.864','2026-11-26 22:42:41.864',NULL);
INSERT INTO `game_admin_permissions` VALUES (307,'system:event:replay','Replay event outbox','回放事件','Replay event outbox','system','API','',1,1,'2026-11-26 22:42:42.110','2026-11-26 22:42:42.110',NULL);
INSERT INTO `game_admin_permissions` VALUES (314,'lobby:list','List lobby sections','','','','API','',1,0,'2026-11-26 22:42:42.412','2026-11-26 22:42:42.412',NULL);
INSERT INTO `game_admin_permissions` VALUES (315,'lobby:view','View lobby section','','','','API','',1,0,'2026-11-26 22:42:42.412','2026-11-26 22:42:42.412',NULL);
INSERT INTO `game_admin_permissions` VALUES (316,'lobby:update','Update lobby sections','','','','API','',1,0,'2026-11-26 22:42:42.412','2026-11-26 22:42:42.412',NULL);
INSERT INTO `game_admin_permissions` VALUES (317,'system:site:config:preview','Preview site config','预览站点配置','Preview site config','system','API','',1,1,'2026-11-26 22:42:45.214','2026-11-26 22:42:45.214',NULL);
INSERT INTO `game_admin_permissions` VALUES (318,'system:site:config:publish','Publish site config','发布站点配置','Publish site config','system','API','',1,1,'2026-11-26 22:42:45.214','2026-11-26 22:42:45.214',NULL);
INSERT INTO `game_admin_permissions` VALUES (319,'bet:reconciliation:list','List provider reconciliation runs','','','','API','',1,0,'2026-11-26 22:42:46.493','2026-11-26 22:42:46.493',NULL);
INSERT INTO `game_admin_permissions` VALUES (320,'bet:reconciliation:view','View provider reconciliation details','','','','API','',1,0,'2026-11-26 22:42:46.493','2026-11-26 22:42:46.493',NULL);
INSERT INTO `game_admin_permissions` VALUES (321,'bet:reconciliation:generate','Generate provider reconciliation','','','','API','',1,0,'2026-11-26 22:42:46.493','2026-11-26 22:42:46.493',NULL);
INSERT INTO `game_admin_permissions` VALUES (322,'bet:reconciliation:resolve','Resolve provider reconciliation item','','','','API','',1,0,'2026-11-26 22:42:46.493','2026-11-26 22:42:46.493',NULL);
INSERT INTO `game_admin_permissions` VALUES (323,'operations:announcement:view','View site announcements','查看站点公告','View site announcements','operations','API','',1,1,'2026-11-26 22:42:46.574','2026-11-26 22:42:46.574',NULL);
INSERT INTO `game_admin_permissions` VALUES (324,'operations:announcement:edit','Edit site announcements','编辑站点公告','Edit site announcements','operations','API','',1,1,'2026-11-26 22:42:46.574','2026-11-26 22:42:46.574',NULL);
INSERT INTO `game_admin_permissions` VALUES (325,'operations:announcement:delete','Delete site announcements','删除站点公告','Delete site announcements','operations','API','',1,1,'2026-11-26 22:42:46.574','2026-11-26 22:42:46.574',NULL);
INSERT INTO `game_admin_permissions` VALUES (326,'operations:popup:view','View site popups','查看站点弹窗','View site popups','operations','API','',1,1,'2026-11-26 22:42:46.574','2026-11-26 22:42:46.574',NULL);
INSERT INTO `game_admin_permissions` VALUES (327,'operations:popup:edit','Edit site popups','编辑站点弹窗','Edit site popups','operations','API','',1,1,'2026-11-26 22:42:46.574','2026-11-26 22:42:46.574',NULL);
INSERT INTO `game_admin_permissions` VALUES (328,'operations:popup:delete','Delete site popups','删除站点弹窗','Delete site popups','operations','API','',1,1,'2026-11-26 22:42:46.574','2026-11-26 22:42:46.574',NULL);
INSERT INTO `game_admin_permissions` VALUES (334,'system:popup-component:list','List popup components','查看弹窗组件','List Popup Components','system','API','/api/v1/admin/system/popup-components',1,1,'2026-11-26 22:42:46.860','2026-11-26 22:42:46.860',NULL);
INSERT INTO `game_admin_permissions` VALUES (335,'system:popup-component:create','Create popup component','登记弹窗组件','Create Popup Component','system','API','/api/v1/admin/system/popup-components',1,1,'2026-11-26 22:42:46.860','2026-11-26 22:42:46.860',NULL);
INSERT INTO `game_admin_permissions` VALUES (336,'system:popup-component:update','Update popup component','编辑弹窗组件','Update Popup Component','system','API','/api/v1/admin/system/popup-components/:id',1,1,'2026-11-26 22:42:46.860','2026-11-26 22:42:46.860',NULL);
INSERT INTO `game_admin_permissions` VALUES (337,'system:popup-component:delete','Delete popup component','删除弹窗组件','Delete Popup Component','system','API','/api/v1/admin/system/popup-components/:id',1,1,'2026-11-26 22:42:46.860','2026-11-26 22:42:46.860',NULL);
INSERT INTO `game_admin_permissions` VALUES (338,'support:player:read','Read player support data','玩家支持查询','Read Player Support Data','support','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (339,'support:timeline:read','Read player timeline','玩家时间线','Read Player Timeline','support','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (340,'support:case:read','Read support cases','查看支持工单','Read Support Cases','support','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (341,'support:case:write','Manage support cases','管理支持工单','Manage Support Cases','support','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (342,'support:export','Export support records','导出支持记录','Export Support Records','support','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (343,'monitor:payment:read','Read payment monitor','支付监控','Read Payment Monitor','monitor','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (344,'monitor:payment:review','Mark payment for review','标记支付人工复核','Mark Payment for Review','monitor','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (345,'monitor:provider:read','Read provider monitor','Provider监控','Read Provider Monitor','monitor','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (346,'operations:note:write','Write operational notes','添加内部运营备注','Write Operational Notes','operations','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (347,'maintenance:site:read','Read site maintenance','查看站点维护','Read Site Maintenance','maintenance','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 23:06:42.259',NULL);
INSERT INTO `game_admin_permissions` VALUES (348,'maintenance:site:write','Manage site maintenance','管理站点维护','Manage Site Maintenance','maintenance','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 23:06:42.259',NULL);
INSERT INTO `game_admin_permissions` VALUES (349,'maintenance:provider:read','Read provider maintenance','查看Provider维护','Read Provider Maintenance','maintenance','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (350,'maintenance:provider:write','Manage provider maintenance','管理Provider维护','Manage Provider Maintenance','maintenance','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (351,'maintenance:game:read','Read game maintenance','查看游戏维护','Read Game Maintenance','maintenance','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (352,'maintenance:game:write','Manage game maintenance','管理游戏维护','Manage Game Maintenance','maintenance','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (353,'health:system:read','Read system health','系统健康','Read System Health','health','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (354,'jobs:read','Read jobs and retries','查看任务与重试','Read Jobs and Retries','jobs','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (355,'jobs:retry','Retry failed work','重试失败任务','Retry Failed Work','jobs','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (356,'jobs:execute','Execute supported jobs','手动执行任务','Execute Supported Jobs','jobs','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (357,'alerts:read','Read operational alerts','查看告警','Read Operational Alerts','alerts','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (358,'alerts:acknowledge','Acknowledge alerts','确认告警','Acknowledge Alerts','alerts','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (359,'alerts:resolve','Resolve alerts','解决告警','Resolve Alerts','alerts','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (360,'backup:read','Read backup tasks','查看备份','Read Backup Tasks','backup','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (361,'backup:create','Create backup tasks','创建备份','Create Backup Tasks','backup','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (362,'backup:restore','Restore verified backup','恢复备份','Restore Verified Backup','backup','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (363,'backup:delete','Delete backup metadata','删除备份','Delete Backup Metadata','backup','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (364,'release:read','Read configuration releases','查看发布记录','Read Configuration Releases','release','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (365,'release:publish','Publish configuration releases','发布配置','Publish Configuration Releases','release','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (366,'release:rollback','Rollback configuration releases','回滚配置','Rollback Configuration Releases','release','API','',1,1,'2026-11-26 22:42:47.008','2026-11-26 22:42:47.008',NULL);
INSERT INTO `game_admin_permissions` VALUES (367,'operation:social-link:list','List social links','查看客服与社媒','View social links','operations','API','/api/v1/admin/social-links',1,1,'2026-11-26 22:42:47.307','2026-11-26 22:42:47.307',NULL);
INSERT INTO `game_admin_permissions` VALUES (368,'operation:social-link:create','Create social link','新增客服与社媒','Create social link','operations','API','/api/v1/admin/social-links',1,1,'2026-11-26 22:42:47.307','2026-11-26 22:42:47.307',NULL);
INSERT INTO `game_admin_permissions` VALUES (369,'operation:social-link:update','Update social link','编辑客服与社媒','Update social link','operations','API','/api/v1/admin/social-links/:id',1,1,'2026-11-26 22:42:47.307','2026-11-26 22:42:47.307',NULL);
INSERT INTO `game_admin_permissions` VALUES (370,'operation:social-link:delete','Delete social link','删除客服与社媒','Delete social link','operations','API','/api/v1/admin/social-links/:id',1,1,'2026-11-26 22:42:47.307','2026-11-26 22:42:47.307',NULL);
INSERT INTO `game_admin_permissions` VALUES (371,'operation:h5-navigation:list','List H5 navigation','查看 H5 菜单','View H5 navigation','operations','API','/api/v1/admin/h5-navigation',1,1,'2026-11-26 22:42:47.520','2026-11-26 22:42:47.520',NULL);
INSERT INTO `game_admin_permissions` VALUES (372,'operation:h5-navigation:create','Create H5 navigation','新增 H5 菜单','Create H5 navigation','operations','API','/api/v1/admin/h5-navigation',1,1,'2026-11-26 22:42:47.520','2026-11-26 22:42:47.520',NULL);
INSERT INTO `game_admin_permissions` VALUES (373,'operation:h5-navigation:update','Update H5 navigation','编辑 H5 菜单','Update H5 navigation','operations','API','/api/v1/admin/h5-navigation/:id',1,1,'2026-11-26 22:42:47.520','2026-11-26 22:42:47.520',NULL);
INSERT INTO `game_admin_permissions` VALUES (374,'operation:h5-navigation:delete','Delete H5 navigation','删除 H5 菜单','Delete H5 navigation','operations','API','/api/v1/admin/h5-navigation/:id',1,1,'2026-11-26 22:42:47.520','2026-11-26 22:42:47.520',NULL);
/*!40000 ALTER TABLE `game_admin_permissions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_role_menus` WRITE;
/*!40000 ALTER TABLE `game_admin_role_menus` DISABLE KEYS */;
INSERT INTO `game_admin_role_menus` VALUES (1,1,1,'2026-11-26 22:42:40.351','2026-11-26 22:42:47.496',NULL);
INSERT INTO `game_admin_role_menus` VALUES (2,1,2,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (3,1,3,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (4,1,4,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (5,1,5,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (6,1,6,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (7,1,7,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (8,1,8,'2026-11-26 22:42:40.351','2026-11-26 22:42:41.518','2026-11-26 22:42:41.518');
INSERT INTO `game_admin_role_menus` VALUES (9,1,9,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (10,1,10,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (11,1,11,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (12,1,12,'2026-11-26 22:42:40.351','2026-11-26 22:42:40.351',NULL);
INSERT INTO `game_admin_role_menus` VALUES (16,1,13,'2026-11-26 22:42:40.603','2026-11-26 22:42:40.603',NULL);
INSERT INTO `game_admin_role_menus` VALUES (17,1,15,'2026-11-26 22:42:41.821','2026-11-26 22:42:41.821',NULL);
INSERT INTO `game_admin_role_menus` VALUES (18,1,14,'2026-11-26 22:42:41.821','2026-11-26 22:42:41.821',NULL);
INSERT INTO `game_admin_role_menus` VALUES (22,1,17,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (23,1,18,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (24,1,19,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (25,1,20,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (26,1,21,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (27,1,22,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (28,1,23,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (29,1,24,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (30,1,25,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (31,1,26,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (32,1,27,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (33,1,28,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (34,1,29,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (35,1,30,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (36,1,31,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (37,1,32,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
INSERT INTO `game_admin_role_menus` VALUES (38,1,33,'2026-11-26 22:42:47.014','2026-11-26 22:42:47.014',NULL);
/*!40000 ALTER TABLE `game_admin_role_menus` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_role_permissions` WRITE;
/*!40000 ALTER TABLE `game_admin_role_permissions` DISABLE KEYS */;
INSERT INTO `game_admin_role_permissions` VALUES (1,1,125,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (2,1,126,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (3,1,127,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (4,1,128,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (5,1,129,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (6,1,130,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (7,1,131,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (8,1,132,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (9,1,133,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (10,1,134,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (11,1,135,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (12,1,136,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (13,1,137,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (14,1,138,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (15,1,139,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (16,1,140,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (17,1,141,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (18,1,142,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (19,1,143,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (20,1,144,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (21,1,145,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (22,1,146,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (23,1,147,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (24,1,148,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (25,1,149,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (26,1,150,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (27,1,151,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (28,1,152,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (29,1,153,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (30,1,154,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (31,1,155,'2026-11-26 22:42:40.346','2026-11-26 22:42:41.517','2026-11-26 22:42:41.517');
INSERT INTO `game_admin_role_permissions` VALUES (32,1,156,'2026-11-26 22:42:40.346','2026-11-26 22:42:41.517','2026-11-26 22:42:41.517');
INSERT INTO `game_admin_role_permissions` VALUES (33,1,157,'2026-11-26 22:42:40.346','2026-11-26 22:42:41.517','2026-11-26 22:42:41.517');
INSERT INTO `game_admin_role_permissions` VALUES (34,1,158,'2026-11-26 22:42:40.346','2026-11-26 22:42:41.517','2026-11-26 22:42:41.517');
INSERT INTO `game_admin_role_permissions` VALUES (35,1,159,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (36,1,160,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (37,1,161,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (38,1,162,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (39,1,163,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (40,1,164,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (41,1,165,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (42,1,166,'2026-11-26 22:42:40.346','2026-11-26 22:42:40.346',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (64,1,200,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (65,1,176,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (66,1,175,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (67,1,178,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (68,1,177,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (69,1,180,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (70,1,182,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (71,1,179,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (72,1,181,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (73,1,185,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (74,1,183,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (75,1,184,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (76,1,169,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (77,1,173,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (78,1,172,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (79,1,167,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (80,1,199,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (81,1,174,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (82,1,190,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (83,1,193,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (84,1,194,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (85,1,188,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (86,1,186,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (87,1,192,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (88,1,191,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (89,1,189,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (90,1,187,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (91,1,171,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (92,1,170,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (93,1,168,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (94,1,198,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (95,1,197,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (96,1,195,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (97,1,196,'2026-11-26 22:42:40.524','2026-11-26 22:42:40.524',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (127,1,207,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (128,1,211,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (129,1,206,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (130,1,202,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (131,1,204,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (132,1,208,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (133,1,205,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (134,1,210,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (135,1,209,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (136,1,212,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (137,1,203,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (138,1,201,'2026-11-26 22:42:40.564','2026-11-26 22:42:40.564',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (142,1,219,'2026-11-26 22:42:40.600','2026-11-26 22:42:40.600',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (143,1,215,'2026-11-26 22:42:40.600','2026-11-26 22:42:40.600',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (144,1,217,'2026-11-26 22:42:40.600','2026-11-26 22:42:40.600',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (145,1,213,'2026-11-26 22:42:40.600','2026-11-26 22:42:40.600',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (146,1,220,'2026-11-26 22:42:40.600','2026-11-26 22:42:40.600',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (147,1,218,'2026-11-26 22:42:40.600','2026-11-26 22:42:40.600',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (148,1,216,'2026-11-26 22:42:40.600','2026-11-26 22:42:40.600',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (149,1,214,'2026-11-26 22:42:40.600','2026-11-26 22:42:40.600',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (158,1,227,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (159,1,229,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (160,1,226,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (161,1,228,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (162,1,223,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (163,1,225,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (164,1,222,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (165,1,224,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (166,1,231,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (167,1,233,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (168,1,230,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (169,1,232,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (170,1,235,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (171,1,237,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (172,1,234,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (173,1,238,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (174,1,236,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (175,1,239,'2026-11-26 22:42:40.692','2026-11-26 22:42:40.692',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (189,1,240,'2026-11-26 22:42:40.714','2026-11-26 22:42:40.714',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (190,1,250,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (191,1,255,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (192,1,257,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (193,1,254,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (194,1,256,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (195,1,252,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (196,1,263,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (197,1,265,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (198,1,262,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (199,1,264,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (200,1,253,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (201,1,267,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (202,1,246,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (203,1,241,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (204,1,245,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (205,1,244,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (206,1,243,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (207,1,242,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (208,1,249,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (209,1,266,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (210,1,251,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (211,1,259,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (212,1,261,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (213,1,258,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (214,1,260,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (215,1,248,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (216,1,247,'2026-11-26 22:42:40.828','2026-11-26 22:42:40.828',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (221,1,272,'2026-11-26 22:42:41.142','2026-11-26 22:42:41.818',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (222,1,269,'2026-11-26 22:42:41.142','2026-11-26 22:42:41.142',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (223,1,271,'2026-11-26 22:42:41.142','2026-11-26 22:42:41.142',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (224,1,268,'2026-11-26 22:42:41.142','2026-11-26 22:42:41.142',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (225,1,270,'2026-11-26 22:42:41.142','2026-11-26 22:42:41.142',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (228,1,278,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (229,1,277,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (230,1,284,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (231,1,280,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (232,1,282,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (233,1,281,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (234,1,283,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (235,1,273,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (236,1,274,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (237,1,279,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (238,1,275,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (239,1,285,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (240,1,276,'2026-11-26 22:42:41.250','2026-11-26 22:42:41.250',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (243,1,1,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (244,1,2,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (245,1,3,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (246,1,4,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (247,1,5,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (248,1,6,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (249,1,7,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (250,1,8,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (251,1,9,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (252,1,10,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (253,1,11,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (254,1,12,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (255,1,13,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (256,1,14,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (257,1,15,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (258,1,16,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (259,1,17,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (260,1,18,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (261,1,19,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (262,1,20,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (263,1,21,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (264,1,22,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (265,1,23,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (266,1,24,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (267,1,25,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (268,1,26,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (269,1,27,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (270,1,28,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (271,1,29,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (272,1,30,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (273,1,31,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (274,1,32,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (275,1,33,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (276,1,34,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (277,1,35,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (278,1,36,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (279,1,37,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (280,1,38,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (281,1,39,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (282,1,40,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (283,1,41,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (284,1,42,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (285,1,43,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (286,1,44,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (287,1,45,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (288,1,46,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (289,1,47,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (290,1,48,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (291,1,49,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (292,1,50,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (293,1,51,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (294,1,52,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (295,1,53,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (296,1,54,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (297,1,55,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (298,1,56,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (299,1,57,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (300,1,58,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (301,1,59,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (302,1,60,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (303,1,61,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (304,1,62,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (305,1,63,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (306,1,64,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (307,1,65,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (308,1,66,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (309,1,67,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (310,1,68,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (311,1,69,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (312,1,70,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (313,1,71,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (314,1,72,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (315,1,73,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (316,1,74,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (317,1,75,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (318,1,76,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (319,1,77,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (320,1,78,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (321,1,79,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (322,1,80,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (323,1,81,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (324,1,82,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (325,1,83,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (326,1,84,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (327,1,85,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (328,1,86,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (329,1,87,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (330,1,88,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (331,1,89,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (332,1,90,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (333,1,91,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (334,1,92,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (335,1,93,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (336,1,94,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (337,1,95,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (338,1,96,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (339,1,97,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (340,1,98,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (341,1,99,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (342,1,100,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (343,1,101,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (344,1,102,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (345,1,103,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (346,1,104,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (347,1,105,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (348,1,106,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (349,1,107,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (350,1,108,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (351,1,109,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (352,1,110,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (353,1,111,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (354,1,112,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (355,1,113,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (356,1,114,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (357,1,115,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (358,1,116,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (359,1,117,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (360,1,118,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (361,1,119,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (362,1,120,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (363,1,121,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (364,1,122,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (365,1,123,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (366,1,124,'2026-11-26 22:42:41.274','2026-11-26 22:42:41.274',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (370,1,287,'2026-11-26 22:42:41.325','2026-11-26 22:42:41.325',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (371,1,286,'2026-11-26 22:42:41.325','2026-11-26 22:42:41.325',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (376,1,298,'2026-11-26 22:42:41.744','2026-11-26 22:42:41.744',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (377,1,297,'2026-11-26 22:42:41.744','2026-11-26 22:42:41.744',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (378,1,295,'2026-11-26 22:42:41.744','2026-11-26 22:42:41.744',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (388,1,302,'2026-11-26 22:42:41.817','2026-11-26 22:42:41.818',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (389,1,300,'2026-11-26 22:42:41.817','2026-11-26 22:42:41.818',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (390,1,301,'2026-11-26 22:42:41.817','2026-11-26 22:42:41.818',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (395,1,306,'2026-11-26 22:42:41.864','2026-11-26 22:42:41.864',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (396,1,304,'2026-11-26 22:42:41.864','2026-11-26 22:42:41.864',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (397,1,305,'2026-11-26 22:42:41.864','2026-11-26 22:42:41.864',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (398,1,307,'2026-11-26 22:42:42.110','2026-11-26 22:42:42.110',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (406,1,314,'2026-11-26 22:42:42.412','2026-11-26 22:42:42.412',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (407,1,316,'2026-11-26 22:42:42.412','2026-11-26 22:42:42.412',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (408,1,315,'2026-11-26 22:42:42.412','2026-11-26 22:42:42.412',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (409,1,317,'2026-11-26 22:42:45.215','2026-11-26 22:42:45.215',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (410,1,318,'2026-11-26 22:42:45.215','2026-11-26 22:42:45.215',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (412,1,321,'2026-11-26 22:42:46.494','2026-11-26 22:42:46.494',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (413,1,319,'2026-11-26 22:42:46.494','2026-11-26 22:42:46.494',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (414,1,322,'2026-11-26 22:42:46.494','2026-11-26 22:42:46.494',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (415,1,320,'2026-11-26 22:42:46.494','2026-11-26 22:42:46.494',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (419,1,325,'2026-11-26 22:42:46.575','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (420,1,324,'2026-11-26 22:42:46.575','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (421,1,323,'2026-11-26 22:42:46.575','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (422,1,328,'2026-11-26 22:42:46.575','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (423,1,327,'2026-11-26 22:42:46.575','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (424,1,326,'2026-11-26 22:42:46.575','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (433,1,335,'2026-11-26 22:42:46.860','2026-11-26 22:42:46.860',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (434,1,337,'2026-11-26 22:42:46.860','2026-11-26 22:42:46.860',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (435,1,334,'2026-11-26 22:42:46.860','2026-11-26 22:42:46.860',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (436,1,336,'2026-11-26 22:42:46.860','2026-11-26 22:42:46.860',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (440,1,357,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (441,1,358,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (442,1,359,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (443,1,360,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (444,1,361,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (445,1,362,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (446,1,363,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (447,1,353,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (448,1,354,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (449,1,355,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (450,1,356,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (451,1,347,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (452,1,348,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (453,1,349,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (454,1,350,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (455,1,351,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (456,1,352,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (457,1,343,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (458,1,344,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (459,1,345,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (460,1,346,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (461,1,364,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (462,1,365,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (463,1,366,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (464,1,338,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (465,1,339,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (466,1,340,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (467,1,341,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (468,1,342,'2026-11-26 22:42:47.010','2026-11-26 22:42:47.010',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (471,1,368,'2026-11-26 22:42:47.308','2026-11-26 22:42:47.308',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (472,1,370,'2026-11-26 22:42:47.308','2026-11-26 22:42:47.308',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (473,1,367,'2026-11-26 22:42:47.308','2026-11-26 22:42:47.308',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (474,1,369,'2026-11-26 22:42:47.308','2026-11-26 22:42:47.308',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (478,1,372,'2026-11-26 22:42:47.521','2026-11-26 22:42:47.521',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (479,1,374,'2026-11-26 22:42:47.521','2026-11-26 22:42:47.521',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (480,1,371,'2026-11-26 22:42:47.521','2026-11-26 22:42:47.521',NULL);
INSERT INTO `game_admin_role_permissions` VALUES (481,1,373,'2026-11-26 22:42:47.521','2026-11-26 22:42:47.521',NULL);
/*!40000 ALTER TABLE `game_admin_role_permissions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_roles` WRITE;
/*!40000 ALTER TABLE `game_admin_roles` DISABLE KEYS */;
INSERT INTO `game_admin_roles` VALUES (1,'super-admin','Super Administrator','超级管理员','Super Administrator','系统受保护的超级管理员角色','System-protected super administrator role','ALL',1,1,0,'2026-11-26 22:42:40.341','2026-11-26 22:42:40.341',NULL);
/*!40000 ALTER TABLE `game_admin_roles` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_admin_sessions` WRITE;
/*!40000 ALTER TABLE `game_admin_sessions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_admin_sessions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agent_channels` WRITE;
/*!40000 ALTER TABLE `game_agent_channels` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agent_channels` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agent_commission_records` WRITE;
/*!40000 ALTER TABLE `game_agent_commission_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agent_commission_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agent_commission_rules` WRITE;
/*!40000 ALTER TABLE `game_agent_commission_rules` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agent_commission_rules` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agent_hierarchy_history` WRITE;
/*!40000 ALTER TABLE `game_agent_hierarchy_history` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agent_hierarchy_history` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agent_settlement_items` WRITE;
/*!40000 ALTER TABLE `game_agent_settlement_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agent_settlement_items` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agent_settlements` WRITE;
/*!40000 ALTER TABLE `game_agent_settlements` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agent_settlements` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agent_wallet_transactions` WRITE;
/*!40000 ALTER TABLE `game_agent_wallet_transactions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agent_wallet_transactions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agent_wallets` WRITE;
/*!40000 ALTER TABLE `game_agent_wallets` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agent_wallets` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_agents` WRITE;
/*!40000 ALTER TABLE `game_agents` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_agents` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_app_download_records` WRITE;
/*!40000 ALTER TABLE `game_app_download_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_app_download_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_app_versions` WRITE;
/*!40000 ALTER TABLE `game_app_versions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_app_versions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_asset_library` WRITE;
/*!40000 ALTER TABLE `game_asset_library` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_asset_library` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_backup_tasks` WRITE;
/*!40000 ALTER TABLE `game_backup_tasks` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_backup_tasks` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_banners` WRITE;
/*!40000 ALTER TABLE `game_banners` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_banners` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_bet_records` WRITE;
/*!40000 ALTER TABLE `game_bet_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_bet_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_bet_reports` WRITE;
/*!40000 ALTER TABLE `game_bet_reports` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_bet_reports` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_bonus_records` WRITE;
/*!40000 ALTER TABLE `game_bonus_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_bonus_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_bonus_status_histories` WRITE;
/*!40000 ALTER TABLE `game_bonus_status_histories` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_bonus_status_histories` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_cash_audit_records` WRITE;
/*!40000 ALTER TABLE `game_cash_audit_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_cash_audit_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_cms_pages` WRITE;
/*!40000 ALTER TABLE `game_cms_pages` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_cms_pages` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_commission_plans` WRITE;
/*!40000 ALTER TABLE `game_commission_plans` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_commission_plans` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_commission_records` WRITE;
/*!40000 ALTER TABLE `game_commission_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_commission_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_config_releases` WRITE;
/*!40000 ALTER TABLE `game_config_releases` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_config_releases` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_content_visibility` WRITE;
/*!40000 ALTER TABLE `game_content_visibility` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_content_visibility` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_country_configs` WRITE;
/*!40000 ALTER TABLE `game_country_configs` DISABLE KEYS */;
INSERT INTO `game_country_configs` VALUES (1,0,'PK','PK','巴基斯坦','Pakistan','PKR','巴基斯坦卢比','Pakistani Rupee','₨',2,'Asia/Karachi','en','[\"en\", \"ur\"]','+92',10,'{\"logoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"appIconUrl\": \"/static/h5/images/brand-56mvp.avif\", \"headerLogoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"loadingLogoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"loadingAnimationUrl\": \"/static/lottie/loading.json\"}','{\"textColor\": \"#ffffff\", \"buttonColor\": \"#2f69d1\", \"primaryColor\": \"#3d8aef\", \"backgroundColor\": \"#050b18\"}',1,1,'2026-11-26 22:42:40.597','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (2,0,'US','US','美国','United States','USD','美元','US Dollar','$',2,'America/New_York','en','[\"en\"]','+1',10,'{\"logoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"headerLogoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"loadingLogoUrl\": \"/static/h5/images/brand-56mvp.avif\"}','{\"buttonColor\": \"#2f69d1\", \"primaryColor\": \"#3d8aef\"}',1,0,'2026-11-26 22:42:40.597','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (3,0,'MX','MX','墨西哥','Mexico','MXN','墨西哥比索','Mexican Peso','$',2,'America/Mexico_City','es-MX','[\"es-MX\", \"en\"]','+52',10,'{\"logoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"headerLogoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"loadingLogoUrl\": \"/static/h5/images/brand-56mvp.avif\"}','{\"buttonColor\": \"#2f69d1\", \"primaryColor\": \"#3d8aef\"}',1,0,'2026-11-26 22:42:40.597','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (4,0,'JP','JP','日本','Japan','JPY','日元','Japanese Yen','¥',0,'Asia/Tokyo','ja','[\"ja\", \"en\"]','+81',10,'{\"logoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"headerLogoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"loadingLogoUrl\": \"/static/h5/images/brand-56mvp.avif\"}','{\"buttonColor\": \"#2f69d1\", \"primaryColor\": \"#3d8aef\"}',1,0,'2026-11-26 22:42:40.597','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (5,0,'TH','TH','泰国','Thailand','THB','泰铢','Thai Baht','฿',2,'Asia/Bangkok','th','[\"th\", \"en\"]','+66',9,'{\"logoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"headerLogoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"loadingLogoUrl\": \"/static/h5/images/brand-56mvp.avif\"}','{\"buttonColor\": \"#2f69d1\", \"primaryColor\": \"#3d8aef\"}',1,0,'2026-11-26 22:42:40.597','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (6,0,'ID','ID','印度尼西亚','Indonesia','IDR','印度尼西亚盾','Indonesian Rupiah','Rp',0,'Asia/Jakarta','id','[\"id\", \"en\"]','+62',10,'{\"logoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"headerLogoUrl\": \"/static/h5/images/brand-56mvp.avif\", \"loadingLogoUrl\": \"/static/h5/images/brand-56mvp.avif\"}','{\"buttonColor\": \"#2f69d1\", \"primaryColor\": \"#3d8aef\"}',1,0,'2026-11-26 22:42:40.597','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (7,0,'BR','BR','巴西','Brazil','BRL','巴西雷亚尔','Brazilian Real','R$',2,'America/Sao_Paulo','pt-BR','[\"pt-BR\", \"en\"]','+55',11,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (8,0,'MY','MY','马来西亚','Malaysia','MYR','马来西亚林吉特','Malaysian Ringgit','RM',2,'Asia/Kuala_Lumpur','en','[\"en\", \"ms\", \"zh-CN\"]','+60',9,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (9,0,'PH','PH','菲律宾','Philippines','PHP','菲律宾比索','Philippine Peso','₱',2,'Asia/Manila','en','[\"en\", \"fil\"]','+63',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (10,0,'VN','VN','越南','Vietnam','VND','越南盾','Vietnamese Dong','₫',0,'Asia/Ho_Chi_Minh','vi','[\"vi\", \"en\"]','+84',9,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (11,0,'IN','IN','印度','India','INR','印度卢比','Indian Rupee','₹',2,'Asia/Kolkata','en','[\"en\", \"hi\"]','+91',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (12,0,'BD','BD','孟加拉国','Bangladesh','BDT','孟加拉塔卡','Bangladeshi Taka','৳',2,'Asia/Dhaka','bn','[\"bn\", \"en\"]','+880',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (13,0,'SG','SG','新加坡','Singapore','SGD','新加坡元','Singapore Dollar','S$',2,'Asia/Singapore','en','[\"en\", \"zh-CN\", \"ms\", \"ta\"]','+65',8,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (14,0,'NG','NG','尼日利亚','Nigeria','NGN','尼日利亚奈拉','Nigerian Naira','₦',2,'Africa/Lagos','en','[\"en\"]','+234',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (15,0,'ZA','ZA','南非','South Africa','ZAR','南非兰特','South African Rand','R',2,'Africa/Johannesburg','en','[\"en\"]','+27',9,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (16,0,'CA','CA','加拿大','Canada','CAD','加拿大元','Canadian Dollar','C$',2,'America/Toronto','en','[\"en\", \"fr\"]','+1',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (17,0,'GB','GB','英国','United Kingdom','GBP','英镑','Pound Sterling','£',2,'Europe/London','en','[\"en\"]','+44',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (18,0,'AU','AU','澳大利亚','Australia','AUD','澳大利亚元','Australian Dollar','A$',2,'Australia/Sydney','en','[\"en\"]','+61',9,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (19,0,'TR','TR','土耳其','Türkiye','TRY','土耳其里拉','Turkish Lira','₺',2,'Europe/Istanbul','tr','[\"tr\", \"en\"]','+90',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (20,0,'AR','AR','阿根廷','Argentina','ARS','阿根廷比索','Argentine Peso','$',2,'America/Argentina/Buenos_Aires','es-AR','[\"es-AR\", \"en\"]','+54',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (21,0,'CO','CO','哥伦比亚','Colombia','COP','哥伦比亚比索','Colombian Peso','$',2,'America/Bogota','es-CO','[\"es-CO\", \"en\"]','+57',10,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (22,0,'PE','PE','秘鲁','Peru','PEN','秘鲁索尔','Peruvian Sol','S/',2,'America/Lima','es-PE','[\"es-PE\", \"en\"]','+51',9,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
INSERT INTO `game_country_configs` VALUES (23,0,'AE','AE','阿联酋','United Arab Emirates','AED','阿联酋迪拉姆','UAE Dirham','د.إ',2,'Asia/Dubai','ar','[\"ar\", \"en\"]','+971',9,'{}','{}',1,0,'2026-11-26 22:42:47.035','2026-11-26 22:42:47.035',NULL);
/*!40000 ALTER TABLE `game_country_configs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_daily_agent_summaries` WRITE;
/*!40000 ALTER TABLE `game_daily_agent_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_daily_agent_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_daily_channel_summaries` WRITE;
/*!40000 ALTER TABLE `game_daily_channel_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_daily_channel_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_daily_financial_summaries` WRITE;
/*!40000 ALTER TABLE `game_daily_financial_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_daily_financial_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_daily_game_summaries` WRITE;
/*!40000 ALTER TABLE `game_daily_game_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_daily_game_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_daily_promotion_summaries` WRITE;
/*!40000 ALTER TABLE `game_daily_promotion_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_daily_promotion_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_daily_provider_summaries` WRITE;
/*!40000 ALTER TABLE `game_daily_provider_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_daily_provider_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_daily_risk_summaries` WRITE;
/*!40000 ALTER TABLE `game_daily_risk_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_daily_risk_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_daily_user_summaries` WRITE;
/*!40000 ALTER TABLE `game_daily_user_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_daily_user_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_deposit_amount_options` WRITE;
/*!40000 ALTER TABLE `game_deposit_amount_options` DISABLE KEYS */;
INSERT INTO `game_deposit_amount_options` VALUES (1,0,'deposit-100','PKR',100.00,30.00,20.00,1,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (2,0,'deposit-300','PKR',300.00,100.00,40.00,2,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (3,0,'deposit-500','PKR',500.00,170.00,80.00,3,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (4,0,'deposit-1000','PKR',1000.00,300.00,200.00,4,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (5,0,'deposit-3000','PKR',3000.00,1000.00,500.00,5,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (6,0,'deposit-5000','PKR',5000.00,1700.00,800.00,6,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (7,0,'deposit-10000','PKR',10000.00,3500.00,1500.00,7,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (8,0,'deposit-30000','PKR',30000.00,10000.00,5000.00,8,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (9,0,'deposit-50000','PKR',50000.00,17000.00,8000.00,9,1,'2026-11-26 22:42:47.985','2026-11-26 22:42:47.985',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (16,1,'deposit-100','PKR',100.00,30.00,20.00,1,1,'2026-11-26 22:42:48.579','2026-11-26 22:42:48.579',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (17,1,'deposit-300','PKR',300.00,100.00,40.00,2,1,'2026-11-26 22:42:48.579','2026-11-26 22:42:48.579',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (18,1,'deposit-500','PKR',500.00,170.00,80.00,3,1,'2026-11-26 22:42:48.580','2026-11-26 22:42:48.580',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (19,1,'deposit-1000','PKR',1000.00,300.00,200.00,4,1,'2026-11-26 22:42:48.580','2026-11-26 22:42:48.580',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (20,1,'deposit-3000','PKR',3000.00,1000.00,500.00,5,1,'2026-11-26 22:42:48.580','2026-11-26 22:42:48.580',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (21,1,'deposit-5000','PKR',5000.00,1700.00,800.00,6,1,'2026-11-26 22:42:48.580','2026-11-26 22:42:48.580',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (22,1,'deposit-10000','PKR',10000.00,3500.00,1500.00,7,1,'2026-11-26 22:42:48.580','2026-11-26 22:42:48.580',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (23,1,'deposit-30000','PKR',30000.00,10000.00,5000.00,8,1,'2026-11-26 22:42:48.580','2026-11-26 22:42:48.580',NULL);
INSERT INTO `game_deposit_amount_options` VALUES (24,1,'deposit-50000','PKR',50000.00,17000.00,8000.00,9,1,'2026-11-26 22:42:48.580','2026-11-26 22:42:48.580',NULL);
/*!40000 ALTER TABLE `game_deposit_amount_options` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_deposit_callback_logs` WRITE;
/*!40000 ALTER TABLE `game_deposit_callback_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_deposit_callback_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_deposit_orders` WRITE;
/*!40000 ALTER TABLE `game_deposit_orders` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_deposit_orders` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_deposit_voucher_rewards` WRITE;
/*!40000 ALTER TABLE `game_deposit_voucher_rewards` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_deposit_voucher_rewards` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_deposit_vouchers` WRITE;
/*!40000 ALTER TABLE `game_deposit_vouchers` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_deposit_vouchers` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_categories` WRITE;
/*!40000 ALTER TABLE `game_game_categories` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_game_categories` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_category_relations` WRITE;
/*!40000 ALTER TABLE `game_game_category_relations` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_game_category_relations` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_collection_relations` WRITE;
/*!40000 ALTER TABLE `game_game_collection_relations` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_game_collection_relations` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_collections` WRITE;
/*!40000 ALTER TABLE `game_game_collections` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_game_collections` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_launch_logs` WRITE;
/*!40000 ALTER TABLE `game_game_launch_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_game_launch_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_maintenance` WRITE;
/*!40000 ALTER TABLE `game_game_maintenance` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_game_maintenance` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_providers` WRITE;
/*!40000 ALTER TABLE `game_game_providers` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_game_providers` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_tag_relations` WRITE;
/*!40000 ALTER TABLE `game_game_tag_relations` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_game_tag_relations` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_game_tags` WRITE;
/*!40000 ALTER TABLE `game_game_tags` DISABLE KEYS */;
INSERT INTO `game_game_tags` VALUES (1,0,'hot','热门','Hot','',10,1,'2026-11-26 22:42:40.826','2026-11-26 22:42:40.826',NULL);
INSERT INTO `game_game_tags` VALUES (2,0,'new','新品','New','',20,1,'2026-11-26 22:42:40.826','2026-11-26 22:42:40.826',NULL);
INSERT INTO `game_game_tags` VALUES (3,0,'featured','推荐','Featured','',30,1,'2026-11-26 22:42:40.826','2026-11-26 22:42:40.826',NULL);
INSERT INTO `game_game_tags` VALUES (4,0,'exclusive','独家','Exclusive','',40,1,'2026-11-26 22:42:40.826','2026-11-26 22:42:40.826',NULL);
INSERT INTO `game_game_tags` VALUES (5,0,'high-rtp','高返奖率','High RTP','',50,1,'2026-11-26 22:42:40.826','2026-11-26 22:42:40.826',NULL);
INSERT INTO `game_game_tags` VALUES (6,0,'mexico-popular','墨西哥热门','Mexico Popular','',60,1,'2026-11-26 22:42:40.826','2026-11-26 22:42:40.826',NULL);
/*!40000 ALTER TABLE `game_game_tags` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_games` WRITE;
/*!40000 ALTER TABLE `game_games` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_games` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_h5_navigation_items` WRITE;
/*!40000 ALTER TABLE `game_h5_navigation_items` DISABLE KEYS */;
INSERT INTO `game_h5_navigation_items` VALUES (1,0,0,'home','首页','Home',NULL,0,0,0,0,0,0,'BOTTOM_NAV','OPEN_INTERNAL_PAGE','/','/','',0,'','',10,'NONE','ALL','[]','[]',NULL,NULL,1,NULL,NULL,'PUBLISHED',1,0,0,NULL,'2026-11-26 22:42:41.495',0,'2026-11-26 22:42:41.495','2026-11-26 22:42:41.495',NULL);
INSERT INTO `game_h5_navigation_items` VALUES (2,0,0,'offers','优惠','Offers',NULL,0,0,0,0,0,0,'BOTTOM_NAV','OPEN_INTERNAL_PAGE','/offers','/offers','',0,'','',20,'NONE','ALL','[]','[]',NULL,NULL,1,NULL,NULL,'PUBLISHED',1,0,0,NULL,'2026-11-26 22:42:41.495',0,'2026-11-26 22:42:41.495','2026-11-26 22:42:41.495',NULL);
INSERT INTO `game_h5_navigation_items` VALUES (3,0,0,'share','分享','Share',NULL,0,0,0,0,0,0,'BOTTOM_NAV','OPEN_INTERNAL_PAGE','/share','/share','',0,'','',40,'NONE','ALL','[]','[]',NULL,NULL,1,NULL,NULL,'PUBLISHED',1,0,0,NULL,'2026-11-26 22:42:41.495',0,'2026-11-26 22:42:41.495','2026-11-26 22:42:41.495',NULL);
INSERT INTO `game_h5_navigation_items` VALUES (4,0,0,'profile','我的','Profile',NULL,0,0,0,0,0,0,'BOTTOM_NAV','OPEN_INTERNAL_PAGE','/profile','/profile','',0,'','',50,'NONE','ALL','[]','[]',NULL,NULL,1,NULL,NULL,'PUBLISHED',1,0,0,NULL,'2026-11-26 22:42:41.495',0,'2026-11-26 22:42:41.495','2026-11-26 22:42:41.495',NULL);
INSERT INTO `game_h5_navigation_items` VALUES (5,1,0,'home','首页','Home',NULL,0,0,0,0,0,0,'BOTTOM_NAV','OPEN_INTERNAL_PAGE','/','/','',0,'','',10,'NONE','ALL','[]','[]',NULL,NULL,1,NULL,NULL,'PUBLISHED',1,0,0,NULL,'2026-11-26 22:42:48.577',0,'2026-11-26 22:42:48.577','2026-11-26 22:42:48.577',NULL);
INSERT INTO `game_h5_navigation_items` VALUES (6,1,0,'offers','优惠','Offers',NULL,0,0,0,0,0,0,'BOTTOM_NAV','OPEN_INTERNAL_PAGE','/offers','/offers','',0,'','',20,'NONE','ALL','[]','[]',NULL,NULL,1,NULL,NULL,'PUBLISHED',1,0,0,NULL,'2026-11-26 22:42:48.577',0,'2026-11-26 22:42:48.577','2026-11-26 22:42:48.577',NULL);
INSERT INTO `game_h5_navigation_items` VALUES (7,1,0,'profile','我的','Profile',NULL,0,0,0,0,0,0,'BOTTOM_NAV','OPEN_INTERNAL_PAGE','/profile','/profile','',0,'','',50,'NONE','ALL','[]','[]',NULL,NULL,1,NULL,NULL,'PUBLISHED',1,0,0,NULL,'2026-11-26 22:42:48.577',0,'2026-11-26 22:42:48.577','2026-11-26 22:42:48.577',NULL);
INSERT INTO `game_h5_navigation_items` VALUES (8,1,0,'share','分享','Share',NULL,0,0,0,0,0,0,'BOTTOM_NAV','OPEN_INTERNAL_PAGE','/share','/share','',0,'','',40,'NONE','ALL','[]','[]',NULL,NULL,1,NULL,NULL,'PUBLISHED',1,0,0,NULL,'2026-11-26 22:42:48.577',0,'2026-11-26 22:42:48.577','2026-11-26 22:42:48.577',NULL);
/*!40000 ALTER TABLE `game_h5_navigation_items` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_home_feature_cards` WRITE;
/*!40000 ALTER TABLE `game_home_feature_cards` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_home_feature_cards` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_hourly_platform_summaries` WRITE;
/*!40000 ALTER TABLE `game_hourly_platform_summaries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_hourly_platform_summaries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_invitation_chest_configs` WRITE;
/*!40000 ALTER TABLE `game_invitation_chest_configs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_invitation_chest_configs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_invitation_wheel_prizes` WRITE;
/*!40000 ALTER TABLE `game_invitation_wheel_prizes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_invitation_wheel_prizes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_invitation_wheel_rounds` WRITE;
/*!40000 ALTER TABLE `game_invitation_wheel_rounds` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_invitation_wheel_rounds` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_invitation_wheel_spins` WRITE;
/*!40000 ALTER TABLE `game_invitation_wheel_spins` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_invitation_wheel_spins` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_invites` WRITE;
/*!40000 ALTER TABLE `game_invites` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_invites` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_lobby_sections` WRITE;
/*!40000 ALTER TABLE `game_lobby_sections` DISABLE KEYS */;
INSERT INTO `game_lobby_sections` VALUES (1,0,'top-banner','banner','顶部 Banner','Top Banner','carousel','banner','home',10,10,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Top Banner\", \"titleZh\": \"顶部 Banner\", \"maxItems\": 10, \"sortOrder\": 10, \"sourceRef\": \"home\", \"sourceType\": \"banner\", \"sectionCode\": \"top-banner\", \"sectionType\": \"banner\", \"displayStyle\": \"carousel\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (2,0,'jackpot','jackpot','奖池','Jackpot','ticker','jackpot','',1,20,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Jackpot\", \"titleZh\": \"奖池\", \"maxItems\": 1, \"sortOrder\": 20, \"sourceRef\": \"\", \"sourceType\": \"jackpot\", \"sectionCode\": \"jackpot\", \"sectionType\": \"jackpot\", \"displayStyle\": \"ticker\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (3,0,'hot-games','game-list','热门游戏','Hot Games','horizontal','flag','hot',12,30,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Hot Games\", \"titleZh\": \"热门游戏\", \"maxItems\": 12, \"sortOrder\": 30, \"sourceRef\": \"hot\", \"sourceType\": \"flag\", \"sectionCode\": \"hot-games\", \"sectionType\": \"game-list\", \"displayStyle\": \"horizontal\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (4,0,'new-games','game-list','新游戏','New Games','horizontal','flag','new',12,40,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"New Games\", \"titleZh\": \"新游戏\", \"maxItems\": 12, \"sortOrder\": 40, \"sourceRef\": \"new\", \"sourceType\": \"flag\", \"sectionCode\": \"new-games\", \"sectionType\": \"game-list\", \"displayStyle\": \"horizontal\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (5,0,'recommended-games','game-list','推荐游戏','Recommended Games','horizontal','flag','featured',12,50,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Recommended Games\", \"titleZh\": \"推荐游戏\", \"maxItems\": 12, \"sortOrder\": 50, \"sourceRef\": \"featured\", \"sourceType\": \"flag\", \"sectionCode\": \"recommended-games\", \"sectionType\": \"game-list\", \"displayStyle\": \"horizontal\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (6,0,'continue-playing','game-list','继续游戏','Continue Playing','horizontal','user','continue-playing',12,60,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Continue Playing\", \"titleZh\": \"继续游戏\", \"maxItems\": 12, \"sortOrder\": 60, \"sourceRef\": \"continue-playing\", \"sourceType\": \"user\", \"sectionCode\": \"continue-playing\", \"sectionType\": \"game-list\", \"displayStyle\": \"horizontal\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (7,0,'recently-played','game-list','最近玩过','Recently Played','horizontal','user','recently-played',12,70,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Recently Played\", \"titleZh\": \"最近玩过\", \"maxItems\": 12, \"sortOrder\": 70, \"sourceRef\": \"recently-played\", \"sourceType\": \"user\", \"sectionCode\": \"recently-played\", \"sectionType\": \"game-list\", \"displayStyle\": \"horizontal\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (8,0,'categories','category-list','分类','Categories','grid','category','',20,80,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Categories\", \"titleZh\": \"分类\", \"maxItems\": 20, \"sortOrder\": 80, \"sourceRef\": \"\", \"sourceType\": \"category\", \"sectionCode\": \"categories\", \"sectionType\": \"category-list\", \"displayStyle\": \"grid\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (9,0,'providers','provider-list','平台','Providers','grid','provider','',20,90,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Providers\", \"titleZh\": \"平台\", \"maxItems\": 20, \"sortOrder\": 90, \"sourceRef\": \"\", \"sourceType\": \"provider\", \"sectionCode\": \"providers\", \"sectionType\": \"provider-list\", \"displayStyle\": \"grid\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (10,0,'footer','footer','底部','Footer','footer','config','',1,100,1,'','','{}','{\"config\": {}, \"enabled\": 1, \"titleEn\": \"Footer\", \"titleZh\": \"底部\", \"maxItems\": 1, \"sortOrder\": 100, \"sourceRef\": \"\", \"sourceType\": \"config\", \"sectionCode\": \"footer\", \"sectionType\": \"footer\", \"displayStyle\": \"footer\", \"supportedCountries\": \"\", \"supportedCurrencies\": \"\"}','PUBLISHED',1,'2026-11-26 22:42:42.410',0,NULL,0,'2026-11-26 22:42:42.410','2026-11-26 22:42:45.212',NULL);
INSERT INTO `game_lobby_sections` VALUES (11,1,'top-banner','banner','顶部 Banner','Top Banner','carousel','banner','home',10,10,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.576','2026-11-26 22:42:48.576',NULL);
INSERT INTO `game_lobby_sections` VALUES (12,1,'jackpot','jackpot','奖池','Jackpot','ticker','jackpot','',1,20,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.576','2026-11-26 22:42:48.576',NULL);
INSERT INTO `game_lobby_sections` VALUES (13,1,'hot-games','game-list','热门游戏','Hot Games','horizontal','flag','hot',12,30,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.576','2026-11-26 22:42:48.576',NULL);
INSERT INTO `game_lobby_sections` VALUES (14,1,'new-games','game-list','新游戏','New Games','horizontal','flag','new',12,40,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.576','2026-11-26 22:42:48.576',NULL);
INSERT INTO `game_lobby_sections` VALUES (15,1,'recommended-games','game-list','推荐游戏','Recommended Games','horizontal','flag','featured',12,50,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.576','2026-11-26 22:42:48.576',NULL);
INSERT INTO `game_lobby_sections` VALUES (16,1,'continue-playing','game-list','继续游戏','Continue Playing','horizontal','user','continue-playing',12,60,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.576','2026-11-26 22:42:48.576',NULL);
INSERT INTO `game_lobby_sections` VALUES (17,1,'recently-played','game-list','最近玩过','Recently Played','horizontal','user','recently-played',12,70,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.576','2026-11-26 22:42:48.576',NULL);
INSERT INTO `game_lobby_sections` VALUES (18,1,'categories','category-list','分类','Categories','grid','category','',20,80,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.577','2026-11-26 22:42:48.577',NULL);
INSERT INTO `game_lobby_sections` VALUES (19,1,'providers','provider-list','平台','Providers','grid','provider','',20,90,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.577','2026-11-26 22:42:48.577',NULL);
INSERT INTO `game_lobby_sections` VALUES (20,1,'footer','footer','底部','Footer','footer','config','',1,100,1,'','','{}',NULL,'PUBLISHED',1,NULL,0,NULL,0,'2026-11-26 22:42:48.577','2026-11-26 22:42:48.577',NULL);
/*!40000 ALTER TABLE `game_lobby_sections` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_manual_adjustment_approvals` WRITE;
/*!40000 ALTER TABLE `game_manual_adjustment_approvals` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_manual_adjustment_approvals` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_manual_adjustments` WRITE;
/*!40000 ALTER TABLE `game_manual_adjustments` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_manual_adjustments` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_notices` WRITE;
/*!40000 ALTER TABLE `game_notices` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_notices` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_notification_deliveries` WRITE;
/*!40000 ALTER TABLE `game_notification_deliveries` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_notification_deliveries` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_notification_jobs` WRITE;
/*!40000 ALTER TABLE `game_notification_jobs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_notification_jobs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_notifications` WRITE;
/*!40000 ALTER TABLE `game_notifications` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_notifications` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_operation_music` WRITE;
/*!40000 ALTER TABLE `game_operation_music` DISABLE KEYS */;
INSERT INTO `game_operation_music` VALUES (1,0,'2026-11-26 22:42:41.140','2026-11-26 22:42:41.140',NULL,'lobby-bgm','大厅背景音乐','Lobby Background Music','/static/h5/audio/backgound_music.mp3','LOBBY',1,70,10,1,NULL,NULL);
INSERT INTO `game_operation_music` VALUES (2,1,'2026-11-26 22:42:48.587','2026-11-26 22:42:48.587',NULL,'lobby-bgm','大厅背景音乐','Lobby Background Music','/static/h5/audio/backgound_music.mp3','LOBBY',1,70,10,1,NULL,NULL);
/*!40000 ALTER TABLE `game_operation_music` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_operational_notes` WRITE;
/*!40000 ALTER TABLE `game_operational_notes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_operational_notes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_payment_channels` WRITE;
/*!40000 ALTER TABLE `game_payment_channels` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_payment_channels` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_payment_methods` WRITE;
/*!40000 ALTER TABLE `game_payment_methods` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_payment_methods` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_popup_campaign_events` WRITE;
/*!40000 ALTER TABLE `game_popup_campaign_events` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_popup_campaign_events` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_popup_campaigns` WRITE;
/*!40000 ALTER TABLE `game_popup_campaigns` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_popup_campaigns` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_popup_components` WRITE;
/*!40000 ALTER TABLE `game_popup_components` DISABLE KEYS */;
INSERT INTO `game_popup_components` VALUES (1,'DEPOSIT_POPUP','充值活动弹窗','Deposit Popup','金币与首充引导动画','Animated deposit and first-deposit prompt','DEPOSIT','','','1.0.0','OPEN_DEPOSIT','[\"OPEN_DEPOSIT\", \"OPEN_PROMOTION\", \"CLOSE_ONLY\"]',1,10,0,0,'2026-11-26 22:42:46.858','2026-11-26 22:42:46.858',NULL);
INSERT INTO `game_popup_components` VALUES (2,'WELCOME_BONUS_POPUP','欢迎奖励弹窗','Welcome Bonus Popup','欢迎礼物与奖励动画','Welcome gift and reward animation','PROMOTION','','','1.0.0','OPEN_PROMOTION','[\"OPEN_PROMOTION\", \"OPEN_COUPON\", \"CLOSE_ONLY\"]',1,20,0,0,'2026-11-26 22:42:46.858','2026-11-26 22:42:46.858',NULL);
INSERT INTO `game_popup_components` VALUES (3,'CASHBACK_POPUP','返水活动弹窗','Cashback Popup','返水金币与奖励动画','Cashback coin and reward animation','PROMOTION','','','1.0.0','OPEN_PROMOTION','[\"OPEN_PROMOTION\", \"OPEN_INTERNAL_PAGE\", \"CLOSE_ONLY\"]',1,30,0,0,'2026-11-26 22:42:46.858','2026-11-26 22:42:46.858',NULL);
INSERT INTO `game_popup_components` VALUES (4,'VIP_UPGRADE_POPUP','VIP 升级弹窗','VIP Upgrade Popup','VIP 徽章升级动画','Animated VIP upgrade badge','VIP','','','1.0.0','OPEN_VIP','[\"OPEN_VIP\", \"OPEN_PROMOTION\", \"CLOSE_ONLY\"]',1,40,0,0,'2026-11-26 22:42:46.858','2026-11-26 22:42:46.858',NULL);
INSERT INTO `game_popup_components` VALUES (5,'APP_DOWNLOAD_POPUP','App 下载弹窗','App Download Popup','App 下载与安装引导动画','Animated app install prompt','DOWNLOAD','','','1.0.0','OPEN_DOWNLOAD','[\"OPEN_DOWNLOAD\", \"OPEN_EXTERNAL_URL\", \"CLOSE_ONLY\"]',1,50,0,0,'2026-11-26 22:42:46.858','2026-11-26 22:42:46.858',NULL);
/*!40000 ALTER TABLE `game_popup_components` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_promo_code_redemptions` WRITE;
/*!40000 ALTER TABLE `game_promo_code_redemptions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_promo_code_redemptions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_promo_codes` WRITE;
/*!40000 ALTER TABLE `game_promo_codes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_promo_codes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_promotion_claim_rewards` WRITE;
/*!40000 ALTER TABLE `game_promotion_claim_rewards` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_promotion_claim_rewards` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_promotion_claims` WRITE;
/*!40000 ALTER TABLE `game_promotion_claims` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_promotion_claims` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_promotion_eligible_games` WRITE;
/*!40000 ALTER TABLE `game_promotion_eligible_games` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_promotion_eligible_games` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_promotion_participation_rules` WRITE;
/*!40000 ALTER TABLE `game_promotion_participation_rules` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_promotion_participation_rules` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_promotion_reward_rules` WRITE;
/*!40000 ALTER TABLE `game_promotion_reward_rules` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_promotion_reward_rules` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_provider_callback_logs` WRITE;
/*!40000 ALTER TABLE `game_provider_callback_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_provider_callback_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_provider_category_mapping` WRITE;
/*!40000 ALTER TABLE `game_provider_category_mapping` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_provider_category_mapping` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_provider_maintenance` WRITE;
/*!40000 ALTER TABLE `game_provider_maintenance` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_provider_maintenance` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_provider_players` WRITE;
/*!40000 ALTER TABLE `game_provider_players` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_provider_players` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_provider_sync_log_items` WRITE;
/*!40000 ALTER TABLE `game_provider_sync_log_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_provider_sync_log_items` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_provider_sync_logs` WRITE;
/*!40000 ALTER TABLE `game_provider_sync_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_provider_sync_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_provider_syncs` WRITE;
/*!40000 ALTER TABLE `game_provider_syncs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_provider_syncs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_provider_transactions` WRITE;
/*!40000 ALTER TABLE `game_provider_transactions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_provider_transactions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_push_device_tokens` WRITE;
/*!40000 ALTER TABLE `game_push_device_tokens` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_push_device_tokens` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_ranking_periods` WRITE;
/*!40000 ALTER TABLE `game_ranking_periods` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_ranking_periods` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_ranking_records` WRITE;
/*!40000 ALTER TABLE `game_ranking_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_ranking_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_rebate_jackpots` WRITE;
/*!40000 ALTER TABLE `game_rebate_jackpots` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_rebate_jackpots` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_reconciliation_differences` WRITE;
/*!40000 ALTER TABLE `game_reconciliation_differences` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_reconciliation_differences` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_reconciliation_logs` WRITE;
/*!40000 ALTER TABLE `game_reconciliation_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_reconciliation_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_reconciliation_records` WRITE;
/*!40000 ALTER TABLE `game_reconciliation_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_reconciliation_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_reconciliation_resolutions` WRITE;
/*!40000 ALTER TABLE `game_reconciliation_resolutions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_reconciliation_resolutions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_red_packet_claims` WRITE;
/*!40000 ALTER TABLE `game_red_packet_claims` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_red_packet_claims` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_red_packet_rounds` WRITE;
/*!40000 ALTER TABLE `game_red_packet_rounds` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_red_packet_rounds` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_register_reward_claims` WRITE;
/*!40000 ALTER TABLE `game_register_reward_claims` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_register_reward_claims` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_report_export_tasks` WRITE;
/*!40000 ALTER TABLE `game_report_export_tasks` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_report_export_tasks` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_blacklist_histories` WRITE;
/*!40000 ALTER TABLE `game_risk_blacklist_histories` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_blacklist_histories` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_blacklist_user_relations` WRITE;
/*!40000 ALTER TABLE `game_risk_blacklist_user_relations` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_blacklist_user_relations` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_blacklists` WRITE;
/*!40000 ALTER TABLE `game_risk_blacklists` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_blacklists` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_event_evidence` WRITE;
/*!40000 ALTER TABLE `game_risk_event_evidence` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_event_evidence` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_internal_notes` WRITE;
/*!40000 ALTER TABLE `game_risk_internal_notes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_internal_notes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_review_histories` WRITE;
/*!40000 ALTER TABLE `game_risk_review_histories` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_review_histories` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_reviews` WRITE;
/*!40000 ALTER TABLE `game_risk_reviews` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_reviews` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_rule_actions` WRITE;
/*!40000 ALTER TABLE `game_risk_rule_actions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_rule_actions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_rule_conditions` WRITE;
/*!40000 ALTER TABLE `game_risk_rule_conditions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_rule_conditions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_rule_test_runs` WRITE;
/*!40000 ALTER TABLE `game_risk_rule_test_runs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_rule_test_runs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_risk_rules` WRITE;
/*!40000 ALTER TABLE `game_risk_rules` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_risk_rules` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_rollbacks` WRITE;
/*!40000 ALTER TABLE `game_rollbacks` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_rollbacks` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_rounds` WRITE;
/*!40000 ALTER TABLE `game_rounds` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_rounds` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_site_announcements` WRITE;
/*!40000 ALTER TABLE `game_site_announcements` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_site_announcements` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_site_banners` WRITE;
/*!40000 ALTER TABLE `game_site_banners` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_site_banners` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_site_config` WRITE;
/*!40000 ALTER TABLE `game_site_config` DISABLE KEYS */;
INSERT INTO `game_site_config` (`id`, `site_code`, `site_name`, `primary_domain`, `admin_domain`, `h5_domain`, `android_package_name`, `android_apk_object_key`, `android_apk_name`, `android_apk_size`, `android_apk_uploaded_at`, `ios_bundle_id`, `country_code`, `currency_code`, `default_language`, `timezone`, `brand_config`, `theme_config`, `feature_config`, `maintenance_config`, `seo_config`, `site_name_draft`, `country_code_draft`, `currency_code_draft`, `default_language_draft`, `timezone_draft`, `brand_config_draft`, `theme_config_draft`, `feature_config_draft`, `maintenance_config_draft`, `seo_config_draft`, `config_status`, `config_version`, `config_published_at`, `config_published_by`, `config_draft_updated_at`, `config_draft_updated_by`, `is_initialized`, `enabled`, `status`, `remark`, `created_by`, `updated_by`, `version`, `created_at`, `updated_at`, `deleted_at`) VALUES (1,'main','Casino','','','','','','',0,NULL,'','PK','PKR','en','Asia/Karachi','{\"name\": \"Casino\", \"appName\": \"Casino\"}','{\"text\": \"#ffffff\", \"button\": \"#2f69d1\", \"primary\": \"#3d8aef\", \"background\": \"#050b18\"}','{}','{\"enabled\": false}','{\"title\": \"Casino\"}','Casino','PK','PKR','en','Asia/Karachi','{\"name\": \"Casino\", \"appName\": \"Casino\"}','{\"text\": \"#ffffff\", \"button\": \"#2f69d1\", \"primary\": \"#3d8aef\", \"background\": \"#050b18\"}','{}','{\"enabled\": false}','{\"title\": \"Casino\"}','PUBLISHED',1,'2026-11-26 22:42:48.574',1,NULL,0,1,1,'ONLINE','',1,1,0,'2026-11-26 22:42:48.574','2026-11-26 22:42:48.574',NULL);
/*!40000 ALTER TABLE `game_site_config` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_site_maintenance` WRITE;
/*!40000 ALTER TABLE `game_site_maintenance` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_site_maintenance` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_sms_verification_codes` WRITE;
/*!40000 ALTER TABLE `game_sms_verification_codes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_sms_verification_codes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_social_links` WRITE;
/*!40000 ALTER TABLE `game_social_links` DISABLE KEYS */;
INSERT INTO `game_social_links` VALUES (1,0,'facebook','Facebook','/static/h5/images/social-facebook.webp','https://www.facebook.com/sharer/sharer.php?u={url}&quote={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',1,1,'2026-11-26 22:42:47.983','2026-11-26 22:42:47.983',NULL);
INSERT INTO `game_social_links` VALUES (2,0,'messenger','Messenger','/static/h5/images/social-messenger.webp','fb-messenger://share?link={url}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',2,1,'2026-11-26 22:42:47.983','2026-11-26 22:42:47.983',NULL);
INSERT INTO `game_social_links` VALUES (3,0,'whatsapp','WhatsApp','/static/h5/images/social-whatsapp.webp','https://wa.me/?text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',3,1,'2026-11-26 22:42:47.983','2026-11-26 22:42:47.983',NULL);
INSERT INTO `game_social_links` VALUES (4,0,'telegram','Telegram','/static/h5/images/social-telegram.webp','https://t.me/share/url?url={url}&text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',4,1,'2026-11-26 22:42:47.983','2026-11-26 22:42:47.983',NULL);
INSERT INTO `game_social_links` VALUES (5,0,'instagram','Instagram','/static/h5/images/social-instagram.webp','https://www.instagram.com/','Join 56MVP','Copy my invite link and share it on Instagram: {url}','Invite friends, unlock rewards, and enjoy games together.',5,1,'2026-11-26 22:42:47.983','2026-11-26 22:42:47.983',NULL);
INSERT INTO `game_social_links` VALUES (6,0,'tiktok','TikTok','/static/h5/images/social-tiktok.webp','https://www.tiktok.com/','Join 56MVP','Copy my invite link and share it on TikTok: {url}','Invite friends, unlock rewards, and enjoy games together.',6,1,'2026-11-26 22:42:47.983','2026-11-26 22:42:47.983',NULL);
INSERT INTO `game_social_links` VALUES (7,0,'x','X','/static/h5/images/social-x.webp','https://twitter.com/intent/tweet?url={url}&text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',7,1,'2026-11-26 22:42:47.983','2026-11-26 22:42:47.983',NULL);
INSERT INTO `game_social_links` VALUES (8,1,'facebook','Facebook','/static/h5/images/social-facebook.webp','https://www.facebook.com/sharer/sharer.php?u={url}&quote={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',1,1,'2026-11-26 22:42:48.583','2026-11-26 22:42:48.583',NULL);
INSERT INTO `game_social_links` VALUES (9,1,'messenger','Messenger','/static/h5/images/social-messenger.webp','fb-messenger://share?link={url}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',2,1,'2026-11-26 22:42:48.584','2026-11-26 22:42:48.584',NULL);
INSERT INTO `game_social_links` VALUES (10,1,'whatsapp','WhatsApp','/static/h5/images/social-whatsapp.webp','https://wa.me/?text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',3,1,'2026-11-26 22:42:48.584','2026-11-26 22:42:48.584',NULL);
INSERT INTO `game_social_links` VALUES (11,1,'telegram','Telegram','/static/h5/images/social-telegram.webp','https://t.me/share/url?url={url}&text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',4,1,'2026-11-26 22:42:48.584','2026-11-26 22:42:48.584',NULL);
INSERT INTO `game_social_links` VALUES (12,1,'instagram','Instagram','/static/h5/images/social-instagram.webp','https://www.instagram.com/','Join 56MVP','Copy my invite link and share it on Instagram: {url}','Invite friends, unlock rewards, and enjoy games together.',5,1,'2026-11-26 22:42:48.585','2026-11-26 22:42:48.585',NULL);
INSERT INTO `game_social_links` VALUES (13,1,'tiktok','TikTok','/static/h5/images/social-tiktok.webp','https://www.tiktok.com/','Join 56MVP','Copy my invite link and share it on TikTok: {url}','Invite friends, unlock rewards, and enjoy games together.',6,1,'2026-11-26 22:42:48.585','2026-11-26 22:42:48.585',NULL);
INSERT INTO `game_social_links` VALUES (14,1,'x','X','/static/h5/images/social-x.webp','https://twitter.com/intent/tweet?url={url}&text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',7,1,'2026-11-26 22:42:48.586','2026-11-26 22:42:48.586',NULL);
/*!40000 ALTER TABLE `game_social_links` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_support_case_notes` WRITE;
/*!40000 ALTER TABLE `game_support_case_notes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_support_case_notes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_support_cases` WRITE;
/*!40000 ALTER TABLE `game_support_cases` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_support_cases` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_support_categories` WRITE;
/*!40000 ALTER TABLE `game_support_categories` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_support_categories` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_support_export_tasks` WRITE;
/*!40000 ALTER TABLE `game_support_export_tasks` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_support_export_tasks` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_support_messages` WRITE;
/*!40000 ALTER TABLE `game_support_messages` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_support_messages` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_support_questions` WRITE;
/*!40000 ALTER TABLE `game_support_questions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_support_questions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_support_tickets` WRITE;
/*!40000 ALTER TABLE `game_support_tickets` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_support_tickets` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_system_alerts` WRITE;
/*!40000 ALTER TABLE `game_system_alerts` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_system_alerts` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_system_configs` WRITE;
/*!40000 ALTER TABLE `game_system_configs` DISABLE KEYS */;
INSERT INTO `game_system_configs` VALUES (1,0,'basic.platform_name','平台名称','Platform Name','PLATFORM','Casino Admin','STRING','Platform display name','平台与站点后台左上角名称','Name displayed in platform and site admin headers',1,0,0,NULL,1,'2026-11-26 22:42:40.333','2026-11-26 23:06:42.263',NULL);
INSERT INTO `game_system_configs` VALUES (2,0,'security.session_timeout_minutes','会话超时分钟数','Session Timeout Minutes','SECURITY','720','NUMBER','Admin session timeout','管理员会话超时时间','Administrator session timeout',1,0,0,NULL,1,'2026-11-26 22:42:40.333','2026-11-26 22:42:40.333',NULL);
INSERT INTO `game_system_configs` VALUES (3,0,'security.password_min_length','密码最小长度','Minimum Password Length','SECURITY','10','NUMBER','Minimum administrator password length','管理员密码最小长度','Minimum administrator password length',1,0,0,NULL,1,'2026-11-26 22:42:40.333','2026-11-26 22:42:40.333',NULL);
INSERT INTO `game_system_configs` VALUES (4,0,'user.default_currency','默认用户币种','Default User Currency','USER','PKR','ENUM','Default currency','默认用户币种','Default user currency',1,0,0,'[\"PKR\", \"USD\", \"EUR\"]',1,'2026-11-26 22:42:40.333','2026-11-26 22:42:40.333',NULL);
INSERT INTO `game_system_configs` VALUES (5,0,'game.default_provider','默认游戏供应商','Default Game Provider','GAME','softaggregator','STRING','Default provider','默认游戏供应商','Default game provider',1,0,0,NULL,1,'2026-11-26 22:42:40.333','2026-11-26 22:42:40.333',NULL);
INSERT INTO `game_system_configs` VALUES (6,0,'file.storage_driver','文件存储驱动','File Storage Driver','FILE_STORAGE','local','ENUM','Storage driver','文件存储驱动','File storage driver',1,0,1,'[\"local\", \"s3\", \"oss\", \"r2\"]',1,'2026-11-26 22:42:40.333','2026-11-26 22:42:40.334',NULL);
INSERT INTO `game_system_configs` VALUES (7,0,'operation.register_reward_enabled','注册奖励开关','Register Reward Enabled','OPERATION','true','BOOLEAN','Enable H5 register reward','是否启用 H5 注册奖励','Enable H5 register reward',1,0,0,NULL,1,'2026-11-26 22:42:41.164','2026-11-26 22:42:41.164',NULL);
INSERT INTO `game_system_configs` VALUES (8,0,'operation.register_reward_min_amount','注册奖励最小金额','Register Reward Min Amount','OPERATION','1.00','NUMBER','Minimum random cash reward after registration','注册后随机现金奖励最小金额','Minimum random cash reward after registration',1,0,0,NULL,1,'2026-11-26 22:42:41.164','2026-11-26 22:42:41.164',NULL);
INSERT INTO `game_system_configs` VALUES (9,0,'operation.register_reward_max_amount','注册奖励最大金额','Register Reward Max Amount','OPERATION','20.00','NUMBER','Maximum random cash reward after registration','注册后随机现金奖励最大金额','Maximum random cash reward after registration',1,0,0,NULL,1,'2026-11-26 22:42:41.164','2026-11-26 22:42:41.164',NULL);
INSERT INTO `game_system_configs` VALUES (10,0,'operation.music.playback_mode','背景音乐播放方式','Background Music Playback Mode','OPERATION','RANDOM','ENUM','Background Music Playback Mode','背景音乐播放方式','Background Music Playback Mode',1,0,0,NULL,1,'2026-11-26 22:42:41.300','2026-11-26 22:42:41.300',NULL);
INSERT INTO `game_system_configs` VALUES (11,0,'operation.music.specified_id','指定背景音乐','Specified Background Music','OPERATION','0','NUMBER','Specified Background Music','指定背景音乐','Specified Background Music',1,0,0,NULL,1,'2026-11-26 22:42:41.300','2026-11-26 22:42:41.300',NULL);
INSERT INTO `game_system_configs` VALUES (12,0,'operation.register_reward_audit_multiplier','注册奖励稽核倍数','Register Reward Audit Multiplier','OPERATION','1.0','NUMBER','Valid-bet multiplier required before withdrawing register reward','注册奖励提现前需要完成的有效投注倍数','Valid-bet multiplier required before withdrawing register reward',1,0,0,NULL,1,'2026-11-26 22:42:41.324','2026-11-26 22:42:41.324',NULL);
INSERT INTO `game_system_configs` VALUES (16,0,'basic.platform_logo_url','平台 Logo','Platform Logo','PLATFORM','','STRING','Platform admin logo','平台与站点后台左上角 Logo','Logo displayed in platform and site admin headers',1,0,0,NULL,1,'2026-11-26 22:42:47.818','2026-11-26 23:06:42.263',NULL);
INSERT INTO `game_system_configs` VALUES (17,0,'basic.platform_contact','平台联系方式','Platform Contact','PLATFORM','support@soranopro.com','STRING','Platform contact','后台页脚展示的平台联系方式','Platform contact shown in the admin footer',1,0,0,NULL,1,'2026-11-26 22:42:47.818','2026-11-26 22:42:47.818',NULL);
/*!40000 ALTER TABLE `game_system_configs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_system_health_checks` WRITE;
/*!40000 ALTER TABLE `game_system_health_checks` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_system_health_checks` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_system_retry_records` WRITE;
/*!40000 ALTER TABLE `game_system_retry_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_system_retry_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_system_task_executions` WRITE;
/*!40000 ALTER TABLE `game_system_task_executions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_system_task_executions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_system_tasks` WRITE;
/*!40000 ALTER TABLE `game_system_tasks` DISABLE KEYS */;
INSERT INTO `game_system_tasks` VALUES (1,'REPORT_DAILY_AGGREGATION','报表每日汇总','Daily Report Aggregation','REPORT','report.daily_aggregation','10 0 * * *','{\"timezone\": \"+08:00\"}',600,1,0,1,NULL,'','2026-11-27 22:42:40.339',0,'2026-11-26 22:42:40.339','2026-11-26 22:42:40.339',NULL);
INSERT INTO `game_system_tasks` VALUES (2,'ADMIN_SESSION_CLEANUP','管理员会话清理','Administrator Session Cleanup','CLEANUP','system.session_cleanup','0 * * * *','{}',120,1,0,1,NULL,'','2026-11-26 23:42:40.339',0,'2026-11-26 22:42:40.339','2026-11-26 22:42:40.339',NULL);
INSERT INTO `game_system_tasks` VALUES (3,'EXPIRED_BONUS_PROCESSING','过期奖励处理','Expired Bonus Processing','BONUS','operation.expired_bonus','15 * * * *','{}',300,1,0,1,NULL,'','2026-11-26 23:42:40.339',0,'2026-11-26 22:42:40.339','2026-11-26 22:42:40.339',NULL);
/*!40000 ALTER TABLE `game_system_tasks` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_transaction_facts` WRITE;
/*!40000 ALTER TABLE `game_transaction_facts` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_transaction_facts` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_activity_records` WRITE;
/*!40000 ALTER TABLE `game_user_activity_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_activity_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_admin_notes` WRITE;
/*!40000 ALTER TABLE `game_user_admin_notes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_admin_notes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_agent_attribution_history` WRITE;
/*!40000 ALTER TABLE `game_user_agent_attribution_history` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_agent_attribution_history` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_agent_attributions` WRITE;
/*!40000 ALTER TABLE `game_user_agent_attributions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_agent_attributions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_checkins` WRITE;
/*!40000 ALTER TABLE `game_user_checkins` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_checkins` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_devices` WRITE;
/*!40000 ALTER TABLE `game_user_devices` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_devices` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_favorite_games` WRITE;
/*!40000 ALTER TABLE `game_user_favorite_games` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_favorite_games` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_invitation_chests` WRITE;
/*!40000 ALTER TABLE `game_user_invitation_chests` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_invitation_chests` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_login_records` WRITE;
/*!40000 ALTER TABLE `game_user_login_records` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_login_records` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_notification_reads` WRITE;
/*!40000 ALTER TABLE `game_user_notification_reads` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_notification_reads` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_risk_events` WRITE;
/*!40000 ALTER TABLE `game_user_risk_events` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_risk_events` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_risk_profiles` WRITE;
/*!40000 ALTER TABLE `game_user_risk_profiles` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_risk_profiles` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_sessions` WRITE;
/*!40000 ALTER TABLE `game_user_sessions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_sessions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_settings` WRITE;
/*!40000 ALTER TABLE `game_user_settings` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_settings` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_status_histories` WRITE;
/*!40000 ALTER TABLE `game_user_status_histories` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_status_histories` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_tag_relations` WRITE;
/*!40000 ALTER TABLE `game_user_tag_relations` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_tag_relations` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_tags` WRITE;
/*!40000 ALTER TABLE `game_user_tags` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_tags` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_vip_logs` WRITE;
/*!40000 ALTER TABLE `game_user_vip_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_vip_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_wallets` WRITE;
/*!40000 ALTER TABLE `game_user_wallets` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_wallets` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_user_welfare_cards` WRITE;
/*!40000 ALTER TABLE `game_user_welfare_cards` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_user_welfare_cards` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_users` WRITE;
/*!40000 ALTER TABLE `game_users` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_users` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_vip_adjustment_logs` WRITE;
/*!40000 ALTER TABLE `game_vip_adjustment_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_vip_adjustment_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_vip_benefits` WRITE;
/*!40000 ALTER TABLE `game_vip_benefits` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_vip_benefits` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_vip_levels` WRITE;
/*!40000 ALTER TABLE `game_vip_levels` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_vip_levels` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_vip_requirements` WRITE;
/*!40000 ALTER TABLE `game_vip_requirements` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_vip_requirements` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_wallet_freezes` WRITE;
/*!40000 ALTER TABLE `game_wallet_freezes` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_wallet_freezes` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_wallet_transactions` WRITE;
/*!40000 ALTER TABLE `game_wallet_transactions` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_wallet_transactions` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_welfare_cards` WRITE;
/*!40000 ALTER TABLE `game_welfare_cards` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_welfare_cards` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_withdraw_accounts` WRITE;
/*!40000 ALTER TABLE `game_withdraw_accounts` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_withdraw_accounts` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_withdraw_orders` WRITE;
/*!40000 ALTER TABLE `game_withdraw_orders` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_withdraw_orders` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_withdrawal_approvals` WRITE;
/*!40000 ALTER TABLE `game_withdrawal_approvals` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_withdrawal_approvals` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `game_withdrawal_payout_logs` WRITE;
/*!40000 ALTER TABLE `game_withdrawal_payout_logs` DISABLE KEYS */;
/*!40000 ALTER TABLE `game_withdrawal_payout_logs` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `provider_reconciliation_items` WRITE;
/*!40000 ALTER TABLE `provider_reconciliation_items` DISABLE KEYS */;
/*!40000 ALTER TABLE `provider_reconciliation_items` ENABLE KEYS */;
UNLOCK TABLES;

LOCK TABLES `provider_reconciliation_runs` WRITE;
/*!40000 ALTER TABLE `provider_reconciliation_runs` DISABLE KEYS */;
/*!40000 ALTER TABLE `provider_reconciliation_runs` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

