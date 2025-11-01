package http

import (
	"context"
	"net/http"

	"go.uber.org/zap"
	"github.com/yourusername/gatekeeper/internal/audit"
	"github.com/yourusername/gatekeeper/internal/auth"
	"github.com/yourusername/gatekeeper/internal/log"
	"github.com/yourusername/gatekeeper/internal/policy"
)

// PolicyMiddleware evaluates access control policies for protected routes
type PolicyMiddleware struct {
	policyManager *policy.PolicyManager
	logger        *log.Logger
	auditLogger   audit.AuditLogger
}

// NewPolicyMiddleware creates a new policy middleware
func NewPolicyMiddleware(pm *policy.PolicyManager, logger *log.Logger, auditLogger audit.AuditLogger) *PolicyMiddleware {
	return &PolicyMiddleware{
		policyManager: pm,
		logger:        logger,
		auditLogger:   auditLogger,
	}
}

// Middleware returns an HTTP middleware function
func (pm *PolicyMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get claims from context (set by JWTMiddleware)
			claims, ok := r.Context().Value(ClaimsContextKey).(*auth.Claims)
			if !ok || claims == nil {
				// No claims in context, request already failed auth
				pm.logger.WithFields(
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("decision", "DENIED"),
					zap.String("reason", "no_authentication"),
				).Info("policy decision: access denied")

				// Audit log: Authorization denied - no authentication
				if pm.auditLogger != nil {
					pm.auditLogger.LogAuthzDecision(r.Context(), audit.AuditEvent{
						Result:       audit.ResultDenied,
						Method:       r.Method,
						Endpoint:     r.URL.Path,
						IPAddr:       r.RemoteAddr,
						Error:        "no_authentication",
						PolicyPath:   r.URL.Path,
						PolicyMethod: r.Method,
					})
				}

				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get policies for this route
			policies := pm.policyManager.GetPoliciesForRoute(r.URL.Path, r.Method)

			// If no policies exist for this route, allow access
			if len(policies) == 0 {
				pm.logger.WithFields(
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("address", claims.Address),
					zap.String("decision", "ALLOWED"),
					zap.String("reason", "no_policies"),
					zap.Int("policies", 0),
				).Debug("policy decision: access allowed (no policies)")
				next.ServeHTTP(w, r)
				return
			}

			// Evaluate all policies for the route
			allowed, evalErr := pm.evaluatePolicies(r.Context(), policies, claims.Address, claims)

			// Build log fields
			logFields := []zap.Field{
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.String("address", claims.Address),
				zap.Int("policies", len(policies)),
				zap.Strings("scopes", claims.Scopes),
			}

			decision := "ALLOWED"
			if !allowed {
				decision = "DENIED"
				if evalErr != nil {
					logFields = append(logFields,
						zap.Error(evalErr),
						zap.String("reason", "evaluation_error"),
					)
				} else {
					logFields = append(logFields,
						zap.String("reason", "policy_failed"),
					)
				}
			}

			logFields = append(logFields, zap.String("decision", decision))

			if evalErr != nil {
				pm.logger.WithFields(logFields...).Warn("policy evaluation error")

				// Audit log: Policy evaluation error
				if pm.auditLogger != nil {
					pm.auditLogger.LogAuthzDecision(r.Context(), audit.AuditEvent{
						Result:       audit.ResultDenied,
						UserAddr:     claims.Address,
						Method:       r.Method,
						Endpoint:     r.URL.Path,
						IPAddr:       r.RemoteAddr,
						PolicyPath:   r.URL.Path,
						PolicyMethod: r.Method,
						Error:        "evaluation_error",
						ErrorDetail:  evalErr.Error(),
						Metadata: map[string]interface{}{
							"policies_count": len(policies),
						},
					})
				}

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				pm.logger.WithFields(logFields...).Info("policy decision: access denied")

				// Audit log: Access denied by policy
				if pm.auditLogger != nil {
					pm.auditLogger.LogAuthzDecision(r.Context(), audit.AuditEvent{
						Result:       audit.ResultDenied,
						UserAddr:     claims.Address,
						Method:       r.Method,
						Endpoint:     r.URL.Path,
						IPAddr:       r.RemoteAddr,
						PolicyPath:   r.URL.Path,
						PolicyMethod: r.Method,
						Metadata: map[string]interface{}{
							"policies_count": len(policies),
							"scopes":         claims.Scopes,
						},
					})
				}

				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			// Audit log: Access granted
			if pm.auditLogger != nil {
				pm.auditLogger.LogAuthzDecision(r.Context(), audit.AuditEvent{
					Result:       audit.ResultGranted,
					UserAddr:     claims.Address,
					Method:       r.Method,
					Endpoint:     r.URL.Path,
					IPAddr:       r.RemoteAddr,
					PolicyPath:   r.URL.Path,
					PolicyMethod: r.Method,
					Metadata: map[string]interface{}{
						"policies_count": len(policies),
						"scopes":         claims.Scopes,
					},
				})
			}

			pm.logger.WithFields(logFields...).Debug("policy decision: access allowed")
			next.ServeHTTP(w, r)
		})
	}
}

// evaluatePolicies evaluates all policies for a route
func (pm *PolicyMiddleware) evaluatePolicies(ctx context.Context, policies []*policy.Policy, address string, claims *auth.Claims) (bool, error) {
	if len(policies) == 0 {
		return true, nil
	}

	// If multiple policies exist, ALL must pass (AND logic across policies)
	for _, p := range policies {
		allowed, err := p.Evaluate(ctx, address, claims)
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}

	return true, nil
}

// SetProvider sets the blockchain provider for policy evaluation
func (pm *PolicyMiddleware) SetProvider(provider policy.BlockchainProvider) {
	// Update all ERC20 and ERC721 rules in the manager
	for _, p := range pm.policyManager.GetAllPolicies() {
		for _, rule := range p.Rules {
			if erc20Rule, ok := rule.(*policy.ERC20MinBalanceRule); ok {
				erc20Rule.SetProvider(provider)
			} else if erc721Rule, ok := rule.(*policy.ERC721OwnerRule); ok {
				erc721Rule.SetProvider(provider)
			}
		}
	}
}

// SetCache sets the cache provider for policy evaluation
func (pm *PolicyMiddleware) SetCache(cache policy.CacheProvider) {
	// Update all ERC20 and ERC721 rules in the manager
	for _, p := range pm.policyManager.GetAllPolicies() {
		for _, rule := range p.Rules {
			if erc20Rule, ok := rule.(*policy.ERC20MinBalanceRule); ok {
				erc20Rule.SetCache(cache)
			} else if erc721Rule, ok := rule.(*policy.ERC721OwnerRule); ok {
				erc721Rule.SetCache(cache)
			}
		}
	}
}
