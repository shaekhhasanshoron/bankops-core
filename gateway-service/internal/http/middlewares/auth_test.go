package middleware

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestPermissionLogic tests rbac permission logic
func TestPermissionLogic(t *testing.T) {
	permissions := getRBACPermissions()

	testCases := []struct {
		path     string
		method   string
		role     string
		expected bool
		desc     string
	}{
		// Employee endpoints
		{"/api/v1/employee", "POST", "admin", true, "Admin can create employee"},
		{"/api/v1/employee", "POST", "editor", false, "Editor cannot create employee"},
		{"/api/v1/employee", "POST", "viewer", false, "Viewer cannot create employee"},

		{"/api/v1/employee", "GET", "admin", true, "Admin can list employees"},
		{"/api/v1/employee", "GET", "editor", false, "Editor cannot list employees"},
		{"/api/v1/employee", "GET", "viewer", false, "Viewer cannot list employees"},

		{"/api/v1/employee/:username", "DELETE", "admin", true, "Admin can delete employee"},
		{"/api/v1/employee/:username", "DELETE", "editor", false, "Editor cannot delete employee"},
		{"/api/v1/employee/:username", "DELETE", "viewer", false, "Viewer cannot delete employee"},

		// Customer endpoints
		{"/api/v1/customer", "POST", "admin", true, "Admin can create customer"},
		{"/api/v1/customer", "POST", "editor", true, "Editor can create customer"},
		{"/api/v1/customer", "POST", "viewer", false, "Viewer cannot create customer"},

		{"/api/v1/customer", "GET", "admin", true, "Admin can list customers"},
		{"/api/v1/customer", "GET", "editor", true, "Editor can list customers"},
		{"/api/v1/customer", "GET", "viewer", true, "Viewer can list customers"},

		// Account endpoints
		{"/api/v1/account", "POST", "admin", true, "Admin can create account"},
		{"/api/v1/account", "POST", "editor", true, "Editor can create account"},
		{"/api/v1/account", "POST", "viewer", false, "Viewer cannot create account"},

		{"/api/v1/account/:id/balance", "GET", "admin", true, "Admin can view balance"},
		{"/api/v1/account/:id/balance", "GET", "editor", true, "Editor can view balance"},
		{"/api/v1/account/:id/balance", "GET", "viewer", true, "Viewer can view balance"},

		// Transaction endpoints
		{"/api/v1/transaction/init", "POST", "admin", true, "Admin can initiate transaction"},
		{"/api/v1/transaction/init", "POST", "editor", true, "Editor can initiate transaction"},
		{"/api/v1/transaction/init", "POST", "viewer", false, "Viewer cannot initiate transaction"},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			allowed := checkPermission(permissions, tc.path, tc.method, tc.role)
			assert.Equal(t, tc.expected, allowed, tc.desc)
		})
	}
}

func checkPermission(permissions map[string]map[string]map[string]bool, path, method, role string) bool {
	if routePerms, ok := permissions[path]; ok {
		if methodPerms, ok := routePerms[method]; ok {
			return methodPerms[role]
		}
	}
	return false
}
