package handler

import (
	"casino/internal/singlesite"
	"casino/internal/tenancy"
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSingleH5ScopeIgnoresHostAndClientScope(t *testing.T) {
	old, active := singlesite.Scope()
	t.Cleanup(func() { singlesite.Restore(old, active) })
	singlesite.Bind(tenancy.Scope{TenantID: 10, SiteID: 20})
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "http://another-host.invalid/api/v1/app/config", nil)
	c.Request.Header.Set("X-Tenant-ID", "999")
	c.Request.Header.Set("X-Site-ID", "888")
	c.Set("h5_tenancy_scope", tenancy.Scope{TenantID: 777, SiteID: 666})
	scope := h5TenancyScope(c)
	if scope.TenantID != 10 || scope.SiteID != 20 || scope.Platform {
		t.Fatalf("wrong scope: %+v", scope)
	}
}

func TestSingleSiteHasNoSaaSExpiryOrSuspension(t *testing.T) {
	old, active := singlesite.Scope()
	t.Cleanup(func() { singlesite.Restore(old, active) })
	singlesite.Bind(tenancy.Scope{TenantID: 10, SiteID: 20})
	expired := time.Now().Add(-time.Hour)
	tenant := h5BootstrapTenantRow{ID: 10, ExpiresAt: &expired, Enabled: false, Status: "DISABLED", MaintenanceEnabled: true}
	site := h5BootstrapSiteRow{ID: 20, Enabled: true}
	if h5NormalizeMaintenance(gin.H{}, tenant, site)["enabled"] != false {
		t.Fatal("SaaS status blocked the H5")
	}
	if h5NormalizeMaintenance(gin.H{"enabled": true}, tenant, site)["enabled"] != true {
		t.Fatal("site maintenance setting was ignored")
	}
}
