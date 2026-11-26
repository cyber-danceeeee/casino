package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"github.com/shopspring/decimal"

	"casino/internal/database"
	"casino/internal/model"
)

func (h *gameInvitesHandler) Share(c *gin.Context) {
	db := database.GetDB()
	user, _, ok := h5AuthenticatedUser(db, c)
	if !ok {
		h5Error(c, http.StatusUnauthorized, "authentication required")
		return
	}
	tab := strings.TrimSpace(c.Query("tab"))
	if tab == "" {
		tab = "social"
	}

	var rankingPeriod model.GameRankingPeriods
	_ = db.Where("status IN ?", []int{1, 2}).
		Order("ranking_date DESC, id DESC").
		First(&rankingPeriod).Error

	var rankingRecords []model.GameRankingRecords
	if rankingPeriod.ID > 0 {
		_ = db.Where("ranking_period_id = ?", rankingPeriod.ID).
			Order("rank_no ASC").
			Limit(50).
			Find(&rankingRecords).Error
	}

	var commissionRecords []model.GameAgentCommissionRecords
	_ = db.Where("user_id = ?", user.ID).
		Order("period_start DESC, id DESC").
		Limit(20).
		Find(&commissionRecords).Error

	var agentRules []model.GameAgentCommissionRules
	_ = db.Where("status = ?", 1).Order("level_no ASC").Find(&agentRules).Error

	var directReferrals []model.GameUsers
	_ = db.Select("id", "user_no", "phone_country_code", "phone", "nickname", "avatar_url", "created_at").
		Where("referrer_user_id = ? AND deleted_at IS NULL", user.ID).
		Order("created_at DESC").
		Find(&directReferrals).Error
	directIDs := make([]uint64, 0, len(directReferrals))
	for _, referral := range directReferrals {
		directIDs = append(directIDs, referral.ID)
	}
	var indirectCount int64
	if len(directIDs) > 0 {
		_ = db.Model(&model.GameUsers{}).
			Where("referrer_user_id IN ? AND deleted_at IS NULL", directIDs).
			Count(&indirectCount).Error
	}

	var chestConfigs []model.GameInvitationChestConfigs
	_ = db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&chestConfigs).Error

	var wheelPrizes []model.GameInvitationWheelPrizes
	_ = db.Where("status = ?", 1).Order("sort ASC, id ASC").Find(&wheelPrizes).Error

	response.Success(c, gin.H{
		"tab":         tab,
		"inviteCode":  user.InviteCode,
		"inviteLink":  fmt.Sprintf("%s/register?inviteCode=%s", h5FrontendOrigin(c), user.InviteCode),
		"socialLinks": h5SocialLinks(c, db),
		"actions":     h5ConfigJSONAssets(c, db, "h5.share.actions", []gin.H{}),
		"homeEntries": h5ConfigJSONAssets(c, db, "h5.share.home_entries", []gin.H{}),
		"ranking": gin.H{
			"period":  rankingPeriod,
			"records": h5RankingRecordsPayload(rankingRecords),
			"self":    h5RankingSelf(rankingRecords, user.ID),
		},
		"agent":             h5AgentSharePayload(user, directReferrals, indirectCount, commissionRecords, agentRules),
		"agentRules":        agentRules,
		"commissionRecords": commissionRecords,
		"chestConfigs":      h5ChestConfigsPayload(chestConfigs),
		"chestRewards":      h5ConfigJSON(c, db, "h5.invite.chest_rewards", []string{}),
		"wheelPrizes":       h5WheelPrizesPayload(wheelPrizes),
	})
}

func h5AgentSharePayload(
	user model.GameUsers,
	directReferrals []model.GameUsers,
	indirectCount int64,
	commissionRecords []model.GameAgentCommissionRecords,
	rules []model.GameAgentCommissionRules,
) gin.H {
	claimable := decimal.Zero
	settled := decimal.Zero
	validBet := decimal.Zero
	for _, record := range commissionRecords {
		amount := h5DecimalValue(record.CommissionAmount).Add(h5DecimalValue(record.BonusAmount))
		validBet = validBet.Add(h5DecimalValue(record.ValidBetAmount))
		if record.Status == 1 {
			claimable = claimable.Add(amount)
		}
		if record.Status == 2 {
			settled = settled.Add(amount)
		}
	}

	directRate := decimal.Zero
	indirectRate := decimal.Zero
	if len(rules) > 0 {
		directRate = h5DecimalValue(rules[0].CommissionRate)
		indirectRate = h5DecimalValue(rules[0].BonusRate)
	}

	referrals := make([]gin.H, 0, len(directReferrals))
	for _, referral := range directReferrals {
		referrals = append(referrals, gin.H{
			"id":           referral.ID,
			"userNo":       referral.UserNo,
			"displayName":  h5FirstNonEmpty(referral.Nickname, h5MaskedPhone(referral.Phone)),
			"phone":        h5MaskedPhone(referral.Phone),
			"avatarUrl":    h5AssetURL(referral.AvatarURL),
			"level":        "direct",
			"registeredAt": referral.CreatedAt,
		})
	}

	return gin.H{
		"profile": gin.H{
			"displayName": h5FirstNonEmpty(user.PhoneCountryCode, "+92") + " " + user.Phone,
			"userNo":      user.UserNo,
			"inviteCode":  user.InviteCode,
			"avatarUrl":   h5AssetURL(user.AvatarURL),
		},
		"summary": gin.H{
			"claimableCommission": claimable.StringFixed(2),
			"totalSettlement":     settled.StringFixed(2),
			"directSettlement":    settled.StringFixed(2),
			"indirectSettlement":  "0.00",
			"directReferrals":     len(directReferrals),
			"indirectReferrals":   indirectCount,
			"directCommission":    settled.StringFixed(2),
			"indirectCommission":  "0.00",
			"directRate":          directRate.StringFixed(4),
			"indirectRate":        indirectRate.StringFixed(4),
			"totalBettors":        len(directReferrals) + int(indirectCount),
			"validBetAmount":      validBet.StringFixed(2),
			"settlementPeriod":    time.Now().Format("20060102"),
		},
		"referrals": referrals,
	}
}

func h5RankingRecordsPayload(records []model.GameRankingRecords) []gin.H {
	result := make([]gin.H, 0, len(records))
	for _, record := range records {
		result = append(result, gin.H{
			"id":           record.ID,
			"rankNo":       record.RankNo,
			"displayName":  record.DisplayName,
			"avatarUrl":    h5AssetURL(record.AvatarURL),
			"totalAmount":  h5DecimalString(record.TotalAmount),
			"rewardAmount": h5DecimalString(record.RewardAmount),
			"status":       record.Status,
		})
	}
	return result
}

func h5ChestConfigsPayload(configs []model.GameInvitationChestConfigs) []gin.H {
	result := make([]gin.H, 0, len(configs))
	for _, config := range configs {
		result = append(result, gin.H{
			"id":              config.ID,
			"chestCode":       config.ChestCode,
			"title":           config.Title,
			"requiredInvites": config.RequiredInvites,
			"rewardCash":      h5DecimalString(config.RewardCash),
			"rewardBonus":     h5DecimalString(config.RewardBonus),
			"iconURL":         h5AssetURL(config.IconURL),
			"sort":            config.Sort,
			"status":          config.Status,
		})
	}
	return result
}

func h5WheelPrizesPayload(prizes []model.GameInvitationWheelPrizes) []gin.H {
	result := make([]gin.H, 0, len(prizes))
	for _, prize := range prizes {
		result = append(result, gin.H{
			"id":         prize.ID,
			"prizeCode":  prize.PrizeCode,
			"label":      prize.Label,
			"rewardType": prize.RewardType,
			"minAmount":  h5DecimalString(prize.MinAmount),
			"maxAmount":  h5DecimalString(prize.MaxAmount),
			"weight":     prize.Weight,
			"iconURL":    h5AssetURL(prize.IconURL),
			"sort":       prize.Sort,
			"status":     prize.Status,
		})
	}
	return result
}

func h5RankingSelf(records []model.GameRankingRecords, userID uint64) gin.H {
	for _, record := range records {
		if record.UserID == userID {
			return gin.H{
				"rankNo":       record.RankNo,
				"displayName":  record.DisplayName,
				"avatarUrl":    h5AssetURL(record.AvatarURL),
				"totalAmount":  h5DecimalString(record.TotalAmount),
				"rewardAmount": h5DecimalString(record.RewardAmount),
			}
		}
	}
	return gin.H{"rankNo": 0, "totalAmount": "0.00", "rewardAmount": "0.00"}
}
