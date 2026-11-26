package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-dev-frame/sponge/pkg/gin/response"

	"casino/internal/adminauth"
	"casino/internal/database"
	"casino/internal/singlesite"
	"casino/internal/storage"
	"casino/internal/systemadmin"
)

type AdminSystemHandler struct{ service *systemadmin.Service }

func NewAdminSystemHandler() *AdminSystemHandler {
	service := systemadmin.New(database.GetDB())
	if err := service.EnsureMenuCatalog(); err != nil {
		log.Printf("initialize admin menu catalog: %v", err)
	}
	systemadmin.StartScheduler(service)
	return &AdminSystemHandler{service: service}
}

func (h *AdminSystemHandler) Navigation(c *gin.Context) {
	actor, ok := adminauth.FromContext(c)
	if !ok {
		systemError(c, systemadmin.NewError("AUTHENTICATION_REQUIRED", "authentication required"))
		return
	}
	data, err := h.service.Navigation(systemActor(c), actor.Can("*"))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) Login(c *gin.Context) {
	var input systemadmin.LoginInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.AdminLogin(input, c.ClientIP(), c.GetHeader("User-Agent"), systemCorrelationID(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ConfirmMFASetup(c *gin.Context) {
	var input systemadmin.MFAChallengeInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.ConfirmMFASetup(input, systemCorrelationID(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) VerifyMFA(c *gin.Context) {
	var input systemadmin.MFAChallengeInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.VerifyMFA(input, systemCorrelationID(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) CurrentAdmin(c *gin.Context) {
	actor, ok := adminauth.FromContext(c)
	if !ok {
		systemError(c, systemadmin.NewError("AUTHENTICATION_REQUIRED", "authentication required"))
		return
	}
	systemWrite(c, actor, nil)
}

func (h *AdminSystemHandler) H5Markets(c *gin.Context) {
	data, err := h.service.ListCountries(systemadmin.Query{Page: 1, PageSize: 200, Values: map[string]string{"status": "1"}})
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) SaveH5Settings(c *gin.Context) {
	var input systemadmin.SiteInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.SaveSingleSiteSettings(input, systemActor(c)))
}

func (h *AdminSystemHandler) BrandOverview(c *gin.Context) {
	data, err := h.service.BrandOverview(systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) Logout(c *gin.Context) {
	token := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	systemWrite(c, nil, h.service.Logout(token))
}

func (h *AdminSystemHandler) AccountProfile(c *gin.Context) {
	data, err := h.service.AccountProfile(systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) UpdateAccountProfile(c *gin.Context) {
	var input systemadmin.AccountProfileInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateAccountProfile(input, systemActor(c)))
}

func (h *AdminSystemHandler) ChangeAccountPassword(c *gin.Context) {
	var input systemadmin.ChangePasswordInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.ChangeAccountPassword(input, systemActor(c)))
}

func (h *AdminSystemHandler) AccountMFAStatus(c *gin.Context) {
	data, err := h.service.AccountMFAStatus(systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) StartAccountMFASetup(c *gin.Context) {
	var input systemadmin.AccountMFAInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.StartAccountMFASetup(input, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ConfirmAccountMFASetup(c *gin.Context) {
	var input systemadmin.MFAChallengeInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.ConfirmAccountMFASetup(input, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) DisableAccountMFA(c *gin.Context) {
	var input systemadmin.AccountMFAInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DisableAccountMFA(input, systemActor(c)))
}

func (h *AdminSystemHandler) RegenerateAccountRecoveryCodes(c *gin.Context) {
	var input systemadmin.AccountMFAInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.RegenerateAccountRecoveryCodes(input, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) AccountSessions(c *gin.Context) {
	data, err := h.service.AccountSessions(systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) RevokeAccountSession(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	systemWrite(c, nil, h.service.RevokeAccountSession(id, systemActor(c)))
}

func (h *AdminSystemHandler) RevokeOtherAccountSessions(c *gin.Context) {
	systemWrite(c, nil, h.service.RevokeOtherAccountSessions(systemActor(c)))
}

func (h *AdminSystemHandler) AccountLoginRecords(c *gin.Context) {
	data, err := h.service.AccountLoginRecords(systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ListAdmins(c *gin.Context) {
	data, err := h.service.ListAdmins(systemQuery(c, adminSystemKeys), systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetAdmin(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.AdminDetail(id, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreateAdmin(c *gin.Context) {
	var input systemadmin.AdminInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.CreateAdmin(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdateAdmin(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.AdminInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateAdmin(id, input, systemActor(c)))
}
func (h *AdminSystemHandler) AdminStatus(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input statusReason
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.AdminStatus(id, input.Enabled, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) ResetPassword(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	value, err := h.service.ResetPassword(id, input.Reason, systemActor(c))
	systemWrite(c, map[string]any{"temporaryPassword": value}, err)
}
func (h *AdminSystemHandler) ResetMFA(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.ResetMFA(id, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) ForceLogout(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.ForceLogout(id, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) AssignAdminRoles(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.IDsInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.AssignAdminRoles(id, input, systemActor(c)))
}
func (h *AdminSystemHandler) MoveAdminDepartment(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input struct {
		DepartmentID *uint64 `json:"departmentID"`
		Reason       string  `json:"reason"`
	}
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.MoveAdminDepartment(id, input.DepartmentID, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) ListRoles(c *gin.Context) {
	data, err := h.service.ListRoles(systemQuery(c, roleSystemKeys), systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetRole(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.RoleDetail(id, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreateRole(c *gin.Context) {
	var input systemadmin.RoleInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.CreateRole(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdateRole(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.RoleInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateRole(id, input, systemActor(c)))
}
func (h *AdminSystemHandler) RoleStatus(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input statusReason
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.RoleStatus(id, input.Enabled, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) AssignRolePermissions(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.IDsInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.AssignRolePermissions(id, input, systemActor(c)))
}
func (h *AdminSystemHandler) AssignRoleMenus(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.IDsInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.AssignRoleMenus(id, input, systemActor(c)))
}
func (h *AdminSystemHandler) CopyRole(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.RoleInput
	if !systemBind(c, &input) {
		return
	}
	newID, err := h.service.CopyRole(id, input, systemActor(c))
	systemWrite(c, map[string]any{"id": newID}, err)
}
func (h *AdminSystemHandler) DeleteRole(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DeleteRole(id, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) ListPermissions(c *gin.Context) {
	data, err := h.service.ListPermissions(systemQuery(c, permissionSystemKeys))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetPermission(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.PermissionDetail(id)
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreatePermission(c *gin.Context) {
	var input systemadmin.PermissionInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.CreatePermission(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdatePermission(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.PermissionInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdatePermission(id, input, systemActor(c)))
}
func (h *AdminSystemHandler) PermissionStatus(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input statusReason
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.PermissionStatus(id, input.Enabled, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) DeletePermission(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DeletePermission(id, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) ListDepartments(c *gin.Context) {
	data, err := h.service.ListDepartments(systemQuery(c, departmentSystemKeys))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetDepartment(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.DepartmentDetail(id)
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreateDepartment(c *gin.Context) {
	var input systemadmin.DepartmentInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.CreateDepartment(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdateDepartment(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input departmentCommand
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateDepartment(id, input.DepartmentInput, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) DepartmentStatus(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input statusReason
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DepartmentStatus(id, input.Enabled, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) DeleteDepartment(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DeleteDepartment(id, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) ListMenus(c *gin.Context) {
	data, err := h.service.ListMenus(systemQuery(c, menuSystemKeys))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetMenu(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.MenuDetail(id)
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreateMenu(c *gin.Context) {
	var input systemadmin.MenuInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.CreateMenu(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdateMenu(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input menuCommand
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateMenu(id, input.MenuInput, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) MenuStatus(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input statusReason
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.MenuStatus(id, input.Enabled, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) CopyMenu(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.MenuInput
	if !systemBind(c, &input) {
		return
	}
	newID, err := h.service.CopyMenu(id, input, systemActor(c))
	systemWrite(c, map[string]any{"id": newID}, err)
}
func (h *AdminSystemHandler) DeleteMenu(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DeleteMenu(id, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) ListConfigs(c *gin.Context) {
	query := systemQuery(c, configSystemKeys)
	query.Values["adminScope"] = "PLATFORM"
	data, err := h.service.ListConfigs(query)
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) PlatformBranding(c *gin.Context) {
	data, err := h.service.PlatformBranding(c.Request.Context())
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) UpdateConfig(c *gin.Context) {
	var input systemadmin.ConfigInput
	if !systemBind(c, &input) {
		return
	}
	actor, _ := adminauth.FromContext(c)
	systemWrite(c, nil, h.service.UpdateConfig(c.Param("key"), input, actor.Can(adminauth.PermissionSystemConfigSecretUpdate), systemActor(c)))
}

func (h *AdminSystemHandler) GetSite(c *gin.Context) {
	id, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	data, err := h.service.SiteDetail(id, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) PreviewSiteConfig(c *gin.Context) {
	id, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	data, err := h.service.PreviewSiteConfig(id, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) PublishSiteConfig(c *gin.Context) {
	id, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.PublishSiteConfigWithKey(id, input.Reason, systemIdempotencyKey(c, ""), systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ListSiteAnnouncements(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	data, err := h.service.ListSiteAnnouncements(siteID, systemQuery(c, []string{"keyword", "publishStatus"}), systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetSiteAnnouncement(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.GetSiteAnnouncement(siteID, id, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreateSiteAnnouncement(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	var input systemadmin.AnnouncementInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SaveSiteAnnouncement(siteID, 0, input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdateSiteAnnouncement(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.AnnouncementInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SaveSiteAnnouncement(siteID, id, input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) CopySiteAnnouncement(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	id, ok := systemID(c)
	if !ok {
		return
	}
	newID, err := h.service.CopySiteAnnouncement(siteID, id, systemActor(c))
	systemWrite(c, map[string]any{"id": newID}, err)
}
func (h *AdminSystemHandler) DeleteSiteAnnouncement(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	id, ok := systemID(c)
	if !ok {
		return
	}
	systemWrite(c, nil, h.service.DeleteSiteAnnouncement(siteID, id, systemActor(c)))
}
func (h *AdminSystemHandler) ReorderSiteAnnouncements(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	var input systemadmin.AnnouncementReorderInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.ReorderSiteAnnouncements(siteID, input, systemActor(c)))
}

func (h *AdminSystemHandler) ListPopupComponents(c *gin.Context) {
	data, err := h.service.ListPopupComponents(systemQuery(c, []string{"keyword", "category", "enabled"}))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) AvailablePopupComponents(c *gin.Context) {
	data, err := h.service.AvailablePopupComponents()
	systemWrite(c, map[string]any{"list": data}, err)
}
func (h *AdminSystemHandler) CreatePopupComponent(c *gin.Context) {
	var input systemadmin.PopupComponentInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SavePopupComponent(0, input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdatePopupComponent(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.PopupComponentInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SavePopupComponent(id, input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) DeletePopupComponent(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DeletePopupComponent(id, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) ListSitePopups(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	data, err := h.service.ListSitePopups(siteID, systemQuery(c, []string{"keyword", "publishStatus"}), systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetSitePopup(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.GetSitePopup(siteID, id, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreateSitePopup(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	var input systemadmin.PopupInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SaveSitePopup(siteID, 0, input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdateSitePopup(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.PopupInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.SaveSitePopup(siteID, id, input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) CopySitePopup(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	id, ok := systemID(c)
	if !ok {
		return
	}
	newID, err := h.service.CopySitePopup(siteID, id, systemActor(c))
	systemWrite(c, map[string]any{"id": newID}, err)
}
func (h *AdminSystemHandler) DeleteSitePopup(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	id, ok := systemID(c)
	if !ok {
		return
	}
	systemWrite(c, nil, h.service.DeleteSitePopup(siteID, id, systemActor(c)))
}
func (h *AdminSystemHandler) ReorderSitePopups(c *gin.Context) {
	siteID, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	var input systemadmin.PopupReorderInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.ReorderSitePopups(siteID, input, systemActor(c)))
}

func (h *AdminSystemHandler) UpdateSite(c *gin.Context) {
	id, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	var input systemadmin.SiteInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateSite(id, input, systemActor(c)))
}
func (h *AdminSystemHandler) SiteStatus(c *gin.Context) {
	id, ok := systemNamedID(c, "siteId")
	if !ok {
		return
	}
	var input statusReason
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.SiteStatus(id, input.Enabled, input.Reason, systemActor(c)))
}

func (h *AdminSystemHandler) ListCountries(c *gin.Context) {
	data, err := h.service.ListCountries(systemQuery(c, countrySystemKeys))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetCountry(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.CountryDetail(id)
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreateCountry(c *gin.Context) {
	var input systemadmin.CountryInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.CreateCountry(input, systemActor(c))
	systemWrite(c, gin.H{"id": id}, err)
}
func (h *AdminSystemHandler) UpdateCountry(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.CountryInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateCountry(id, input, systemActor(c)))
}
func (h *AdminSystemHandler) DeleteCountry(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DeleteCountry(id, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) CountryStatus(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input statusReason
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.CountryStatus(id, input.Enabled, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) SetDefaultCountry(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.SetDefaultCountry(id, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) UploadCountryAsset(c *gin.Context) {
	kind := strings.TrimSpace(c.DefaultPostForm("kind", "image"))
	allowed := map[string]map[string]bool{
		"image":  {".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".avif": true, ".gif": true, ".svg": true, ".ico": true},
		"lottie": {".json": true},
	}
	if _, ok := allowed[kind]; !ok {
		systemError(c, systemadmin.NewError("INVALID_ASSET_KIND", "invalid asset kind"))
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		systemError(c, systemadmin.NewError("FILE_REQUIRED", "file is required"))
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	limit := int64(10 << 20)
	if kind == "lottie" {
		limit = 5 << 20
	}
	if !allowed[kind][ext] || file.Size <= 0 || file.Size > limit {
		systemError(c, systemadmin.NewError("INVALID_ASSET", "unsupported file type or size"))
		return
	}
	name, err := storage.RandomName(ext)
	if err != nil {
		systemError(c, err)
		return
	}
	asset, err := storage.UploadFileHeader(c.Request.Context(), "admin/uploads/system/"+kind+"/"+name, file)
	if err != nil {
		systemError(c, err)
		return
	}
	if err = h.service.RecordCountryAssetUpload(kind, asset.Key, systemActor(c)); err != nil {
		systemError(c, err)
		return
	}
	response.Success(c, asset)
}
func (h *AdminSystemHandler) UploadSiteAPK(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		systemError(c, systemadmin.NewError("FILE_REQUIRED", "file is required"))
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".apk" || file.Size <= 0 || file.Size > 300<<20 {
		systemError(c, systemadmin.NewError("INVALID_APK", "APK must be a non-empty .apk file no larger than 300 MB"))
		return
	}
	name, err := storage.RandomName(ext)
	if err != nil {
		systemError(c, err)
		return
	}
	asset, err := storage.UploadFileHeader(c.Request.Context(), "admin/uploads/apk/"+name, file)
	if err != nil {
		systemError(c, err)
		return
	}
	response.Success(c, map[string]any{"url": asset.URL, "key": asset.Key, "name": file.Filename, "size": asset.Size, "mimeType": asset.MimeType})
}
func (h *AdminSystemHandler) UploadPlatformAsset(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		systemError(c, systemadmin.NewError("FILE_REQUIRED", "file is required"))
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true, ".avif": true, ".gif": true, ".svg": true}
	if !allowed[ext] || file.Size <= 0 || file.Size > 10<<20 {
		systemError(c, systemadmin.NewError("INVALID_ASSET", "unsupported image type or size"))
		return
	}
	name, err := storage.RandomName(ext)
	if err != nil {
		systemError(c, err)
		return
	}
	asset, err := storage.UploadFileHeader(c.Request.Context(), "admin/uploads/platform/"+name, file)
	if err != nil {
		systemError(c, err)
		return
	}
	response.Success(c, asset)
}
func (h *AdminSystemHandler) AppConfig(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	db := database.GetDB()
	scope := h5ContentScope(c)
	site := h5BootstrapSite(db, scope.SiteID)
	countryCode := h5FirstNonEmpty(site.CountryCode, c.Query("country"))
	data, err := h.service.PublicAppConfig(countryCode, c.Query("language"))
	if err == nil && site.ID > 0 {
		country := h5BootstrapCountry(db, countryCode)
		brand := h5NormalizeBranding(h5BootstrapJSON(site.BrandConfig), site, country)
		branding, _ := data["branding"].(map[string]any)
		if branding == nil {
			branding = map[string]any{}
		}
		for source, target := range map[string]string{
			"logo": "logoUrl", "favicon": "faviconUrl", "loadingLogo": "loadingLogoUrl", "splash": "loadingBackgroundUrl",
		} {
			if value := h5AnyString(brand, source, ""); value != "" {
				branding[target] = value
			}
		}
		theme := h5NormalizeTheme(h5BootstrapJSON(site.ThemeConfig))
		themeValues, _ := branding["theme"].(map[string]any)
		if themeValues == nil {
			themeValues = map[string]any{}
		}
		for source, target := range map[string]string{
			"primary": "primaryColor", "secondary": "secondaryColor", "button": "buttonColor", "background": "backgroundColor", "text": "textColor",
		} {
			if value := h5AnyString(theme, source, ""); value != "" {
				themeValues[target] = value
			}
		}
		branding["appName"] = h5FirstNonEmpty(h5AnyString(brand, "appName", ""), site.SiteName)
		branding["siteName"] = site.SiteName
		data["seo"] = h5BootstrapJSON(site.SEOConfig)
		if languages, ok := data["language"].(map[string]any); ok && site.DefaultLanguage != "" {
			languages["default"] = site.DefaultLanguage
			if c.Query("language") == "" {
				languages["current"] = site.DefaultLanguage
			}
		}
		branding["theme"] = themeValues
		data["branding"] = branding
		data["maintenance"] = h5EffectiveMaintenance(c, db, scope)
		data["siteId"] = scope.SiteID
		data["siteId"] = scope.SiteID
		data["resourceVersion"] = strings.TrimSpace(stringValue(data["resourceVersion"])) + "-site-" + strconv.FormatUint(site.ConfigVersion, 10)
		if countryData, ok := data["country"].(map[string]any); ok {
			countryData["timezone"] = h5FirstNonEmpty(site.Timezone, country.Timezone)
		}
	}
	systemWrite(c, data, err)
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func (h *AdminSystemHandler) ListAuditLogs(c *gin.Context) {
	data, err := h.service.ListAuditLogs(systemQuery(c, auditSystemKeys))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetAuditLog(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.AuditLogDetail(id)
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) ListLoginLogs(c *gin.Context) {
	data, err := h.service.ListLoginLogs(systemQuery(c, loginSystemKeys))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetLoginLog(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.LoginLogDetail(id)
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ListTasks(c *gin.Context) {
	data, err := h.service.ListTasks(systemQuery(c, taskSystemKeys))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetTask(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.TaskDetail(id)
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) CreateTask(c *gin.Context) {
	var input systemadmin.TaskInput
	if !systemBind(c, &input) {
		return
	}
	id, err := h.service.CreateTask(input, systemActor(c))
	systemWrite(c, map[string]any{"id": id}, err)
}
func (h *AdminSystemHandler) UpdateTask(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input taskCommand
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.UpdateTask(id, input.TaskInput, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) TaskStatus(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input statusReason
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.TaskStatus(id, input.Enabled, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) RunTask(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.RunTaskNow(id, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) DeleteTask(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.ReasonInput
	if !systemBind(c, &input) {
		return
	}
	systemWrite(c, nil, h.service.DeleteTask(id, input.Reason, systemActor(c)))
}
func (h *AdminSystemHandler) TaskExecutions(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.ListTaskExecutions(id, systemQuery(c, []string{"status", "startedFrom", "startedTo"}))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) ListEvents(c *gin.Context) {
	data, err := h.service.ListEvents(systemQuery(c, eventSystemKeys), systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) GetEvent(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	data, err := h.service.EventDetail(id, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) DispatchEvents(c *gin.Context) {
	var input systemadmin.EventDispatchInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.DispatchEvents(input, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) ReplayEvent(c *gin.Context) {
	id, ok := systemID(c)
	if !ok {
		return
	}
	var input systemadmin.EventReplayInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.ReplayEvent(id, input, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) ReplayEventRange(c *gin.Context) {
	var input systemadmin.EventReplayRangeInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.ReplayEventRange(input, systemActor(c))
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) ReplayConsumer(c *gin.Context) {
	var input systemadmin.EventReplayConsumerInput
	if !systemBind(c, &input) {
		return
	}
	data, err := h.service.ReplayConsumer(input, systemActor(c))
	systemWrite(c, data, err)
}

func (h *AdminSystemHandler) Monitoring(c *gin.Context) {
	data, err := h.service.Monitoring()
	systemWrite(c, data, err)
}
func (h *AdminSystemHandler) MonitoringService(c *gin.Context) {
	data, err := h.service.MonitoringService(c.Param("service"))
	systemWrite(c, data, err)
}

type statusReason struct {
	Enabled bool   `json:"enabled"`
	Reason  string `json:"reason"`
}
type departmentCommand struct {
	systemadmin.DepartmentInput
	Reason string `json:"reason"`
}
type menuCommand struct {
	systemadmin.MenuInput
	Reason string `json:"reason"`
}
type taskCommand struct {
	systemadmin.TaskInput
	Reason string `json:"reason"`
}

func systemActor(c *gin.Context) systemadmin.Actor {
	actor, _ := adminauth.FromContext(c)
	if singlesite.ControlResource(c.Request.URL.Path) {
		actor.Platform = true
		actor.SiteID = 0
	}
	return systemadmin.Actor{ID: actor.ID, SessionID: actor.SessionID, Username: actor.Username, SiteID: actor.SiteID, Platform: actor.Platform, IP: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CorrelationID: systemCorrelationID(c), Sensitive: actor.Can(adminauth.PermissionSensitiveView)}
}
func systemCorrelationID(c *gin.Context) string {
	value := strings.TrimSpace(c.GetHeader("X-Correlation-ID"))
	if value == "" {
		value = strings.TrimSpace(c.GetHeader("X-Request-ID"))
	}
	return value
}
func systemID(c *gin.Context) (uint64, bool) {
	return systemNamedID(c, "id")
}
func systemNamedID(c *gin.Context, name string) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || id == 0 {
		systemError(c, systemadmin.NewError("INVALID_REQUEST", "invalid identifier"))
		return 0, false
	}
	return id, true
}
func systemBind(c *gin.Context, out any) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		systemError(c, systemadmin.NewError("INVALID_REQUEST", "invalid request body"))
		return false
	}
	return true
}
func systemQuery(c *gin.Context, keys []string) systemadmin.Query {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	values := map[string]string{}
	for _, key := range keys {
		if value := strings.TrimSpace(c.Query(key)); value != "" {
			values[key] = value
		}
	}
	return systemadmin.Query{Page: page, PageSize: size, Values: values}
}
func systemIdempotencyKey(c *gin.Context, fallback string) string {
	if value := strings.TrimSpace(c.GetHeader("Idempotency-Key")); value != "" {
		return value
	}
	return strings.TrimSpace(fallback)
}
func systemWrite(c *gin.Context, data any, err error) {
	if err != nil {
		systemError(c, err)
		return
	}
	response.Success(c, data)
}
func systemError(c *gin.Context, err error) {
	var serviceErr *systemadmin.Error
	if errors.As(err, &serviceErr) {
		status := http.StatusBadRequest
		if serviceErr.Code == "PERMISSION_DENIED" {
			status = http.StatusForbidden
		}
		c.JSON(http.StatusOK, gin.H{"code": status, "msg": serviceErr.Message, "data": gin.H{"errorCode": serviceErr.Code}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusInternalServerError, "msg": "internal error", "data": gin.H{"errorCode": "INTERNAL_ERROR"}})
}

var adminSystemKeys = []string{"administratorID", "siteID", "site", "username", "displayName", "department", "role", "status", "mobile", "email", "createdFrom", "createdTo", "lastLoginFrom", "lastLoginTo"}
var roleSystemKeys = []string{"code", "name", "status", "dataScope", "createdFrom", "createdTo"}
var permissionSystemKeys = []string{"code", "name", "module", "permissionType", "status"}
var departmentSystemKeys = []string{"code", "name", "status", "parentID"}
var menuSystemKeys = []string{"code", "name", "menuType", "status", "permissionCode"}
var configSystemKeys = []string{"key", "name", "category", "valueType", "editable", "sensitive"}
var siteSystemKeys = []string{"siteCode", "siteName", "primaryDomain", "adminDomain", "h5Domain", "countryCode", "currencyCode", "status", "lifecycleStatus", "createdFrom", "createdTo"}
var countrySystemKeys = []string{"countryCode", "countryName", "currencyCode", "status", "isDefault", "createdFrom", "createdTo"}
var auditSystemKeys = []string{"siteID", "operator", "module", "action", "targetType", "result", "ip", "correlationID", "createdFrom", "createdTo"}
var loginSystemKeys = []string{"administrator", "loginResult", "ip", "country", "device", "browser", "createdFrom", "createdTo"}
var taskSystemKeys = []string{"taskCode", "taskName", "taskType", "status", "lastRunResult", "nextRunFrom", "nextRunTo"}
var eventSystemKeys = []string{"eventType", "status", "aggregateType", "aggregateID"}
