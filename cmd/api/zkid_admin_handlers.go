package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/zkid"
)

type zkidAdminContainer struct {
	svc *zkid.Service
}

func newZKIDAdminHandlers(db *sql.DB, cfg *zkid.Config) *zkidAdminContainer {
	return &zkidAdminContainer{svc: zkid.NewService(db, cfg)}
}

type recoveryCodeRequest struct {
	UserID string `json:"user_id"`
	Count  int    `json:"count,omitempty"`
}

type revokeCodeRequest struct {
	UserID string `json:"user_id"`
	CodeID string `json:"code_id"`
}

// getRecoveryCodesHandler allows admins to view recovery codes for a user (UUID-only)
func (c *zkidAdminContainer) getRecoveryCodesHandler(w http.ResponseWriter, r *http.Request) {
	// Verify admin permissions via RBAC middleware
	_, _, userRole, _, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		auth.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required", "AUTH_REQUIRED", "")
		return
	}
	if userRole != "system_admin" && userRole != "enterprise_admin" {
		auth.WriteErrorResponse(w, http.StatusForbidden, "Admin access required", "ADMIN_REQUIRED", "")
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		auth.WriteErrorResponse(w, http.StatusBadRequest, "user_id required", "MISSING_USER_ID", "")
		return
	}

	// Get count parameter (default 10)
	countStr := r.URL.Query().Get("count")
	count := 10
	if countStr != "" {
		if c, err := strconv.Atoi(countStr); err == nil && c > 0 && c <= 50 {
			count = c
		}
	}

	// Generate new recovery codes for the user
	codes, err := c.svc.GenerateRecoveryCodes(userID, count)
	if err != nil {
		auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to generate recovery codes", "GENERATION_FAILED", err.Error())
		return
	}

	// Log admin action (UUID-only, no email exposure)
	log.Printf("[ZKID_ADMIN] Admin %s generated %d recovery codes for user %s", userRole, count, userID)

	auth.WriteSuccessResponse(w, map[string]interface{}{
		"user_id": userID,
		"count":   len(codes),
		"codes":   codes,
	})
}

// revokeRecoveryCodeHandler allows admins to revoke a specific recovery code
func (c *zkidAdminContainer) revokeRecoveryCodeHandler(w http.ResponseWriter, r *http.Request) {
	// Verify admin permissions via RBAC middleware
	_, _, userRole, _, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		auth.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required", "AUTH_REQUIRED", "")
		return
	}
	if userRole != "system_admin" && userRole != "enterprise_admin" {
		auth.WriteErrorResponse(w, http.StatusForbidden, "Admin access required", "ADMIN_REQUIRED", "")
		return
	}

	var req revokeCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", err.Error())
		return
	}

	if req.UserID == "" || req.CodeID == "" {
		auth.WriteErrorResponse(w, http.StatusBadRequest, "user_id and code_id required", "MISSING_PARAMS", "")
		return
	}

	// Revoke the recovery code (mark as used)
	success, err := c.svc.RevokeRecoveryCode(req.UserID, req.CodeID)
	if err != nil {
		auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to revoke recovery code", "REVOKE_FAILED", err.Error())
		return
	}

	if !success {
		auth.WriteErrorResponse(w, http.StatusNotFound, "Recovery code not found or already used", "CODE_NOT_FOUND", "")
		return
	}

	// Log admin action (UUID-only, no email exposure)
	log.Printf("[ZKID_ADMIN] Admin %s revoked recovery code %s for user %s", userRole, req.CodeID, req.UserID)

	auth.WriteSuccessResponse(w, map[string]interface{}{
		"user_id": req.UserID,
		"code_id": req.CodeID,
		"revoked": true,
		"message": "Recovery code revoked successfully",
	})
}

// getZKIDStatsHandler provides ZKID statistics for admin monitoring
func (c *zkidAdminContainer) getZKIDStatsHandler(w http.ResponseWriter, r *http.Request) {
	// Verify admin permissions via RBAC middleware
	_, _, userRole, _, err := auth.GetUserFromContext(r.Context())
	if err != nil {
		auth.WriteErrorResponse(w, http.StatusUnauthorized, "Authentication required", "AUTH_REQUIRED", "")
		return
	}
	if userRole != "system_admin" && userRole != "enterprise_admin" {
		auth.WriteErrorResponse(w, http.StatusForbidden, "Admin access required", "ADMIN_REQUIRED", "")
		return
	}

	stats, err := c.svc.GetStats()
	if err != nil {
		auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get ZKID stats", "STATS_FAILED", err.Error())
		return
	}

	auth.WriteSuccessResponse(w, stats)
}
