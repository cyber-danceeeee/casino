-- Repairs operation data for tenants created before private operation defaults
-- were provisioned. Platform rows remain tenant_id=0/site_id=0; all tenant
-- rows below are independent copies and can be edited without cross-tenant data.

DROP PROCEDURE IF EXISTS tenant_operation_defaults_repair_drop_index;
DROP PROCEDURE IF EXISTS tenant_operation_defaults_repair_add_currency;
DELIMITER $$
CREATE PROCEDURE tenant_operation_defaults_repair_add_currency()
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_schema=DATABASE() AND table_name='game_deposit_amount_options' AND column_name='currency_code'
  ) THEN
    ALTER TABLE game_deposit_amount_options ADD COLUMN currency_code VARCHAR(8) NOT NULL DEFAULT 'PKR' AFTER option_code;
  END IF;
END$$
CREATE PROCEDURE tenant_operation_defaults_repair_drop_index()
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=DATABASE() AND table_name='game_operation_music' AND index_name='uk_game_operation_music_code'
  ) THEN
    ALTER TABLE game_operation_music DROP INDEX uk_game_operation_music_code;
  END IF;
END$$
DELIMITER ;
CALL tenant_operation_defaults_repair_add_currency();
CALL tenant_operation_defaults_repair_drop_index();
DROP PROCEDURE tenant_operation_defaults_repair_add_currency;
DROP PROCEDURE tenant_operation_defaults_repair_drop_index;

INSERT INTO game_social_links
(tenant_id,site_id,platform_code,label,icon_url,share_url,share_title,share_text,share_description,sort,status,created_at,updated_at)
SELECT d.* FROM (
  SELECT 0 tenant_id,0 site_id,'facebook' platform_code,'Facebook' label,'/static/h5/images/social-facebook.webp' icon_url,'https://www.facebook.com/sharer/sharer.php?u={url}&quote={text}' share_url,'Join 56MVP' share_title,'Join me on 56MVP and claim daily rewards. Register here: {url}' share_text,'Invite friends, unlock rewards, and enjoy games together.' share_description,1 sort,1 status,NOW(3) created_at,NOW(3) updated_at
  UNION ALL SELECT 0,0,'messenger','Messenger','/static/h5/images/social-messenger.webp','fb-messenger://share?link={url}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',2,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'whatsapp','WhatsApp','/static/h5/images/social-whatsapp.webp','https://wa.me/?text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',3,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'telegram','Telegram','/static/h5/images/social-telegram.webp','https://t.me/share/url?url={url}&text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',4,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'instagram','Instagram','/static/h5/images/social-instagram.webp','https://www.instagram.com/','Join 56MVP','Copy my invite link and share it on Instagram: {url}','Invite friends, unlock rewards, and enjoy games together.',5,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'tiktok','TikTok','/static/h5/images/social-tiktok.webp','https://www.tiktok.com/','Join 56MVP','Copy my invite link and share it on TikTok: {url}','Invite friends, unlock rewards, and enjoy games together.',6,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'x','X','/static/h5/images/social-x.webp','https://twitter.com/intent/tweet?url={url}&text={text}','Join 56MVP','Join me on 56MVP and claim daily rewards. Register here: {url}','Invite friends, unlock rewards, and enjoy games together.',7,1,NOW(3),NOW(3)
) d
WHERE NOT EXISTS (SELECT 1 FROM game_social_links e WHERE e.tenant_id=d.tenant_id AND e.site_id=d.site_id AND e.platform_code=d.platform_code AND e.deleted_at IS NULL);

INSERT INTO game_operation_music
(tenant_id,site_id,code,name_zh,name_en,audio_url,scene,loop_play,volume_percent,sort_order,enabled,created_at,updated_at)
SELECT d.* FROM (
  SELECT 0 tenant_id,0 site_id,'lobby-bgm' code,'大厅背景音乐' name_zh,'Lobby Background Music' name_en,'/static/h5/audio/backgound_music.mp3' audio_url,'LOBBY' scene,1 loop_play,70 volume_percent,10 sort_order,1 enabled,NOW(3) created_at,NOW(3) updated_at
) d
WHERE NOT EXISTS (SELECT 1 FROM game_operation_music e WHERE e.tenant_id=d.tenant_id AND e.site_id=d.site_id AND e.code=d.code AND e.deleted_at IS NULL);

INSERT INTO game_deposit_amount_options
(tenant_id,site_id,option_code,currency_code,amount,bonus_reward,cash_reward,sort,status,created_at,updated_at)
SELECT d.* FROM (
  SELECT 0 tenant_id,0 site_id,'deposit-100' option_code,'PKR' currency_code,100 amount,30 bonus_reward,20 cash_reward,1 sort,1 status,NOW(3) created_at,NOW(3) updated_at
  UNION ALL SELECT 0,0,'deposit-300','PKR',300,100,40,2,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'deposit-500','PKR',500,170,80,3,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'deposit-1000','PKR',1000,300,200,4,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'deposit-3000','PKR',3000,1000,500,5,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'deposit-5000','PKR',5000,1700,800,6,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'deposit-10000','PKR',10000,3500,1500,7,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'deposit-30000','PKR',30000,10000,5000,8,1,NOW(3),NOW(3)
  UNION ALL SELECT 0,0,'deposit-50000','PKR',50000,17000,8000,9,1,NOW(3),NOW(3)
) d
WHERE NOT EXISTS (SELECT 1 FROM game_deposit_amount_options e WHERE e.tenant_id=d.tenant_id AND e.site_id=d.site_id AND e.option_code=d.option_code AND e.currency_code=d.currency_code AND e.deleted_at IS NULL);

-- Older site-banner rows did not store the legacy static image URL in their
-- draft payload. Backfill it so the admin can preview and edit those rows.
UPDATE game_site_banners sb
JOIN game_banners b ON b.tenant_id=0 AND b.site_id=0 AND b.deleted_at IS NULL
 AND (b.banner_code=sb.name OR b.title_en=sb.name OR b.title_zh=sb.name OR b.title=sb.name)
SET sb.draft_payload=JSON_SET(COALESCE(sb.draft_payload,JSON_OBJECT()),
    '$.desktopImageUrl',COALESCE(b.image_url,''),
    '$.mobileImageUrl',COALESCE(NULLIF(b.mobile_image_url,''),b.image_url,'')),
    sb.published_payload=JSON_SET(COALESCE(sb.published_payload,JSON_OBJECT()),
    '$.desktopImageUrl',COALESCE(b.image_url,''),
    '$.mobileImageUrl',COALESCE(NULLIF(b.mobile_image_url,''),b.image_url,''))
WHERE sb.tenant_id<>0 AND sb.site_id<>0 AND sb.deleted_at IS NULL;

INSERT INTO game_social_links
(tenant_id,site_id,platform_code,label,icon_url,share_url,share_title,share_text,share_description,sort,status,created_at,updated_at)
SELECT s.tenant_id,s.id,p.platform_code,p.label,p.icon_url,p.share_url,p.share_title,p.share_text,p.share_description,p.sort,p.status,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_social_links p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_social_links t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id AND t.platform_code=p.platform_code AND t.deleted_at IS NULL
  );

INSERT INTO game_operation_music
(tenant_id,site_id,code,name_zh,name_en,audio_url,scene,loop_play,volume_percent,sort_order,enabled,start_at,end_at,created_at,updated_at)
SELECT s.tenant_id,s.id,p.code,p.name_zh,p.name_en,p.audio_url,p.scene,p.loop_play,p.volume_percent,p.sort_order,p.enabled,p.start_at,p.end_at,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_operation_music p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_operation_music t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id AND t.code=p.code AND t.deleted_at IS NULL
  );

INSERT INTO game_deposit_amount_options
(tenant_id,site_id,option_code,currency_code,amount,bonus_reward,cash_reward,sort,status,created_at,updated_at)
SELECT s.tenant_id,s.id,p.option_code,COALESCE(NULLIF(UPPER(s.currency_code),''),'PKR'),p.amount,p.bonus_reward,p.cash_reward,p.sort,p.status,NOW(3),NOW(3)
FROM game_saas_sites s
JOIN game_deposit_amount_options p ON p.tenant_id=0 AND p.site_id=0 AND p.deleted_at IS NULL
WHERE s.tenant_id<>0 AND s.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM game_deposit_amount_options t
    WHERE t.tenant_id=s.tenant_id AND t.site_id=s.id
      AND t.currency_code=COALESCE(NULLIF(UPPER(s.currency_code),''),'PKR')
      AND t.option_code=p.option_code AND t.deleted_at IS NULL
  );
