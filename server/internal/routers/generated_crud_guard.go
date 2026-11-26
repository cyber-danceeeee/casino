package routers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"casino/internal/adminauth"
)

var generatedCRUDResources = map[string]struct{}{
	"activities":               {},
	"activity-assets":          {},
	"activity-reward-rules":    {},
	"activity-rules":           {},
	"agent-commission-records": {},
	"agent-commission-rules":   {},
	"app-download-records":     {},
	"app-versions":             {},
	"banners":                  {},
	"bet-records":              {},
	"bet-reports":              {},
	"bonus-records":            {},
	"cash-audit-records":       {},
	"cms-pages":                {},
	"deposit-orders":           {},
	"game-categories":          {},
	"game-category-relations":  {},
	"game-launch-logs":         {},
	"game-providers":           {},
	"games":                    {},
	"home-feature-cards":       {},
	"invitation-chest-configs": {},
	"invitation-wheel-prizes":  {},
	"invitation-wheel-rounds":  {},
	"invitation-wheel-spins":   {},
	"invites":                  {},
	"notices":                  {},
	"payment-channels":         {},
	"payment-methods":          {},
	"ranking-periods":          {},
	"ranking-records":          {},
	"rebate-jackpots":          {},
	"red-packet-claims":        {},
	"red-packet-rounds":        {},
	"sms-verification-codes":   {},
	"support-categories":       {},
	"support-messages":         {},
	"support-questions":        {},
	"support-tickets":          {},
	"system-configs":           {},
	"user-activity-records":    {},
	"user-checkins":            {},
	"user-devices":             {},
	"user-favorite-games":      {},
	"user-invitation-chests":   {},
	"user-sessions":            {},
	"user-settings":            {},
	"user-vip-logs":            {},
	"user-wallets":             {},
	"user-welfare-cards":       {},
	"users":                    {},
	"vip-levels":               {},
	"vip-requirements":         {},
	"wallet-transactions":      {},
	"welfare-cards":            {},
	"withdraw-accounts":        {},
	"withdraw-orders":          {},
}

func generatedCRUDGuard(prefix string) gin.HandlerFunc {
	prefix = "/" + strings.Trim(prefix, "/")
	return func(c *gin.Context) {
		if !isGeneratedCRUDRequest(c, prefix) {
			c.Next()
			return
		}
		actor, ok := adminauth.FromContext(c)
		if !ok {
			actor, ok = adminauth.ResolveActor(c)
			if ok {
				adminauth.StoreActor(c, actor)
			}
		}
		if !ok || !actor.Can("*") {
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "admin authentication required",
				"data": gin.H{"errorCode": "ADMIN_AUTHENTICATION_REQUIRED"},
			})
			return
		}
		c.Next()
	}
}

func isGeneratedCRUDRequest(c *gin.Context, prefix string) bool {
	path := strings.Trim(c.Request.URL.Path, "/")
	base := strings.Trim(prefix, "/")
	if base != "" {
		if path == base {
			return false
		}
		if !strings.HasPrefix(path, base+"/") {
			return false
		}
		path = strings.TrimPrefix(path, base+"/")
	}
	if path == "" || strings.HasPrefix(path, "admin/") {
		return false
	}
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 {
		return false
	}
	if _, ok := generatedCRUDResources[parts[0]]; !ok {
		return false
	}
	switch c.Request.Method {
	case http.MethodPost:
		return len(parts) == 1 || (len(parts) == 2 && parts[1] == "list")
	case http.MethodPut, http.MethodDelete:
		return len(parts) == 2 && isPathID(parts[1])
	case http.MethodGet:
		if parts[0] == "games" || parts[0] == "activities" {
			return false
		}
		return len(parts) == 2 && isPathID(parts[1])
	default:
		return false
	}
}

func isPathID(value string) bool {
	if value == "" {
		return false
	}
	_, err := strconv.ParseUint(value, 10, 64)
	return err == nil
}
