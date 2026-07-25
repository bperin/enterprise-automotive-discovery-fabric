package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"enterprise-search/internal/auth"
)

func TestPrincipalCanAccess(t *testing.T) {
	dealerPrincipal := &auth.Principal{
		SubjectID:   "usr-dealer-1",
		Email:       "dealer@austin.com",
		Roles:       []string{"parts_manager"},
		Groups:      []string{"dealers-tx"},
		Dealerships: []string{"dealer-austin-78701"},
		Brands:      []string{"ApexMotors"},
	}

	adminPrincipal := &auth.Principal{
		SubjectID: "usr-admin-1",
		Roles:     []string{"admin"},
	}

	t.Run("Public Policy Access", func(t *testing.T) {
		publicPolicy := auth.AccessPolicy{Visibility: auth.VisibilityPublic}
		if !dealerPrincipal.CanAccess(publicPolicy) {
			t.Errorf("dealer should access public policy")
		}
		var nilPrincipal *auth.Principal
		if !nilPrincipal.CanAccess(publicPolicy) {
			t.Errorf("unauthenticated user should access public policy")
		}
	})

	t.Run("Restricted Policy Access by Dealership", func(t *testing.T) {
		policy := auth.AccessPolicy{
			Visibility: auth.VisibilityRestricted,
			TenantIDs:  []string{"dealer-austin-78701"},
		}

		if !dealerPrincipal.CanAccess(policy) {
			t.Errorf("dealer should access matching dealership policy")
		}

		otherDealer := &auth.Principal{
			SubjectID:   "usr-dealer-2",
			Dealerships: []string{"dealer-dallas-75201"},
		}
		if otherDealer.CanAccess(policy) {
			t.Errorf("other dealer should NOT access restricted Austin dealership policy")
		}
	})

	t.Run("Admin Bypass", func(t *testing.T) {
		restrictedPolicy := auth.AccessPolicy{
			Visibility: auth.VisibilityRestricted,
			SubjectIDs: []string{"some-other-user"},
		}

		if !adminPrincipal.CanAccess(restrictedPolicy) {
			t.Errorf("admin should bypass restricted access checks")
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	authenticator := auth.NewMockAuthenticator()
	handler := auth.Middleware(authenticator, "read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, ok := auth.PrincipalFromContext(r.Context())
		if !ok || p == nil {
			t.Errorf("expected principal in context")
			http.Error(w, "no principal", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	t.Run("Valid Bearer Token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer dev-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}
