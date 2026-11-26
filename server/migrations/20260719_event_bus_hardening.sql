DROP PROCEDURE IF EXISTS event_bus_add_column;
DROP PROCEDURE IF EXISTS event_bus_modify_column;
DROP PROCEDURE IF EXISTS event_bus_add_index;
DROP PROCEDURE IF EXISTS event_bus_add_unique_index;
DROP PROCEDURE IF EXISTS event_bus_drop_index;
DROP PROCEDURE IF EXISTS event_bus_drop_check;
DROP PROCEDURE IF EXISTS event_bus_add_check;
DELIMITER $$
CREATE PROCEDURE event_bus_add_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD COLUMN `',column_name_value,'` ',definition_value);
    PREPARE event_bus_statement FROM @statement_value; EXECUTE event_bus_statement; DEALLOCATE PREPARE event_bus_statement;
  END IF;
END$$
CREATE PROCEDURE event_bus_modify_column(IN table_name_value VARCHAR(64), IN column_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema=DATABASE() AND table_name=table_name_value AND column_name=column_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` MODIFY COLUMN `',column_name_value,'` ',definition_value);
    PREPARE event_bus_statement FROM @statement_value; EXECUTE event_bus_statement; DEALLOCATE PREPARE event_bus_statement;
  END IF;
END$$
CREATE PROCEDURE event_bus_add_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD INDEX `',index_name_value,'` ',definition_value);
    PREPARE event_bus_statement FROM @statement_value; EXECUTE event_bus_statement; DEALLOCATE PREPARE event_bus_statement;
  END IF;
END$$
CREATE PROCEDURE event_bus_add_unique_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD UNIQUE KEY `',index_name_value,'` ',definition_value);
    PREPARE event_bus_statement FROM @statement_value; EXECUTE event_bus_statement; DEALLOCATE PREPARE event_bus_statement;
  END IF;
END$$
CREATE PROCEDURE event_bus_drop_index(IN table_name_value VARCHAR(64), IN index_name_value VARCHAR(64))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND EXISTS (SELECT 1 FROM information_schema.statistics WHERE table_schema=DATABASE() AND table_name=table_name_value AND index_name=index_name_value) THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` DROP INDEX `',index_name_value,'`');
    PREPARE event_bus_statement FROM @statement_value; EXECUTE event_bus_statement; DEALLOCATE PREPARE event_bus_statement;
  END IF;
END$$
CREATE PROCEDURE event_bus_drop_check(IN table_name_value VARCHAR(64), IN constraint_name_value VARCHAR(64))
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema=DATABASE() AND table_name=table_name_value AND constraint_name=constraint_name_value AND constraint_type='CHECK') THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` DROP CHECK `',constraint_name_value,'`');
    PREPARE event_bus_statement FROM @statement_value; EXECUTE event_bus_statement; DEALLOCATE PREPARE event_bus_statement;
  END IF;
END$$
CREATE PROCEDURE event_bus_add_check(IN table_name_value VARCHAR(64), IN constraint_name_value VARCHAR(64), IN definition_value TEXT)
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema=DATABASE() AND table_name=table_name_value)
     AND NOT EXISTS (SELECT 1 FROM information_schema.table_constraints WHERE table_schema=DATABASE() AND table_name=table_name_value AND constraint_name=constraint_name_value AND constraint_type='CHECK') THEN
    SET @statement_value=CONCAT('ALTER TABLE `',table_name_value,'` ADD CONSTRAINT `',constraint_name_value,'` CHECK ',definition_value);
    PREPARE event_bus_statement FROM @statement_value; EXECUTE event_bus_statement; DEALLOCATE PREPARE event_bus_statement;
  END IF;
END$$
DELIMITER ;

CALL event_bus_add_column('event_outbox','event_version','INT NOT NULL DEFAULT 1 AFTER `event_type`');
CALL event_bus_add_column('event_outbox','trace_id','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `idempotency_key`');
CALL event_bus_add_column('event_outbox','causation_id','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `correlation_id`');
CALL event_bus_add_column('event_outbox','locked_by','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `published_at`');
CALL event_bus_add_column('event_outbox','locked_at','DATETIME(3) NULL AFTER `locked_by`');
CALL event_bus_add_column('event_outbox','occurred_at','DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) AFTER `last_error`');
CALL event_bus_add_index('event_outbox','idx_event_outbox_processing','(status, locked_at, id)');
CALL event_bus_add_index('event_outbox','idx_event_outbox_trace','(trace_id, correlation_id, causation_id)');

CALL event_bus_add_column('event_deliveries','consumer_name','VARCHAR(128) NOT NULL DEFAULT '''' AFTER `consumer`');
CALL event_bus_modify_column('event_deliveries','consumer','VARCHAR(128) NOT NULL DEFAULT ''''');
UPDATE event_deliveries SET consumer_name=consumer WHERE consumer_name='' AND consumer<>'';
CALL event_bus_drop_index('event_deliveries','uk_event_delivery_consumer');
CALL event_bus_add_unique_index('event_deliveries','uk_event_delivery_consumer','(event_id, consumer_name)');
CALL event_bus_add_index('event_deliveries','idx_event_delivery_consumer_name_status','(consumer_name, status, updated_at)');

UPDATE event_outbox SET status='PROCESSING' WHERE status='DISPATCHING';
UPDATE event_outbox SET event_version=1 WHERE event_version IS NULL OR event_version=0;
UPDATE event_outbox SET occurred_at=created_at WHERE occurred_at IS NULL;

CALL event_bus_drop_check('event_outbox','chk_event_outbox_status');
CALL event_bus_drop_check('event_deliveries','chk_event_delivery_status');
CALL event_bus_add_check('event_outbox','chk_event_outbox_status','(status IN (''PENDING'',''PROCESSING'',''PUBLISHED'',''FAILED'',''DEAD''))');
CALL event_bus_add_check('event_deliveries','chk_event_delivery_status','(status IN (''PENDING'',''SUCCESS'',''FAILED''))');

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected,created_at,updated_at)
VALUES
('system:event:replay','Replay event outbox','回放事件','Replay event outbox','system','API',1,1,NOW(3),NOW(3));

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at,deleted_at)
SELECT r.id,p.id,NOW(3),NOW(3),NULL
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN ('system:event:replay')
WHERE r.code='super-admin' AND r.tenant_id=0 AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

DROP PROCEDURE IF EXISTS event_bus_add_column;
DROP PROCEDURE IF EXISTS event_bus_modify_column;
DROP PROCEDURE IF EXISTS event_bus_add_index;
DROP PROCEDURE IF EXISTS event_bus_add_unique_index;
DROP PROCEDURE IF EXISTS event_bus_drop_index;
DROP PROCEDURE IF EXISTS event_bus_drop_check;
DROP PROCEDURE IF EXISTS event_bus_add_check;
