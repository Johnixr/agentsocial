package api

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"agentsocial/internal/config"
	"agentsocial/internal/core"

	"github.com/gin-gonic/gin"
)

// ScanRequest is the body for POST /api/v1/scan.
type ScanRequest struct {
	TaskID   string   `json:"task_id" binding:"required"`
	Keywords []string `json:"keywords" binding:"required"`
}

// Scan handles POST /api/v1/scan.
// It finds matching tasks based on keyword embeddings.
// Beacon tasks are returned to Radar agents, and vice versa.
func Scan(database *sql.DB, cfg *config.Config, embClient *core.EmbeddingClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent, ok := getAgent(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		var req ScanRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request body: " + err.Error(),
			})
			return
		}

		// Look up the agent's task to determine its mode.
		var taskMode string
		err := database.QueryRow(
			"SELECT mode FROM tasks WHERE (id = ? OR task_id = ?) AND agent_id = ?",
			req.TaskID, req.TaskID, agent.ID,
		).Scan(&taskMode)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "task_not_found",
				"message": "Task not found or does not belong to this agent",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to look up task",
			})
			return
		}

		// Compute embedding from keywords.
		keywordsText := strings.Join(req.Keywords, " ")
		queryEmbedding, err := embClient.GetEmbedding(keywordsText)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "embedding_error",
				"message": "Failed to compute embedding: " + err.Error(),
			})
			return
		}

		// Find matches.
		matches, err := core.FindMatches(database, queryEmbedding, agent.ID, cfg.ScanMaxResults, cfg.ScanMinScore)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "matching_error",
				"message": "Failed to find matches: " + err.Error(),
			})
			return
		}

		// Filter by complementary mode: radar sees beacons, beacons see radars.
		var complementaryMode string
		if taskMode == "radar" {
			complementaryMode = "beacon"
		} else {
			complementaryMode = "radar"
		}

		var filtered []core.MatchResult
		for _, m := range matches {
			if m.Mode == complementaryMode {
				filtered = append(filtered, m)
			}
		}

		if filtered == nil {
			filtered = []core.MatchResult{}
		}

		// Suggest next scan time (60 seconds from now).
		nextScanAfter := time.Now().UTC().Add(60 * time.Second).Format(time.RFC3339)

		c.JSON(http.StatusOK, gin.H{
			"matches":         filtered,
			"next_scan_after": nextScanAfter,
		})
	}
}
