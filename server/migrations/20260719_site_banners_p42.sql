CREATE TABLE IF NOT EXISTS `game_site_banners` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `tenant_id` BIGINT UNSIGNED NOT NULL,
  `site_id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(128) NOT NULL DEFAULT '',
  `title` VARCHAR(256) NOT NULL DEFAULT '',
  `subtitle` VARCHAR(500) NOT NULL DEFAULT '',
  `desktop_asset_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `mobile_asset_id` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `desktop_asset_id_draft` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `mobile_asset_id_draft` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `action_type` VARCHAR(32) NOT NULL DEFAULT 'NONE',
  `action_target` VARCHAR(500) NOT NULL DEFAULT '',
  `sort_order` INT NOT NULL DEFAULT 0,
  `language_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `country_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `currency_code` VARCHAR(16) NOT NULL DEFAULT 'ALL',
  `visibility_scope` VARCHAR(24) NOT NULL DEFAULT 'ALL',
  `start_at` DATETIME(3) NULL,
  `end_at` DATETIME(3) NULL,
  `enabled` TINYINT(1) NOT NULL DEFAULT 0,
  `draft_payload` JSON NULL,
  `published_payload` JSON NULL,
  `publish_status` VARCHAR(24) NOT NULL DEFAULT 'DRAFT',
  `config_version` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `updated_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `draft_updated_at` DATETIME(3) NULL,
  `published_at` DATETIME(3) NULL,
  `published_by` BIGINT UNSIGNED NOT NULL DEFAULT 0,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  `deleted_at` DATETIME(3) NULL,
  PRIMARY KEY (`id`),
  KEY `idx_site_banner_scope_sort` (`tenant_id`,`site_id`,`publish_status`,`enabled`,`sort_order`,`id`),
  KEY `idx_site_banner_schedule` (`tenant_id`,`site_id`,`start_at`,`end_at`),
  KEY `idx_site_banner_desktop_asset` (`desktop_asset_id`),
  KEY `idx_site_banner_mobile_asset` (`mobile_asset_id`),
  KEY `idx_site_banner_desktop_asset_draft` (`desktop_asset_id_draft`),
  KEY `idx_site_banner_mobile_asset_draft` (`mobile_asset_id_draft`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO `game_site_banners`
(`tenant_id`,`site_id`,`name`,`title`,`desktop_asset_id`,`mobile_asset_id`,`desktop_asset_id_draft`,`mobile_asset_id_draft`,
 `action_type`,`action_target`,`sort_order`,`language_code`,`country_code`,`visibility_scope`,`start_at`,`end_at`,`enabled`,
 `draft_payload`,`published_payload`,`publish_status`,`config_version`,`created_at`,`updated_at`)
SELECT b.tenant_id,b.site_id,COALESCE(NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.banner_code),COALESCE(NULLIF(b.title_en,''),b.title_zh),
       0,0,0,0,
       CASE UPPER(b.target_type)
         WHEN 'EXTERNAL_URL' THEN 'OPEN_EXTERNAL_URL'
         WHEN 'INTERNAL_ROUTE' THEN 'OPEN_INTERNAL_PAGE'
         WHEN 'GAME' THEN 'OPEN_GAME'
         WHEN 'PROMOTION' THEN 'OPEN_PROMOTION'
         ELSE 'NONE'
       END,
       COALESCE(b.target_value,''),b.sort_order,COALESCE(NULLIF(b.languages,''),'ALL'),COALESCE(NULLIF(b.countries,''),'ALL'),'ALL',
       b.start_at,b.end_at,b.enabled,
       JSON_OBJECT('name',COALESCE(NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.banner_code),'title',COALESCE(NULLIF(b.title_en,''),b.title_zh),
         'subtitle','','desktopAssetID',0,'mobileAssetID',0,'actionType','NONE','actionTarget','',
         'sortOrder',b.sort_order,'languageCode',COALESCE(NULLIF(b.languages,''),'ALL'),'countryCode',COALESCE(NULLIF(b.countries,''),'ALL'),
         'currencyCode','ALL','visibilityScope','ALL','startAt',b.start_at,'endAt',b.end_at,'enabled',FALSE),
       NULL,'DRAFT',0,b.created_at,b.updated_at
FROM game_banners b
WHERE b.deleted_at IS NULL
  AND b.tenant_id > 0 AND b.site_id > 0
  AND NOT EXISTS (
    SELECT 1 FROM game_site_banners sb
    WHERE sb.tenant_id=b.tenant_id AND sb.site_id=b.site_id AND sb.name=COALESCE(NULLIF(b.title_en,''),NULLIF(b.title_zh,''),b.banner_code)
  );
