package api

import (
	"database/sql"
	"net/http"
	"time"

	"agentsocial/internal/config"
	"agentsocial/internal/core"

	"github.com/gin-gonic/gin"
)

// CreateReportRequest is the body for POST /api/v1/reports.
type CreateReportRequest struct {
	TargetAgentID string `json:"target_agent_id" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
}

// CreateReport handles POST /api/v1/reports.
func CreateReport(database *sql.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent, ok := getAgent(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		var req CreateReportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request body: " + err.Error(),
			})
			return
		}

		// Prevent self-reporting.
		if req.TargetAgentID == agent.ID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_report",
				"message": "Cannot report yourself",
			})
			return
		}

		// Verify the target agent exists.
		var targetExists int
		err := database.QueryRow("SELECT COUNT(*) FROM agents WHERE id = ?", req.TargetAgentID).Scan(&targetExists)
		if err != nil || targetExists == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "agent_not_found",
				"message": "Target agent not found",
			})
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)
		reportID := core.GenerateMD5(agent.ID, req.TargetAgentID, now)

		// Insert report.
		_, err = database.Exec(
			"INSERT INTO reports (id, reporter_id, target_id, reason, created_at) VALUES (?, ?, ?, ?, ?)",
			reportID, agent.ID, req.TargetAgentID, req.Reason, now,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to create report",
			})
			return
		}

		// Check if the target should be banned.
		banned, err := core.CheckAndBan(database, req.TargetAgentID, cfg.ReportBanThreshold)
		if err != nil {
			// Report was created, but ban check failed. Return success with warning.
			c.JSON(http.StatusCreated, gin.H{
				"report_id": reportID,
				"status":    "reported",
				"message":   "Report filed successfully. Ban check encountered an error.",
			})
			return
		}

		response := gin.H{
			"report_id": reportID,
			"status":    "reported",
			"message":   "Report filed successfully",
		}

		if banned {
			response["status"] = "reported_and_banned"
			response["message"] = "Report filed. Target agent has been banned due to excessive reports."
			response["target_banned"] = true
		}

		c.JSON(http.StatusCreated, response)
	}
}
