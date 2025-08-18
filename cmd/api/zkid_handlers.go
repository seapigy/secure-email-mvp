package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/zkid"
)

type zkidServiceContainer struct {
	svc *zkid.Service
}

func newZKIDHandlers(db *sql.DB, cfg *zkid.Config) *zkidServiceContainer {
	return &zkidServiceContainer{svc: zkid.NewService(db, cfg)}
}

type createMappingRequest struct {
	UserID        string  `json:"user_id"`
	Email         string  `json:"email"`
	FallbackEmail *string `json:"fallback_email,omitempty"`
}

func (c *zkidServiceContainer) createOrUpdateMappingHandler(w http.ResponseWriter, r *http.Request) {
	var req createMappingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		auth.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", err.Error())
		return
	}
	mapping, err := c.svc.CreateOrUpdateMapping(req.UserID, req.Email, req.FallbackEmail)
	if err != nil {
		auth.WriteErrorResponse(w, http.StatusBadRequest, "Failed to create mapping", "CREATE_FAILED", err.Error())
		return
	}
	auth.WriteSuccessResponse(w, map[string]interface{}{
		"id":      mapping.ID,
		"user_id": mapping.UserID,
	})
}

func (c *zkidServiceContainer) getEmailByUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		auth.WriteErrorResponse(w, http.StatusBadRequest, "user_id required", "MISSING_USER_ID", "")
		return
	}
	email, err := c.svc.GetEmailByUserID(userID)
	if err != nil {
		auth.WriteErrorResponse(w, http.StatusNotFound, "not found", "NOT_FOUND", err.Error())
		return
	}
	auth.WriteSuccessResponse(w, map[string]string{"email": email})
}
