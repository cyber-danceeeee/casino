package systemadmin

import (
	"context"

	"casino/internal/eventbus"
	"casino/internal/tenancy"
)

type EventDispatchInput struct {
	Limit int `json:"limit"`
}

type EventReplayInput struct {
	Force  bool   `json:"force"`
	Reason string `json:"reason"`
}

type EventReplayRangeInput struct {
	FromID    uint64 `json:"fromID"`
	ToID      uint64 `json:"toID"`
	EventType string `json:"eventType"`
	Status    string `json:"status"`
	Limit     int    `json:"limit"`
	Force     bool   `json:"force"`
	Reason    string `json:"reason"`
}

type EventReplayConsumerInput struct {
	Consumer string `json:"consumer"`
	Limit    int    `json:"limit"`
	Force    bool   `json:"force"`
	Reason   string `json:"reason"`
}

func (s *Service) ListEvents(q Query, actor Actor) (Result, error) {
	rows, total, err := eventbus.New(s.db).List(context.Background(), eventScope(actor), q.Values, q.Page, q.PageSize)
	if err != nil {
		return Result{}, eventError(err)
	}
	p, size := page(q)
	return Result{List: rows, Total: total, Page: p, PageSize: size}, nil
}

func (s *Service) EventDetail(id uint64, actor Actor) (map[string]any, error) {
	record, deliveries, err := eventbus.New(s.db).Get(context.Background(), id, eventScope(actor))
	if err != nil {
		return nil, eventError(err)
	}
	return map[string]any{"event": record, "deliveries": deliveries}, nil
}

func (s *Service) DispatchEvents(input EventDispatchInput, actor Actor) (any, error) {
	if err := s.requireEventSuperAdmin(actor); err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}
	eventbus.SignalDispatch()
	result := map[string]any{"accepted": true, "limit": limit, "mode": "RABBITMQ_OUTBOX_WAKEUP"}
	if err := audit(s.db, actor, "system:event:dispatch", "system.event.dispatch", "event_outbox", 0, nil, result, "manual event dispatch"); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ReplayEvent(id uint64, input EventReplayInput, actor Actor) (any, error) {
	if err := s.requireEventSuperAdmin(actor); err != nil {
		return nil, err
	}
	if err := requireReason(input.Reason); err != nil {
		return nil, err
	}
	record, err := eventbus.New(s.db).ReplayEvent(context.Background(), id, eventScope(actor), input.Force)
	if err != nil {
		return nil, eventError(err)
	}
	if err := audit(s.db, actor, "system:event:replay", "system.event.replay", "event_outbox", id, nil, map[string]any{"force": input.Force}, input.Reason); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *Service) ReplayEventRange(input EventReplayRangeInput, actor Actor) (any, error) {
	if err := s.requireEventSuperAdmin(actor); err != nil {
		return nil, err
	}
	result, err := eventbus.New(s.db).ReplayRange(context.Background(), eventScope(actor), eventbus.ReplayRangeOptions{
		FromID: input.FromID, ToID: input.ToID, EventType: input.EventType, Status: input.Status, Limit: input.Limit, Force: input.Force,
	})
	if err != nil {
		return nil, eventError(err)
	}
	if err := audit(s.db, actor, "system:event:replay", "system.event.replay_range", "event_outbox", 0, nil, map[string]any{"input": input, "result": result}, input.Reason); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) ReplayConsumer(input EventReplayConsumerInput, actor Actor) (any, error) {
	if err := s.requireEventSuperAdmin(actor); err != nil {
		return nil, err
	}
	result, err := eventbus.New(s.db).ReplayConsumer(context.Background(), eventScope(actor), eventbus.ReplayConsumerOptions{
		Consumer: input.Consumer, Limit: input.Limit, Force: input.Force,
	})
	if err != nil {
		return nil, eventError(err)
	}
	if err := audit(s.db, actor, "system:event:replay", "system.event.replay_consumer", "event_outbox", 0, nil, map[string]any{"input": input, "result": result}, input.Reason); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *Service) requireEventSuperAdmin(actor Actor) error {
	if !actor.Platform || actor.TenantID != 0 {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	if !s.db.Migrator().HasTable("game_admin_account_roles") || !s.db.Migrator().HasTable("game_admin_roles") {
		return nil
	}
	var count int64
	_ = s.db.Table("game_admin_account_roles ar").
		Joins("JOIN game_admin_roles r ON r.id=ar.role_id AND r.code='super-admin' AND r.tenant_id=0 AND r.status=1 AND r.deleted_at IS NULL").
		Where("ar.admin_id=? AND ar.deleted_at IS NULL", actor.ID).
		Count(&count).Error
	if count == 0 {
		return NewError("PERMISSION_DENIED", "permission denied")
	}
	return nil
}

func eventScope(actor Actor) tenancy.Scope {
	return tenancy.Scope{TenantID: actor.TenantID, SiteID: actor.SiteID, Platform: actor.Platform}
}

func eventError(err error) error {
	if err == nil {
		return nil
	}
	if code := eventbus.ErrorCode(err); code != "INTERNAL_ERROR" {
		return NewError(code, err.Error())
	}
	return err
}
