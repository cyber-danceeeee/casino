package notificationaudience

import (
	"errors"
	"strings"

	"casino/internal/tenancy"

	"gorm.io/gorm"
)

var (
	ErrAudienceRequired    = errors.New("notification audience value is required")
	ErrUnsupportedAudience = errors.New("notification audience type is not supported")
)

func Values(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			values = append(values, trimmed)
		}
	}
	return values
}

// Users returns an aliased game_users query (alias u) for an audience.
func Users(db *gorm.DB, audienceType, audienceValue string) (*gorm.DB, error) {
	return UsersScoped(db, tenancy.Scope{}, audienceType, audienceValue)
}

// UsersScoped returns an aliased game_users query constrained to a tenant/site.
func UsersScoped(db *gorm.DB, scope tenancy.Scope, audienceType, audienceValue string) (*gorm.DB, error) {
	values := Values(audienceValue)
	users := tenancy.ApplySite(db.Table("game_users AS u"), "u", scope).Where("u.deleted_at IS NULL")
	typ := strings.ToUpper(strings.TrimSpace(audienceType))
	if typ != "ALL_USERS" && len(values) == 0 {
		return nil, ErrAudienceRequired
	}

	switch typ {
	case "ALL_USERS":
	case "SELECTED_USERS":
		users = users.Where("u.id IN ? OR u.user_no IN ?", values, values)
	case "USER_TAG":
		users = users.Joins("JOIN game_user_tag_relations AS utr ON utr.user_id=u.id AND utr.deleted_at IS NULL").
			Joins("JOIN game_user_tags AS ut ON ut.id=utr.tag_id AND ut.deleted_at IS NULL").
			Where("ut.id IN ? OR ut.name IN ?", values, values).Distinct("u.id")
		users = tenancy.ApplySite(users, "utr", scope)
		users = tenancy.ApplySite(users, "ut", scope)
	case "VIP_LEVEL":
		users = users.Where("u.vip_level IN ?", values)
	case "COUNTRY":
		users = users.Where("u.country_code IN ?", values)
	case "REGISTRATION_CHANNEL":
		users = users.Where("u.registration_channel IN ?", values)
	case "RISK_STATUS":
		users = users.Joins("JOIN game_user_risk_profiles AS urp ON urp.user_id=u.id AND urp.deleted_at IS NULL").
			Where("urp.risk_status IN ?", values).Distinct("u.id")
		users = tenancy.ApplySite(users, "urp", scope)
	default:
		return nil, ErrUnsupportedAudience
	}
	return users, nil
}
