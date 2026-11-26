SET NAMES utf8mb4;
-- One-time removal of obsolete commercial ownership metadata. The installer
-- takes a complete backup before executing this file. Business IDs and player
-- wallet/order amounts are never rewritten.
DROP PROCEDURE IF EXISTS casino_remove_legacy;
DELIMITER $$
CREATE PROCEDURE casino_remove_legacy()
BEGIN
 DECLARE selected_site BIGINT UNSIGNED DEFAULT 0;
 DECLARE selected_owner BIGINT UNSIGNED DEFAULT 0;
 DECLARE site_count INT DEFAULT 0;
 DECLARE owner_count INT DEFAULT 0;
 DECLARE conflicts BIGINT DEFAULT 0;
 DECLARE complete INT DEFAULT 0;
 DECLARE table_value VARCHAR(64);
 DECLARE index_value VARCHAR(64);
 DECLARE column_value VARCHAR(64);
 DECLARE renamed_index VARCHAR(64);
 DECLARE columns_cursor CURSOR FOR
   SELECT TABLE_NAME FROM information_schema.COLUMNS
   WHERE TABLE_SCHEMA=DATABASE() AND COLUMN_NAME='tenant_id';
 DECLARE indexes_cursor CURSOR FOR
   SELECT DISTINCT TABLE_NAME,INDEX_NAME FROM information_schema.STATISTICS
   WHERE TABLE_SCHEMA=DATABASE() AND (INDEX_NAME LIKE '%tenant%' OR INDEX_NAME LIKE '%saas%');
 DECLARE CONTINUE HANDLER FOR NOT FOUND SET complete=1;
 SELECT COUNT(*),COALESCE(MIN(id),0) INTO site_count,selected_site FROM game_saas_sites;
 SELECT COUNT(*),COALESCE(MIN(id),0) INTO owner_count,selected_owner FROM game_saas_tenants;
 IF site_count<>1 OR owner_count<>1 THEN
   SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='Upgrade requires one existing H5 dataset. Separate independent sites into their own databases first.';
 END IF;
 SELECT COUNT(*) INTO conflicts FROM game_users WHERE site_id<>selected_site;
 IF conflicts>0 THEN
   SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='Players outside the selected H5 dataset must be migrated before this upgrade.';
 END IF;
 -- Reject orphaned commercial datasets before any business table is altered.
 OPEN columns_cursor;
 check_ownership: LOOP
  FETCH columns_cursor INTO table_value;
  IF complete=1 THEN LEAVE check_ownership; END IF;
  SET @cleanup_sql=CONCAT('SELECT COUNT(*) INTO @outside_dataset FROM `',table_value,'` WHERE tenant_id NOT IN (0,',selected_owner,')');
  PREPARE cleanup_statement FROM @cleanup_sql; EXECUTE cleanup_statement; DEALLOCATE PREPARE cleanup_statement;
  IF @outside_dataset>0 THEN
   SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='Records from another independent dataset exist; upgrade stopped before removing ownership metadata.';
  END IF;
 END LOOP;
 CLOSE columns_cursor;
 SET complete=0;
 -- Keep real administration identities and permissions, remove sales controls.
 DELETE rp FROM game_admin_role_permissions rp JOIN game_admin_permissions p ON p.id=rp.permission_id
   WHERE p.code LIKE 'system:tenant:%';
 DELETE FROM game_admin_permissions WHERE code LIKE 'system:tenant:%';
 DELETE rm FROM game_admin_role_menus rm JOIN game_admin_menus m ON m.id=rm.menu_id
   WHERE m.route_path LIKE '/platform/%' OR m.route_path LIKE '/system/tenant%' OR m.route_path='/system/sites';
 DELETE FROM game_admin_menus WHERE route_path LIKE '/platform/%' OR route_path LIKE '/system/tenant%' OR route_path='/system/sites';
 UPDATE game_admin_roles SET code='site-operator',name='Site Operator',name_zh='站点运营员',name_en='Site Operator' WHERE code='tenant-owner';
 -- Remove queued jobs and history belonging only to the removed sales modules.
 DELETE d FROM event_deliveries d JOIN event_outbox o ON o.id=d.outbox_id
   WHERE o.event_type LIKE 'tenant.%' OR o.event_type LIKE 'ledger.transaction%';
 DELETE FROM event_outbox WHERE event_type LIKE 'tenant.%' OR event_type LIKE 'ledger.transaction%';
 DELETE FROM audit_event_records WHERE event_type LIKE 'tenant.%' OR event_type LIKE 'ledger.transaction%';
 DELETE FROM game_admin_operation_logs WHERE permission LIKE 'system:tenant:%'
   OR target_type IN ('tenant','tenant_package','tenant_ledger','brand_delivery','domain');
 -- These were commercial-owner ledgers, not player wallet transactions.
 DROP TABLE IF EXISTS ledger_entries,ledger_transactions,ledger_balance_snapshots,ledger_accounts,tenant_credit_limits;
 DROP TABLE IF EXISTS game_saas_brand_deliveries,game_saas_domains,game_saas_tenant_packages,game_tenant_onboarding_email_jobs;
 DROP TABLE IF EXISTS casino_site_binding,dashboard_event_facts;
 RENAME TABLE game_saas_sites TO game_site_config,
   game_tenant_popup_campaigns TO game_popup_campaigns,
   game_tenant_maintenance TO game_site_maintenance,
   game_tenant_operation_overrides TO game_content_visibility;
 OPEN columns_cursor;
 cleanup_columns: LOOP
   FETCH columns_cursor INTO table_value;
   IF complete=1 THEN LEAVE cleanup_columns; END IF;
   SET @cleanup_sql=CONCAT('ALTER TABLE `',table_value,'` DROP COLUMN tenant_id');
   PREPARE cleanup_statement FROM @cleanup_sql; EXECUTE cleanup_statement; DEALLOCATE PREPARE cleanup_statement;
 END LOOP;
 CLOSE columns_cursor;
 DROP TABLE IF EXISTS game_saas_tenants;
 -- A content record may refer either to the provider/default catalog (0) or
 -- this H5 configuration (1). No client-selectable ownership remains.
 IF selected_site<>1 THEN
   BEGIN
    DECLARE done_sites INT DEFAULT 0;
    DECLARE scoped_table VARCHAR(64);
    DECLARE scoped_cursor CURSOR FOR SELECT TABLE_NAME FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND COLUMN_NAME='site_id';
    DECLARE CONTINUE HANDLER FOR NOT FOUND SET done_sites=1;
    OPEN scoped_cursor;
    normalize_sites: LOOP
     FETCH scoped_cursor INTO scoped_table;
     IF done_sites=1 THEN LEAVE normalize_sites; END IF;
     SET @cleanup_sql=CONCAT('UPDATE `',scoped_table,'` SET site_id=1 WHERE site_id=',selected_site);
     PREPARE cleanup_statement FROM @cleanup_sql; EXECUTE cleanup_statement; DEALLOCATE PREPARE cleanup_statement;
    END LOOP;
    CLOSE scoped_cursor;
   END;
   UPDATE game_site_config SET id=1 WHERE id=selected_site;
 END IF;
 ALTER TABLE game_site_config MODIFY id BIGINT UNSIGNED NOT NULL;
 ALTER TABLE game_site_config ADD CONSTRAINT chk_single_h5_config CHECK (id=1);
ALTER TABLE game_site_maintenance RENAME COLUMN allow_tenant_bypass TO allow_admin_bypass, DROP COLUMN allow_platform_bypass;
UPDATE game_site_maintenance SET scope_type='SITE' WHERE scope_type='TENANT';
ALTER TABLE game_asset_library ALTER COLUMN asset_scope SET DEFAULT 'SITE';
UPDATE game_asset_library SET asset_scope='SITE' WHERE asset_scope='TENANT';
UPDATE game_backup_tasks SET scope_type='H5_CONFIGURATION' WHERE scope_type='TENANT_CONFIGURATION';
UPDATE game_admin_permissions SET code=REPLACE(code,'maintenance:tenant:','maintenance:site:') WHERE code LIKE 'maintenance:tenant:%';
ALTER TABLE game_transaction_facts DROP COLUMN ledger_transaction_id, DROP COLUMN ledger_amount_minor;
ALTER TABLE provider_reconciliation_items DROP COLUMN ledger_transaction_id;
ALTER TABLE provider_reconciliation_runs DROP COLUMN ledger_transaction_count;

DELETE rp FROM game_admin_role_permissions rp JOIN game_admin_permissions p ON p.id=rp.permission_id WHERE p.code LIKE 'system:domain:%' OR p.code IN ('system:site:list','system:site:create','system:site:delete');
DELETE FROM game_admin_permissions WHERE code LIKE 'system:domain:%' OR code IN ('system:site:list','system:site:create','system:site:delete');
UPDATE game_admin_permissions SET name=REPLACE(REPLACE(name,'tenant','site'),'SaaS ','H5 '),name_en=REPLACE(REPLACE(name_en,'Tenant','Site'),'SaaS ','H5 '),name_zh=REPLACE(name_zh,'租户','站点');
UPDATE game_admin_menus SET permission_code=REPLACE(permission_code,'maintenance:tenant:','maintenance:site:');
UPDATE game_system_configs SET description=REPLACE(description,'tenant','site'),description_en=REPLACE(description_en,'tenant','site'),description_zh=REPLACE(description_zh,'租户','站点');
DELETE FROM game_system_configs WHERE config_key='finance.platform_usdt_address';

 SET complete=0;
 OPEN indexes_cursor;
 cleanup_indexes: LOOP
   FETCH indexes_cursor INTO table_value,index_value;
   IF complete=1 THEN LEAVE cleanup_indexes; END IF;
   SET renamed_index=CONCAT('idx_clean_',SUBSTRING(SHA2(CONCAT(table_value,index_value),256),1,16));
   SET @cleanup_sql=CONCAT('ALTER TABLE `',table_value,'` RENAME INDEX `',index_value,'` TO `',renamed_index,'`');
   PREPARE cleanup_statement FROM @cleanup_sql; EXECUTE cleanup_statement; DEALLOCATE PREPARE cleanup_statement;
 END LOOP;
 CLOSE indexes_cursor;
END$$
DELIMITER ;
CALL casino_remove_legacy();
DROP PROCEDURE casino_remove_legacy;
