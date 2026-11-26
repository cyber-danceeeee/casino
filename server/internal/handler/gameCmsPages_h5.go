package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"
	"gorm.io/gorm"

	"casino/internal/database"
	"casino/internal/model"
)

func (h *gameCmsPagesHandler) Page(c *gin.Context) {
	pageCode := strings.TrimSpace(c.Param("pageCode"))
	var page model.GameCmsPages
	err := database.GetDB().Where("page_code = ? AND status = ?", pageCode, 1).
		Order("version_no DESC, id DESC").
		First(&page).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Success(c, h5DefaultCMSPage(pageCode))
		return
	}
	if err != nil {
		response.Success(c, h5DefaultCMSPage(pageCode))
		return
	}
	response.Success(c, h5CMSPagePayload(page))
}

func h5DefaultCMSPage(pageCode string) gin.H {
	title := h5Title(pageCode)
	content := "<p>Content is being prepared.</p>"
	switch pageCode {
	case "about-us":
		content = "<p>56MVP provides online entertainment services. Please play responsibly.</p>"
	case "user-agreement":
		content = "<p>By using this service, you confirm that you meet the legal age requirement and accept the platform rules.</p>"
	case "faq":
		content = "<p>Please contact support if your question is not listed.</p>"
	}
	return gin.H{
		"pageCode":    pageCode,
		"title":       title,
		"contentHTML": content,
		"contentHtml": content,
		"versionNo":   1,
		"status":      1,
	}
}

func h5CMSPagePayload(page model.GameCmsPages) gin.H {
	return gin.H{
		"id":          page.ID,
		"pageCode":    page.PageCode,
		"title":       page.Title,
		"contentHTML": page.ContentHTML,
		"contentHtml": page.ContentHTML,
		"versionNo":   page.VersionNo,
		"status":      page.Status,
		"effectiveAt": page.EffectiveAt,
	}
}
