// Package auth provides OIDC authentication, principal resolution, and fine-grained access control policies.
//
// ARCHITECTURAL DESIGN & TRANSFORMATION:
//
// BEFORE (Legacy Enterprise Integration):
//   Source Systems -> Third-Party Connector Vendor -> Third-Party Integration Platform -> Search Platform -> Applications
//   - Problem: Duplicated authentication, fragmented authorization, expensive proprietary connector licenses, and brittle ETL synchronizations.
//
// GCP-NATIVE PLATFORM DESIGN:
//   Source Systems -> Small In-House Source Adapters -> Pub/Sub + Canonical Model -> Search Indexes & Operational Stores -> Stable Discovery API -> Applications & Agents
//   - Layer 1: Unified Trust Layer (OIDC / Google Cloud IAM / Workload Identity)
//   - Layer 2: Integration Layer (Small, reusable, internally owned Go source adapters)
//   - Layer 3: Reusable Data Layer (Canonical domain models and Pub/Sub event streams)
//   - Layer 4: Retrieval Layer (Agent Search, Agent Retrieval, BigQuery SQL)
//   - Layer 5: Reasoning & Orchestration Layer (Grounded agentic graph workflows)
//
// SECURITY BOUNDARY & ACCESS CONTROL:
// Enterprise search requires more than simple user login. Search results must preserve source permissions.
// The search engine MUST filter candidate records using the user's Principal and the record's AccessPolicy
// BEFORE returning results to clients or LLMs. We cannot retrieve restricted records and rely on an LLM not to cite them.
package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"enterprise-search/internal/http/response"
)

// Visibility level for indexed records.
type Visibility string

const (
	VisibilityPublic     Visibility = "public"
	VisibilityInternal   Visibility = "internal"
	VisibilityRestricted Visibility = "restricted"
)

// AccessPolicy defines fine-grained record-level security constraints in DDD.
// Every indexed product, document, fitment assertion, or claim carries an AccessPolicy.
type AccessPolicy struct {
	Visibility Visibility `json:"visibility"`          // "public", "internal", "restricted"
	Groups     []string   `json:"groups,omitempty"`     // Allowed user groups (e.g. "dealers-tx", "supplier-admins")
	Roles      []string   `json:"roles,omitempty"`      // Allowed roles (e.g. "parts-manager", "operator")
	TenantIDs  []string   `json:"tenant_ids,omitempty"` // Allowed dealership / tenant IDs
	SubjectIDs []string   `json:"subject_ids,omitempty"`// Allowed explicit user subject IDs
}

// Principal represents the authenticated security context at the platform boundary.
// It is resolved from OIDC ID tokens or OAuth 2.0 access token inspection.
type Principal struct {
	SubjectID   string   `json:"subject_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Groups      []string `json:"groups"`
	Dealerships []string `json:"dealerships"`
	Brands      []string `json:"brands"`
	Scopes      []string `json:"scopes"`
}

// CanAccess evaluates whether the principal holds sufficient authorization to view a record under the target AccessPolicy.
// Pre-filtering in the search engine using this method ensures restricted commercial terms or supplier pricing
// are never exposed to unauthorized users or LLM synthesis steps.
func (p *Principal) CanAccess(policy AccessPolicy) bool {
	if p == nil {
		return policy.Visibility == VisibilityPublic
	}

	// Admin / superuser bypass
	if p.HasRole("admin") || p.HasRole("superuser") {
		return true
	}

	// Public visibility
	if policy.Visibility == VisibilityPublic || policy.Visibility == "" {
		return true
	}

	// Internal visibility requires authenticated principal
	if policy.Visibility == VisibilityInternal {
		return p.SubjectID != ""
	}

	// Restricted visibility checks explicit permissions
	// 1. Direct Subject ID match
	for _, sub := range policy.SubjectIDs {
		if sub == p.SubjectID {
			return true
		}
	}

	// 2. Tenant / Dealership match
	for _, tenant := range policy.TenantIDs {
		for _, dealership := range p.Dealerships {
			if tenant == dealership {
				return true
			}
		}
	}

	// 3. Group match
	for _, reqGroup := range policy.Groups {
		for _, userGroup := range p.Groups {
			if reqGroup == userGroup {
				return true
			}
		}
	}

	// 4. Role match
	for _, reqRole := range policy.Roles {
		for _, userRole := range p.Roles {
			if reqRole == userRole {
				return true
			}
		}
	}

	return false
}

// HasRole checks if the principal holds a specific role.
func (p *Principal) HasRole(role string) bool {
	for _, r := range p.Roles {
		if r == role || r == "admin" {
			return true
		}
	}
	return false
}

// HasScope checks if the principal holds a required OAuth2 scope.
func (p *Principal) HasScope(required string) bool {
	if required == "" {
		return true
	}
	for _, s := range p.Scopes {
		if s == required || s == "admin" || s == "*" {
			return true
		}
	}
	return false
}

type contextKey string

const principalKey contextKey = "principal"

// WithPrincipal attaches a Principal to the context.
func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// PrincipalFromContext extracts the Principal from context if present.
func PrincipalFromContext(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalKey).(*Principal)
	return p, ok
}

// Authenticator validates an OAuth2/OIDC Bearer token and returns a Principal.
type Authenticator interface {
	AuthenticateToken(ctx context.Context, token string) (*Principal, error)
}

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrMissingAuth  = errors.New("missing authorization header")
)

// MockAuthenticator is a dev/test implementation of Authenticator.
type MockAuthenticator struct {
	ValidTokens map[string]*Principal
}

func NewMockAuthenticator() *MockAuthenticator {
	return &MockAuthenticator{
		ValidTokens: map[string]*Principal{
			"dev-token": {
				SubjectID:   "usr-dev-123",
				Email:       "dev@automotive-fabric.io",
				Roles:       []string{"developer", "operator"},
				Groups:      []string{"dealers-tx", "engineering"},
				Dealerships: []string{"dealer-austin-78701"},
				Brands:      []string{"ApexMotors", "NovaMotors"},
				Scopes:      []string{"read", "write", "search"},
			},
			"dealer-token": {
				SubjectID:   "usr-dealer-456",
				Email:       "manager@austin-apexmotors.com",
				Roles:       []string{"parts_manager"},
				Groups:      []string{"dealers-tx"},
				Dealerships: []string{"dealer-austin-78701"},
				Brands:      []string{"ApexMotors"},
				Scopes:      []string{"read", "search"},
			},
			"admin-token": {
				SubjectID:   "usr-admin-999",
				Email:       "admin@automotive-fabric.io",
				Roles:       []string{"admin"},
				Groups:      []string{"admins"},
				Dealerships: []string{"*"},
				Brands:      []string{"*"},
				Scopes:      []string{"admin", "read", "write", "search"},
			},
		},
	}
}

func (m *MockAuthenticator) AuthenticateToken(ctx context.Context, token string) (*Principal, error) {
	if p, ok := m.ValidTokens[token]; ok {
		return p, nil
	}
	if strings.HasPrefix(token, "test-bearer-") {
		return &Principal{
			SubjectID:   strings.TrimPrefix(token, "test-bearer-"),
			Email:       "testuser@automotive-fabric.io",
			Roles:       []string{"user"},
			Groups:      []string{"default"},
			Dealerships: []string{"dealer-austin-78701"},
			Brands:      []string{"ApexMotors"},
			Scopes:      []string{"read", "search"},
		}, nil
	}
	return nil, ErrInvalidToken
}

// Middleware creates an HTTP handler wrapper enforcing OIDC Bearer authentication and scope checks.
func Middleware(auth Authenticator, requiredScopes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "unauthorized", "Missing Authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Error(w, http.StatusUnauthorized, "invalid_request", "Invalid Authorization header format; expected Bearer token")
				return
			}

			token := parts[1]
			principal, err := auth.AuthenticateToken(r.Context(), token)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid_token", "Authentication failed: "+err.Error())
				return
			}

			for _, requiredScope := range requiredScopes {
				if !principal.HasScope(requiredScope) {
					response.Error(w, http.StatusForbidden, "insufficient_scope", "Forbidden: insufficient OAuth2 scope "+requiredScope)
					return
				}
			}

			ctx := WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
