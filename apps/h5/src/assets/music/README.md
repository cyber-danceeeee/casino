大厅背景音乐放在本目录（MP3 / OGG / WAV），随 H5 构建并生成带内容哈希的资源地址。
播放列表在 `src/config/media.ts` 生成；音量、循环方式和顺序随代码维护。

原服务器配置的 `/static/h5/audio/backgound_music.mp3` 返回 404，实际文件缺失。
在取得要使用的音乐文件前保持空列表，避免请求失效音源。通知音效保留在 `assets/audio`，不会用作背景音乐。
