package gameadmin

import (
	"casino/internal/game/connections"
	"errors"

	"gorm.io/gorm"
)

func (s *Service) ListProviderConnections() (map[string]any, error) {
	return connections.New(s.db).List()
}
func (s *Service) SaveProviderConnection(input connections.Input, actor Actor) (connections.View, error) {
	var result connections.View
	err := s.db.Transaction(func(tx *gorm.DB) error {
		saved, err := connections.New(tx).Save(input)
		if err != nil {
			if errors.Is(err, connections.ErrConflict) {
				return newError("PROVIDER_CONFIG_CONFLICT", err.Error())
			}
			return newError("INVALID_PROVIDER_CONFIGURATION", err.Error())
		}
		result = saved
		var id uint64
		if err := tx.Table("game_game_providers").Select("id").Where("provider_code=? AND site_id=0 AND deleted_at IS NULL", saved.ProviderCode).Take(&id).Error; err != nil {
			return err
		}
		return audit(tx, actor, "game:provider:update", "save_connection", "game_provider", id, nil, saved, input.Reason)
	})
	return result, err
}
