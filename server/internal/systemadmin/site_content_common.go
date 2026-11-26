package systemadmin

import (
	"crypto/sha256"
	"encoding/hex"
	"html"
	"net/url"
	"strconv"
	"strings"
	"time"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"gorm.io/gorm"

	"casino/internal/content"
)

const (
	siteContentDraft     = "DRAFT"
	siteContentPublished = "PUBLISHED"
)

var siteContentActions = map[string]bool{
	"NONE": true, "CLOSE_ONLY": true, "OPEN_INTERNAL_PAGE": true, "OPEN_EXTERNAL_URL": true,
	"OPEN_GAME": true, "OPEN_PROMOTION": true, "OPEN_DEPOSIT": true,
	"OPEN_WITHDRAW": true, "OPEN_SUPPORT": true, "OPEN_COUPON": true,
	"OPEN_VIP": true, "OPEN_DOWNLOAD": true,
}

func (s *Service) contentSiteScope(db *gorm.DB, siteID uint64, actor Actor, lock bool) (content.Scope, error) {
	row, err := s.siteRowForActor(db, siteID, actor, lock)
	if err != nil {
		return content.Scope{}, err
	}
	return content.Scope{SiteID: row.ID, Platform: false}, nil
}

type siteContentDimensions struct {
	CountryCode     string `gorm:"column:country_code"`
	CurrencyCode    string `gorm:"column:currency_code"`
	DefaultLanguage string `gorm:"column:default_language"`
}

func loadSiteContentDimensions(db *gorm.DB, scope content.Scope) (siteContentDimensions, error) {
	var row siteContentDimensions
	err := db.Table("game_site_config").Select("country_code,currency_code,default_language").
		Where("id = ? AND deleted_at IS NULL", scope.SiteID).Take(&row).Error
	if err != nil {
		return row, rowNotFound("site")
	}
	return row, nil
}

func markSiteContentDraft(tx *gorm.DB, scope content.Scope, actorID uint64) error {
	nowValue := now()
	return tx.Table("game_site_config").Where("id = ? AND deleted_at IS NULL", scope.SiteID).
		Updates(map[string]any{"config_status": siteContentDraft, "config_draft_updated_at": nowValue, "updated_by": actorID, "updated_at": nowValue}).Error
}

func normalizeContentDimension(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "ALL"
	}
	return strings.ToUpper(value)
}

func normalizeContentVisibility(value string) (string, error) {
	value = strings.ToUpper(strings.TrimSpace(value))
	if value == "LOGGED_OUT" {
		value = "GUEST"
	}
	if value != "ALL" && value != "GUEST" && value != "LOGGED_IN" {
		return "", NewError("INVALID_VISIBILITY_SCOPE", "invalid visibility scope")
	}
	return value, nil
}

func validateContentTime(start *time.Time, end *time.Time) error {
	if start != nil && end != nil && end.Before(*start) {
		return NewError("INVALID_TIME_RANGE", "end time must be after start time")
	}
	return nil
}

func validateSiteContentAction(action string, target string) (string, string, error) {
	action = strings.ToUpper(strings.TrimSpace(action))
	target = strings.TrimSpace(target)
	if !siteContentActions[action] {
		return "", "", NewError("INVALID_ACTION_TYPE", "invalid action type")
	}
	if action == "OPEN_EXTERNAL_URL" {
		parsed, err := url.ParseRequestURI(target)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return "", "", NewError("INVALID_ACTION_TARGET", "external URL must use https")
		}
	}
	if action != "NONE" && action != "CLOSE_ONLY" && action != "OPEN_DEPOSIT" && action != "OPEN_WITHDRAW" && action != "OPEN_SUPPORT" && action != "OPEN_COUPON" && action != "OPEN_VIP" && action != "OPEN_DOWNLOAD" && target == "" {
		return "", "", NewError("ACTION_TARGET_REQUIRED", "action target is required")
	}
	if action == "NONE" || action == "CLOSE_ONLY" || action == "OPEN_DEPOSIT" || action == "OPEN_WITHDRAW" || action == "OPEN_SUPPORT" || action == "OPEN_COUPON" || action == "OPEN_VIP" || action == "OPEN_DOWNLOAD" {
		target = ""
	}
	if action == "OPEN_INTERNAL_PAGE" && !isSafeInternalContentPath(target) {
		return "", "", NewError("INVALID_ACTION_TARGET", "internal page must be an absolute site path")
	}
	if action == "OPEN_GAME" {
		gameID, err := strconv.ParseUint(target, 10, 64)
		if err != nil || gameID == 0 {
			return "", "", NewError("INVALID_ACTION_TARGET", "game target must be a valid game ID")
		}
	}
	if action == "OPEN_PROMOTION" && !isSafeContentPromotionTarget(target) {
		return "", "", NewError("INVALID_ACTION_TARGET", "promotion target must be a safe slug or site path")
	}
	return action, target, nil
}

func isSafeInternalContentPath(target string) bool {
	if !strings.HasPrefix(target, "/") || strings.HasPrefix(target, "//") {
		return false
	}
	parsed, err := url.ParseRequestURI(target)
	return err == nil && parsed.Scheme == "" && parsed.Host == ""
}

func isSafeContentPromotionTarget(target string) bool {
	if isSafeInternalContentPath(target) {
		return true
	}
	if target == "" || strings.Contains(target, "..") {
		return false
	}
	for _, char := range target {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '/' || char == '.' {
			continue
		}
		return false
	}
	return true
}

func validateSiteImageAsset(db *gorm.DB, scope content.Scope, assetID uint64, required bool) error {
	if assetID == 0 {
		if required {
			return NewError("POPUP_ASSET_REQUIRED", "desktop popup asset is required")
		}
		return nil
	}
	var count int64
	err := db.Table("game_asset_library").Where("id=? AND asset_type='IMAGE' AND status='ACTIVE' AND deleted_at IS NULL", assetID).
		Where("(asset_scope = 'PLATFORM' AND is_public = 1) OR (site_id IN (0, ?))", scope.SiteID).Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return NewError("POPUP_ASSET_INVALID", "popup asset is missing, inactive, or outside the site scope")
	}
	return nil
}

func contentLanguageCandidates(language string) []string {
	language = strings.ToLower(strings.TrimSpace(language))
	if language == "" {
		return []string{""}
	}
	values := []string{language}
	if index := strings.IndexAny(language, "-_"); index > 0 {
		values = append(values, language[:index])
	}
	return values
}

func contentVisibilityValues(loggedIn bool) []string {
	if loggedIn {
		return []string{"ALL", "LOGGED_IN"}
	}
	return []string{"ALL", "GUEST"}
}

func sanitizeAnnouncementHTML(input string) string {
	contextNode := &xhtml.Node{Type: xhtml.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := xhtml.ParseFragment(strings.NewReader(input), contextNode)
	if err != nil {
		return html.EscapeString(strings.TrimSpace(input))
	}
	var out strings.Builder
	for _, node := range nodes {
		renderSafeAnnouncementNode(&out, node)
	}
	return strings.TrimSpace(out.String())
}

func renderSafeAnnouncementNode(out *strings.Builder, node *xhtml.Node) {
	if node.Type == xhtml.TextNode {
		out.WriteString(html.EscapeString(node.Data))
		return
	}
	if node.Type != xhtml.ElementNode {
		return
	}
	tag := strings.ToLower(node.Data)
	if tag == "script" || tag == "iframe" || tag == "style" || tag == "object" || tag == "embed" {
		return
	}
	allowed := map[string]bool{"p": true, "br": true, "strong": true, "em": true, "ul": true, "ol": true, "li": true, "a": true}
	if !allowed[tag] {
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			renderSafeAnnouncementNode(out, child)
		}
		return
	}
	out.WriteByte('<')
	out.WriteString(tag)
	if tag == "a" {
		for _, attribute := range node.Attr {
			if strings.ToLower(attribute.Key) == "href" && safeAnnouncementHref(attribute.Val) {
				out.WriteString(` href="`)
				out.WriteString(html.EscapeString(strings.TrimSpace(attribute.Val)))
				out.WriteString(`" rel="noopener noreferrer"`)
				break
			}
		}
	}
	out.WriteByte('>')
	if tag == "br" {
		return
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		renderSafeAnnouncementNode(out, child)
	}
	out.WriteString("</")
	out.WriteString(tag)
	out.WriteByte('>')
}

func safeAnnouncementHref(value string) bool {
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "/") || strings.HasPrefix(value, "#") {
		return true
	}
	parsed, err := url.ParseRequestURI(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func contentEventHash(values ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(values, "|")))
	return hex.EncodeToString(sum[:])
}
