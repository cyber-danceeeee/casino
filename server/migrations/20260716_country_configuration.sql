CREATE TABLE IF NOT EXISTS game_country_configs (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  country_code VARCHAR(8) NOT NULL,
  iso_country_code VARCHAR(8) NOT NULL DEFAULT '',
  country_name_zh VARCHAR(128) NOT NULL DEFAULT '',
  country_name_en VARCHAR(128) NOT NULL DEFAULT '',
  currency_code VARCHAR(8) NOT NULL,
  currency_name_zh VARCHAR(128) NOT NULL DEFAULT '',
  currency_name_en VARCHAR(128) NOT NULL DEFAULT '',
  currency_symbol VARCHAR(16) NOT NULL DEFAULT '',
  currency_decimal_places TINYINT UNSIGNED NOT NULL DEFAULT 2,
  timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
  default_language VARCHAR(16) NOT NULL DEFAULT 'en',
  supported_languages JSON NULL,
  phone_country_code VARCHAR(8) NOT NULL DEFAULT '+92',
  phone_length TINYINT UNSIGNED NOT NULL DEFAULT 10,
  branding_config JSON NULL,
  theme_config JSON NULL,
  status TINYINT(1) NOT NULL DEFAULT 1,
  is_default TINYINT(1) NOT NULL DEFAULT 0,
  created_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
  deleted_at DATETIME(3) NULL,
  PRIMARY KEY (id),
  UNIQUE KEY uk_country_config_code (country_code),
  KEY idx_country_config_status_default (status,is_default),
  KEY idx_country_config_currency (currency_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO game_country_configs
(country_code,iso_country_code,country_name_zh,country_name_en,currency_code,currency_name_zh,currency_name_en,currency_symbol,currency_decimal_places,timezone,default_language,supported_languages,phone_country_code,phone_length,branding_config,theme_config,status,is_default,created_at,updated_at)
VALUES
('PK','PK','巴基斯坦','Pakistan','PKR','巴基斯坦卢比','Pakistani Rupee','₨',2,'Asia/Karachi','en',JSON_ARRAY('en','ur'),'+92',10,JSON_OBJECT('logoUrl','/static/h5/images/brand-56mvp.avif','headerLogoUrl','/static/h5/images/brand-56mvp.avif','loadingLogoUrl','/static/h5/images/brand-56mvp.avif','appIconUrl','/static/h5/images/brand-56mvp.avif','loadingAnimationUrl','/static/lottie/loading.json'),JSON_OBJECT('primaryColor','#3d8aef','buttonColor','#2f69d1','backgroundColor','#050b18','textColor','#ffffff'),1,1,NOW(3),NOW(3)),
('US','US','美国','United States','USD','美元','US Dollar','$',2,'America/New_York','en',JSON_ARRAY('en'),'+1',10,JSON_OBJECT('logoUrl','/static/h5/images/brand-56mvp.avif','headerLogoUrl','/static/h5/images/brand-56mvp.avif','loadingLogoUrl','/static/h5/images/brand-56mvp.avif'),JSON_OBJECT('primaryColor','#3d8aef','buttonColor','#2f69d1'),1,0,NOW(3),NOW(3)),
('MX','MX','墨西哥','Mexico','MXN','墨西哥比索','Mexican Peso','$',2,'America/Mexico_City','es',JSON_ARRAY('es','en'),'+52',10,JSON_OBJECT('logoUrl','/static/h5/images/brand-56mvp.avif','headerLogoUrl','/static/h5/images/brand-56mvp.avif','loadingLogoUrl','/static/h5/images/brand-56mvp.avif'),JSON_OBJECT('primaryColor','#3d8aef','buttonColor','#2f69d1'),1,0,NOW(3),NOW(3)),
('JP','JP','日本','Japan','JPY','日元','Japanese Yen','¥',0,'Asia/Tokyo','ja',JSON_ARRAY('ja','en'),'+81',10,JSON_OBJECT('logoUrl','/static/h5/images/brand-56mvp.avif','headerLogoUrl','/static/h5/images/brand-56mvp.avif','loadingLogoUrl','/static/h5/images/brand-56mvp.avif'),JSON_OBJECT('primaryColor','#3d8aef','buttonColor','#2f69d1'),1,0,NOW(3),NOW(3)),
('TH','TH','泰国','Thailand','THB','泰铢','Thai Baht','฿',2,'Asia/Bangkok','th',JSON_ARRAY('th','en'),'+66',9,JSON_OBJECT('logoUrl','/static/h5/images/brand-56mvp.avif','headerLogoUrl','/static/h5/images/brand-56mvp.avif','loadingLogoUrl','/static/h5/images/brand-56mvp.avif'),JSON_OBJECT('primaryColor','#3d8aef','buttonColor','#2f69d1'),1,0,NOW(3),NOW(3)),
('ID','ID','印度尼西亚','Indonesia','IDR','印尼盾','Indonesian Rupiah','Rp',0,'Asia/Jakarta','id',JSON_ARRAY('id','en'),'+62',10,JSON_OBJECT('logoUrl','/static/h5/images/brand-56mvp.avif','headerLogoUrl','/static/h5/images/brand-56mvp.avif','loadingLogoUrl','/static/h5/images/brand-56mvp.avif'),JSON_OBJECT('primaryColor','#3d8aef','buttonColor','#2f69d1'),1,0,NOW(3),NOW(3))
ON DUPLICATE KEY UPDATE
  country_name_zh=VALUES(country_name_zh),
  country_name_en=VALUES(country_name_en),
  currency_code=VALUES(currency_code),
  currency_symbol=VALUES(currency_symbol),
  currency_decimal_places=VALUES(currency_decimal_places),
  timezone=VALUES(timezone),
  default_language=VALUES(default_language),
  phone_country_code=VALUES(phone_country_code),
  phone_length=VALUES(phone_length),
  updated_at=NOW(3);

INSERT IGNORE INTO game_admin_permissions
(code,name,name_zh,name_en,module,permission_type,enabled,system_protected)
VALUES
('system:country:list','List country configuration','国家配置列表','List country configuration','system','API',1,1),
('system:country:view','View country configuration','查看国家配置','View country configuration','system','API',1,1),
('system:country:create','Create country configuration','创建国家配置','Create country configuration','system','API',1,1),
('system:country:update','Update country configuration','更新国家配置','Update country configuration','system','API',1,1),
('system:country:delete','Delete country configuration','删除国家配置','Delete country configuration','system','ACTION',1,1),
('system:country:status','Update country configuration status','国家配置状态','Update country configuration status','system','ACTION',1,1),
('system:country:branding','Update country H5 branding','配置国家 H5 品牌','Update country H5 branding','system','ACTION',1,1),
('system:country:set-default','Set default country','设置默认国家','Set default country','system','ACTION',1,1);

INSERT IGNORE INTO game_admin_role_permissions (role_id,permission_id,created_at,updated_at)
SELECT r.id,p.id,NOW(3),NOW(3)
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code LIKE 'system:country:%'
WHERE r.code='super-admin' AND r.deleted_at IS NULL AND p.deleted_at IS NULL;

SET @system_menu_id=(SELECT id FROM game_admin_menus WHERE code='system' AND deleted_at IS NULL LIMIT 1);
INSERT IGNORE INTO game_admin_menus
(parent_id,menu_type,code,name_zh,name_en,route_path,component_path,icon,permission_code,sort_order,visible,enabled,system_protected)
VALUES
(@system_menu_id,'MENU','country-config','国家配置','Country Configuration','/system/countries','system/SystemManagementPage','Position','system:country:list',65,1,1,1);
UPDATE game_admin_menus
SET code='country-config',
    parent_id=@system_menu_id,
    menu_type='MENU',
    name_zh='国家配置',
    name_en='Country Configuration',
    route_path='/system/countries',
    component_path='system/SystemManagementPage',
    icon='Position',
    permission_code='system:country:list',
    sort_order=65,
    visible=1,
    enabled=1,
    updated_at=NOW(3)
WHERE (code='system-country-config' OR code='country-config' OR route_path='/system/countries') AND deleted_at IS NULL;

INSERT IGNORE INTO game_admin_role_menus (role_id,menu_id,created_at,updated_at)
SELECT r.id,m.id,NOW(3),NOW(3)
FROM game_admin_roles r
JOIN game_admin_menus m ON m.code='country-config'
WHERE r.code='super-admin' AND r.deleted_at IS NULL AND m.deleted_at IS NULL;
