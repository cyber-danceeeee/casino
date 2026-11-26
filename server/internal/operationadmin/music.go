package operationadmin

import (
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"casino/internal/content"
	"casino/internal/model"
	"casino/internal/operationdefaults"
	"casino/internal/storage"
)

const (
	musicPlaybackModeKey = "operation.music.playback_mode"
	musicSpecifiedIDKey  = "operation.music.specified_id"
)

type MusicPlaybackOption struct {
	operationdefaults.Metadata
	ID      uint64 `json:"id"`
	NameZh  string `json:"nameZh"`
	NameEn  string `json:"nameEn"`
	Scene   string `json:"scene"`
	Enabled bool   `json:"enabled"`
}

type musicRecord struct {
	model.GameOperationMusic
	operationdefaults.Metadata
}

type MusicPlaybackConfig struct {
	Mode             string                `json:"mode"`
	SpecifiedMusicID uint64                `json:"specifiedMusicID"`
	MusicOptions     []MusicPlaybackOption `json:"musicOptions"`
}

type MusicPlaybackConfigInput struct {
	Mode             string `json:"mode" binding:"required,oneof=RANDOM SPECIFIED"`
	SpecifiedMusicID uint64 `json:"specifiedMusicID"`
	Reason           string `json:"reason" binding:"max=500"`
}

type H5MusicTrack struct {
	ID            uint64 `json:"id"`
	Code          string `json:"code"`
	NameZh        string `json:"nameZh"`
	NameEn        string `json:"nameEn"`
	AudioURL      string `json:"audioUrl"`
	Loop          bool   `json:"loop"`
	VolumePercent uint   `json:"volumePercent"`
}

type MusicInput struct {
	Code          string     `json:"code" binding:"required,max=64"`
	NameZh        string     `json:"nameZh" binding:"required,max=128"`
	NameEn        string     `json:"nameEn" binding:"required,max=128"`
	AudioURL      string     `json:"audioUrl" binding:"required,max=500"`
	Scene         string     `json:"scene" binding:"required,oneof=LOBBY LOGIN GAME_LOADING ACTIVITY GLOBAL"`
	Loop          bool       `json:"loop"`
	VolumePercent uint       `json:"volumePercent" binding:"lte=100"`
	SortOrder     int        `json:"sortOrder"`
	Enabled       bool       `json:"enabled"`
	StartTime     *time.Time `json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
}

func (s *Service) ListMusic(q ListQuery) (PageResult, error) {
	var rows []musicRecord
	db := content.ApplySite(s.db.Model(&model.GameOperationMusic{}), "", q.Scope).Where("deleted_at IS NULL")
	db = contains(db, "code", q.Values["code"])
	if name := strings.TrimSpace(q.Values["name"]); name != "" {
		db = content.ApplySite(s.db.Model(&model.GameOperationMusic{}), "", q.Scope).
			Where("deleted_at IS NULL").
			Where("code LIKE ? OR name_zh LIKE ? OR name_en LIKE ?", "%"+name+"%", "%"+name+"%", "%"+name+"%")
	}
	db = equal(db, "scene", q.Values["scene"])
	if !q.Scope.SiteScoped() {
		db = boolFilter(db, "enabled", q.Values["enabled"])
		result, err := page(db.Order("id DESC"), q, &rows)
		for i := range rows {
			signMusicAssets(&rows[i].GameOperationMusic)
		}
		result.List = rows
		return result, err
	}
	db = boolFilter(db, "enabled", q.Values["enabled"])
	result, err := page(db.Order("id DESC"), q, &rows)
	if err != nil {
		return result, err
	}
	for i := range rows {
		rows[i].Metadata = operationdefaults.Metadata{CanEdit: true, CanDelete: true, Visible: rows[i].Enabled}
		signMusicAssets(&rows[i].GameOperationMusic)
	}
	result.List = rows
	return result, nil
}

func (s *Service) GetMusicPlaybackConfig(scope content.Scope) (MusicPlaybackConfig, error) {
	config := MusicPlaybackConfig{Mode: "RANDOM", MusicOptions: []MusicPlaybackOption{}}
	rows, err := s.systemConfigRows([]string{musicPlaybackModeKey, musicSpecifiedIDKey}, scope)
	if err != nil {
		return config, err
	}
	for _, row := range rows {
		switch row.ConfigKey {
		case musicPlaybackModeKey:
			mode := strings.ToUpper(strings.TrimSpace(row.ConfigValue))
			if mode == "RANDOM" || mode == "SPECIFIED" {
				config.Mode = mode
			}
		case musicSpecifiedIDKey:
			config.SpecifiedMusicID, _ = strconv.ParseUint(strings.TrimSpace(row.ConfigValue), 10, 64)
		}
	}
	result, err := s.ListMusic(ListQuery{Page: 1, PageSize: 100, Scope: scope, Values: map[string]string{}})
	if err != nil {
		return config, err
	}
	if rows, ok := result.List.([]musicRecord); ok {
		for _, row := range rows {
			if row.Scene != "LOBBY" && row.Scene != "GLOBAL" {
				continue
			}
			config.MusicOptions = append(config.MusicOptions, MusicPlaybackOption{Metadata: row.Metadata, ID: row.ID, NameZh: row.NameZh, NameEn: row.NameEn, Scene: row.Scene, Enabled: row.Enabled})
		}
	}
	return config, nil
}

func (s *Service) SaveMusicPlaybackConfig(in MusicPlaybackConfigInput, actor Actor) error {
	in.Mode = strings.ToUpper(strings.TrimSpace(in.Mode))
	if in.Mode != "RANDOM" && in.Mode != "SPECIFIED" {
		return newError("INVALID_PLAYBACK_MODE", "invalid music playback mode")
	}
	if in.Mode == "SPECIFIED" {
		if in.SpecifiedMusicID == 0 {
			return newError("MUSIC_REQUIRED", "specified music is required")
		}
		var count int64
		if err := content.ApplySite(s.db.Model(&model.GameOperationMusic{}), "", actor.Scope).
			Where("id = ? AND enabled = 1 AND scene IN ? AND deleted_at IS NULL", in.SpecifiedMusicID, []string{"LOBBY", "GLOBAL"}).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return newError("MUSIC_NOT_AVAILABLE", "specified music is not available")
		}
	}
	before, err := s.GetMusicPlaybackConfig(actor.Scope)
	if err != nil {
		return err
	}
	after := MusicPlaybackConfig{Mode: in.Mode, SpecifiedMusicID: in.SpecifiedMusicID}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := upsertMusicConfig(tx, actor.Scope, musicPlaybackModeKey, in.Mode, "ENUM"); err != nil {
			return err
		}
		if err := upsertMusicConfig(tx, actor.Scope, musicSpecifiedIDKey, strconv.FormatUint(in.SpecifiedMusicID, 10), "NUMBER"); err != nil {
			return err
		}
		return audit(tx, actor, "operation:music:update", "music.playback_config.update", "operation_music_config", 0, before, after, in.Reason)
	})
}

func (s *Service) H5MusicConfig(scope content.Scope) (map[string]any, error) {
	playback, err := s.GetMusicPlaybackConfig(scope)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	var siteRows []model.GameOperationMusic
	queryMusic := func(currentScope content.Scope) *gorm.DB {
		query := s.db.Model(&model.GameOperationMusic{}).
			Where("enabled = 1 AND deleted_at IS NULL AND scene IN ?", []string{"LOBBY", "GLOBAL"}).
			Where("(start_at IS NULL OR start_at <= ?) AND (end_at IS NULL OR end_at >= ?)", now, now)
		if currentScope.Platform {
			query = query.Where("site_id = 0")
		} else {
			query = content.ApplySite(query, "", currentScope)
		}
		return query.Order("id DESC")
	}
	if err := queryMusic(scope).Find(&siteRows).Error; err != nil {
		return nil, err
	}
	rows := siteRows
	if playback.Mode == "SPECIFIED" {
		selected := rows[:0]
		for _, row := range rows {
			if row.ID == playback.SpecifiedMusicID {
				selected = append(selected, row)
			}
		}
		if len(selected) > 0 {
			rows = selected
		}
	}
	tracks := make([]H5MusicTrack, 0, len(rows))
	for i := range rows {
		signMusicAssets(&rows[i])
		tracks = append(tracks, H5MusicTrack{
			ID: rows[i].ID, Code: rows[i].Code, NameZh: rows[i].NameZh, NameEn: rows[i].NameEn,
			AudioURL: rows[i].AudioURL, Loop: true, VolumePercent: rows[i].VolumePercent,
		})
	}
	return map[string]any{
		"mode":             playback.Mode,
		"specifiedMusicID": playback.SpecifiedMusicID,
		"tracks":           tracks,
	}, nil
}

func (s *Service) SetMusicEnabled(id uint64, enabled bool, actor Actor) error {
	var before model.GameOperationMusic
	if err := content.ApplySite(s.db.Model(&model.GameOperationMusic{}), "", actor.Scope).Where("id = ? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
		return notFound(err, "MUSIC_NOT_FOUND")
	}
	if before.Enabled == enabled {
		return nil
	}
	after := map[string]any{"enabled": enabled, "updated_at": time.Now()}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := content.ApplySite(tx.Table("game_operation_music"), "", actor.Scope).Where("id = ? AND deleted_at IS NULL", id).Updates(after).Error; err != nil {
			return err
		}
		return audit(tx, actor, "operation:music:update", "music.status.update", "operation_music", id, before, after, "")
	})
}

func upsertMusicConfig(tx *gorm.DB, scope content.Scope, key string, value string, valueType string) error {
	names := map[string][2]string{
		musicPlaybackModeKey: {"背景音乐播放方式", "Background Music Playback Mode"},
		musicSpecifiedIDKey:  {"指定背景音乐", "Specified Background Music"},
	}
	name := names[key]
	now := time.Now()
	values := map[string]any{
		"config_value": value, "name_zh": name[0], "name_en": name[1], "category": "OPERATION",
		"value_type": valueType, "description": name[1],
		"description_zh": name[0], "description_en": name[1], "editable": 1, "sensitive": 0,
		"restart_required": 0, "status": 1, "updated_at": now,
	}
	result := content.ApplySiteStrict(tx.Table("game_system_configs"), "", scope).
		Where("config_key = ? AND deleted_at IS NULL", key).
		Updates(values)
	if result.Error != nil || result.RowsAffected > 0 {
		return result.Error
	}
	row := map[string]any{
		"site_id":    scope.SiteID,
		"config_key": key, "name_zh": name[0], "name_en": name[1], "category": "OPERATION",
		"config_value": value, "value_type": valueType, "description": name[1],
		"description_zh": name[0], "description_en": name[1], "editable": 1, "sensitive": 0,
		"restart_required": 0, "status": 1, "created_at": now, "updated_at": now,
	}
	return tx.Table("game_system_configs").Clauses(clause.OnConflict{DoNothing: true}).Create(row).Error
}

func (s *Service) GetMusic(id uint64, scope content.Scope) (map[string]any, error) {
	var row model.GameOperationMusic
	if err := content.ApplySite(s.db.Model(&model.GameOperationMusic{}), "", scope).Where("id = ? AND deleted_at IS NULL", id).First(&row).Error; err != nil {
		return nil, notFound(err, "MUSIC_NOT_FOUND")
	}
	signMusicAssets(&row)
	overview := musicRecord{GameOperationMusic: row, Metadata: operationdefaults.Metadata{CanEdit: true, CanDelete: true, Visible: row.Enabled}}
	logs, _ := s.operationLogs("operation_music", id, scope)
	return map[string]any{"overview": overview, "operationLogs": logs}, nil
}

func (s *Service) SaveMusic(id uint64, in MusicInput, actor Actor) (uint64, error) {
	if err := validateTime(in.StartTime, in.EndTime); err != nil {
		return 0, err
	}
	in.Code = strings.ToLower(strings.TrimSpace(in.Code))
	if !storage.IsAssetURL(in.AudioURL) {
		return 0, newError("INVALID_AUDIO_URL", "audio must be uploaded to configured asset storage")
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var before model.GameOperationMusic
		action, permission := "create", "operation:music:create"
		if id > 0 {
			if err := content.ApplySite(tx.Model(&model.GameOperationMusic{}), "", actor.Scope).Where("id = ? AND deleted_at IS NULL", id).First(&before).Error; err != nil {
				return notFound(err, "MUSIC_NOT_FOUND")
			}
			action, permission = "update", "operation:music:update"
		}
		row := model.GameOperationMusic{SiteID: actor.Scope.SiteID, Code: in.Code, NameZh: strings.TrimSpace(in.NameZh), NameEn: strings.TrimSpace(in.NameEn), AudioURL: operationAssetRef(in.AudioURL), Scene: in.Scene, Loop: true, VolumePercent: in.VolumePercent, SortOrder: in.SortOrder, Enabled: in.Enabled, StartAt: in.StartTime, EndAt: in.EndTime}
		if id == 0 {
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
			id = row.ID
		} else if err := content.ApplySite(tx.Model(&model.GameOperationMusic{}), "", actor.Scope).Where("id = ?", id).Select("code", "name_zh", "name_en", "audio_url", "scene", "loop_play", "volume_percent", "sort_order", "enabled", "start_at", "end_at").Updates(row).Error; err != nil {
			return err
		}
		if err := content.ApplySite(tx.Model(&model.GameOperationMusic{}), "", actor.Scope).Where("id = ?", id).First(&row).Error; err != nil {
			return err
		}
		return audit(tx, actor, permission, action, "operation_music", id, before, row, "")
	})
	return id, err
}

func (s *Service) DeleteMusic(id uint64, actor Actor) error {
	return s.softDelete("game_operation_music", "operation_music", id, "operation:music:delete", actor)
}

func (s *Service) RecordAssetUpload(kind string, assetURL string, actor Actor) error {
	return audit(s.db, actor, "operation:asset:upload", "upload", "operation_asset", 0, nil, map[string]any{"kind": kind, "url": assetURL}, "")
}
