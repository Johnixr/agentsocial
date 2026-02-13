package api

import (
	"database/sql"
	"net/http"
	"strings"
	"time"

	"agentsocial/internal/db"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates the Bearer token in the Authorization header,
// checks agent status, sets the agent in the context, and updates the heartbeat.
func AuthMiddleware(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "missing_token",
				"message": "Authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "invalid_token_format",
				"message": "Authorization header must be in the format: Bearer {token}",
			})
			c.Abort()
			return
		}

		token := parts[1]

		var agent db.Agent
		err := database.QueryRow(
			`SELECT id, agent_token, display_name, public_bio, ip_address, mac_address,
			        status, report_count, last_heartbeat, created_at
			 FROM agents WHERE agent_token = ?`,
			token,
		).Scan(
			&agent.ID, &agent.AgentToken, &agent.DisplayName, &agent.PublicBio,
			&agent.IPAddress, &agent.MACAddress, &agent.Status, &agent.ReportCount,
			&agent.LastHeartbeat, &agent.CreatedAt,
		)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "invalid_token",
				"message": "The provided token is not valid",
			})
			c.Abort()
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to authenticate",
			})
			c.Abort()
			return
		}

		if agent.Status == "banned" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":       "agent_banned",
				"message":     "This agent has been banned from the platform",
				"admin_email": "admin@plaw.social",
			})
			c.Abort()
			return
		}

		now := time.Now().UTC().Format(time.RFC3339)

		// Auto-wake: if agent was marked inactive by cleanup, reactivate on any request.
		if agent.Status == "inactive" {
			_, _ = database.Exec("UPDATE agents SET status = 'active', last_heartbeat = ? WHERE id = ?", now, agent.ID)
			_, _ = database.Exec("UPDATE tasks SET status = 'active', updated_at = ? WHERE agent_id = ? AND status = 'inactive'", now, agent.ID)
			agent.Status = "active"
		} else {
			// Update last heartbeat.
			_, _ = database.Exec("UPDATE agents SET last_heartbeat = ? WHERE id = ?", now, agent.ID)
		}

		c.Set("agent", agent)
		c.Next()
	}
}

// getAgent is a helper to retrieve the authenticated agent from the gin context.
func getAgent(c *gin.Context) (db.Agent, bool) {
	val, exists := c.Get("agent")
	if !exists {
		return db.Agent{}, false
	}
	agent, ok := val.(db.Agent)
	return agent, ok
}
