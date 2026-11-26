INSERT IGNORE INTO game_system_configs
  (config_key,name_zh,name_en,category,config_value,value_type,description,description_zh,description_en,editable,`sensitive`,restart_required,status)
VALUES
  ('operation.register_reward_enabled','注册奖励开关','Register Reward Enabled','OPERATION','true','BOOLEAN','Enable H5 register reward','是否启用 H5 注册奖励','Enable H5 register reward',1,0,0,1),
  ('operation.register_reward_min_amount','注册奖励最小金额','Register Reward Min Amount','OPERATION','1.00','NUMBER','Minimum random cash reward after registration','注册后随机现金奖励最小金额','Minimum random cash reward after registration',1,0,0,1),
  ('operation.register_reward_max_amount','注册奖励最大金额','Register Reward Max Amount','OPERATION','20.00','NUMBER','Maximum random cash reward after registration','注册后随机现金奖励最大金额','Maximum random cash reward after registration',1,0,0,1);
