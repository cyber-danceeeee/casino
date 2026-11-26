INSERT INTO game_system_configs
  (config_key,name_zh,name_en,category,config_value,value_type,description,description_zh,description_en,editable,`sensitive`,restart_required,status)
VALUES
  ('operation.music.playback_mode','背景音乐播放方式','Background Music Playback Mode','OPERATION','RANDOM','ENUM','Background Music Playback Mode','背景音乐播放方式','Background Music Playback Mode',1,0,0,1),
  ('operation.music.specified_id','指定背景音乐','Specified Background Music','OPERATION','0','NUMBER','Specified Background Music','指定背景音乐','Specified Background Music',1,0,0,1)
ON DUPLICATE KEY UPDATE
  name_zh=VALUES(name_zh),name_en=VALUES(name_en),category=VALUES(category),value_type=VALUES(value_type),updated_at=NOW(3);
