INSERT IGNORE INTO game_system_configs
  (config_key,name_zh,name_en,category,config_value,value_type,description,description_zh,description_en,editable,`sensitive`,restart_required,status)
VALUES
  ('operation.register_reward_enabled','注册奖励开关','Register Reward Enabled','OPERATION','true','BOOLEAN','Enable H5 register reward','是否启用 H5 注册奖励','Enable H5 register reward',1,0,0,1),
  ('operation.register_reward_min_amount','注册奖励最小金额','Register Reward Min Amount','OPERATION','1.00','NUMBER','Minimum random cash reward after registration','注册后随机现金奖励最小金额','Minimum random cash reward after registration',1,0,0,1),
  ('operation.register_reward_max_amount','注册奖励最大金额','Register Reward Max Amount','OPERATION','20.00','NUMBER','Maximum random cash reward after registration','注册后随机现金奖励最大金额','Maximum random cash reward after registration',1,0,0,1),
  ('operation.register_reward_audit_multiplier','注册奖励稽核倍数','Register Reward Audit Multiplier','OPERATION','1.0','NUMBER','Valid-bet multiplier required before withdrawing register reward','注册奖励提现前需要完成的有效投注倍数','Valid-bet multiplier required before withdrawing register reward',1,0,0,1);

INSERT IGNORE INTO game_admin_permissions
  (`code`, `name`, `name_zh`, `name_en`, `module`, `permission_type`, `enabled`, `system_protected`)
VALUES
  ('operation:register-reward:view', 'View Register Reward Config', '查看注册奖励配置', 'View Register Reward Config', 'operation', 'API', 1, 1),
  ('operation:register-reward:update', 'Update Register Reward Config', '更新注册奖励配置', 'Update Register Reward Config', 'operation', 'API', 1, 1);

INSERT IGNORE INTO game_admin_role_permissions (`role_id`, `permission_id`, `created_at`, `updated_at`)
SELECT r.id, p.id, NOW(3), NOW(3)
FROM game_admin_roles r
JOIN game_admin_permissions p ON p.code IN ('operation:register-reward:view','operation:register-reward:update')
WHERE r.code = 'super-admin';
