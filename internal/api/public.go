package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ListPublicAgents handles GET /api/v1/public/agents.
// Returns a paginated list of active agents with public info only.
func ListPublicAgents(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

		if page < 1 {
			page = 1
		}
		if perPage < 1 || perPage > 100 {
			perPage = 20
		}

		offset := (page - 1) * perPage

		// Count total active agents.
		var totalAgents int
		err := database.QueryRow("SELECT COUNT(*) FROM agents WHERE status = 'active'").Scan(&totalAgents)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to count agents",
			})
			return
		}

		// Fetch agents with task count.
		rows, err := database.Query(
			`SELECT a.id, a.display_name, a.public_bio, a.last_heartbeat, a.created_at,
			        (SELECT COUNT(*) FROM tasks t WHERE t.agent_id = a.id AND t.status = 'active') as task_count
			 FROM agents a
			 WHERE a.status = 'active'
			 ORDER BY a.created_at DESC
			 LIMIT ? OFFSET ?`,
			perPage, offset,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to list agents",
			})
			return
		}
		defer rows.Close()

		type PublicAgent struct {
			ID            string `json:"id"`
			DisplayName   string `json:"display_name"`
			PublicBio     string `json:"public_bio"`
			TaskCount     int    `json:"task_count"`
			LastHeartbeat string `json:"last_heartbeat"`
			CreatedAt     string `json:"created_at"`
		}

		var agents []PublicAgent
		for rows.Next() {
			var a PublicAgent
			var hb sql.NullString
			if err := rows.Scan(&a.ID, &a.DisplayName, &a.PublicBio, &hb, &a.CreatedAt, &a.TaskCount); err == nil {
				if hb.Valid {
					a.LastHeartbeat = hb.String
				}
				agents = append(agents, a)
			}
		}

		if agents == nil {
			agents = []PublicAgent{}
		}

		totalPages := (totalAgents + perPage - 1) / perPage

		c.JSON(http.StatusOK, gin.H{
			"agents":      agents,
			"total":       totalAgents,
			"page":        page,
			"per_page":    perPage,
			"total_pages": totalPages,
		})
	}
}

// GetPublicAgent handles GET /api/v1/public/agents/:id.
// Returns a single agent's public profile with their active tasks.
func GetPublicAgent(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID := c.Param("id")

		var displayName, publicBio, status, createdAt string
		var lastHeartbeat sql.NullString
		err := database.QueryRow(
			"SELECT display_name, public_bio, status, last_heartbeat, created_at FROM agents WHERE id = ? AND status = 'active'",
			agentID,
		).Scan(&displayName, &publicBio, &status, &lastHeartbeat, &createdAt)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "agent_not_found",
				"message": "Agent not found or is not active",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to look up agent",
			})
			return
		}

		// Fetch public task info.
		rows, err := database.Query(
			"SELECT id, mode, type, title, created_at FROM tasks WHERE agent_id = ? AND status = 'active'",
			agentID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to fetch tasks",
			})
			return
		}
		defer rows.Close()

		type PublicTask struct {
			ID        string `json:"id"`
			Mode      string `json:"mode"`
			Type      string `json:"type"`
			Title     string `json:"title"`
			CreatedAt string `json:"created_at"`
		}

		var tasks []PublicTask
		for rows.Next() {
			var t PublicTask
			if err := rows.Scan(&t.ID, &t.Mode, &t.Type, &t.Title, &t.CreatedAt); err != nil {
				continue
			}
			tasks = append(tasks, t)
		}

		if tasks == nil {
			tasks = []PublicTask{}
		}

		hb := ""
		if lastHeartbeat.Valid {
			hb = lastHeartbeat.String
		}

		c.JSON(http.StatusOK, gin.H{
			"agent": gin.H{
				"id":             agentID,
				"display_name":   displayName,
				"public_bio":     publicBio,
				"last_heartbeat": hb,
				"created_at":     createdAt,
			},
			"tasks": tasks,
		})
	}
}

// GetPublicTask handles GET /api/v1/public/tasks/:id.
// Returns a single task's public info with agent details, for shareable task links.
func GetPublicTask(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID := c.Param("id")

		var id, agentID, mode, taskType, title, createdAt string
		var agentDisplayName, agentPublicBio string
		err := database.QueryRow(
			`SELECT t.id, t.agent_id, t.mode, t.type, t.title, t.created_at,
			        a.display_name, a.public_bio
			 FROM tasks t
			 JOIN agents a ON t.agent_id = a.id
			 WHERE (t.id = ? OR t.task_id = ?) AND t.status = 'active' AND a.status = 'active'`,
			taskID, taskID,
		).Scan(&id, &agentID, &mode, &taskType, &title, &createdAt, &agentDisplayName, &agentPublicBio)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "task_not_found",
				"message": "Task not found or is not active",
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

		c.JSON(http.StatusOK, gin.H{
			"task": gin.H{
				"id":         id,
				"mode":       mode,
				"type":       taskType,
				"title":      title,
				"created_at": createdAt,
			},
			"agent": gin.H{
				"id":           agentID,
				"display_name": agentDisplayName,
				"public_bio":   agentPublicBio,
			},
		})
	}
}

// GetPublicStats handles GET /api/v1/public/stats.
// Returns platform-wide statistics.
func GetPublicStats(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var totalAgents, activeAgents24h, totalTasks, totalConversations int
		var beaconTasks, radarTasks int

		// Total active agents.
		_ = database.QueryRow("SELECT COUNT(*) FROM agents WHERE status = 'active'").Scan(&totalAgents)

		// Active agents in the last 24 hours.
		cutoff := time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
		_ = database.QueryRow(
			"SELECT COUNT(*) FROM agents WHERE status = 'active' AND last_heartbeat >= ?",
			cutoff,
		).Scan(&activeAgents24h)

		// Total active tasks.
		_ = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'active'").Scan(&totalTasks)

		// Tasks by mode.
		_ = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'active' AND mode = 'beacon'").Scan(&beaconTasks)
		_ = database.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'active' AND mode = 'radar'").Scan(&radarTasks)

		// Distinct task types.
		typeRows, err := database.Query("SELECT type, COUNT(*) as cnt FROM tasks WHERE status = 'active' GROUP BY type ORDER BY cnt DESC")
		tasksByType := make(map[string]int)
		if err == nil {
			defer typeRows.Close()
			for typeRows.Next() {
				var taskType string
				var cnt int
				if err := typeRows.Scan(&taskType, &cnt); err == nil {
					tasksByType[taskType] = cnt
				}
			}
		}

		// Total conversations.
		_ = database.QueryRow("SELECT COUNT(*) FROM conversations").Scan(&totalConversations)

		// Total matches (conversations that were accepted or beyond).
		var totalMatches int
		_ = database.QueryRow(
			"SELECT COUNT(*) FROM conversations WHERE state NOT IN ('pending_acceptance', 'concluded_no_match')",
		).Scan(&totalMatches)

		c.JSON(http.StatusOK, gin.H{
			"total_agents":       totalAgents,
			"active_agents_24h":  activeAgents24h,
			"total_tasks":        totalTasks,
			"tasks_by_mode": gin.H{
				"beacon": beaconTasks,
				"radar":  radarTasks,
			},
			"tasks_by_type":      tasksByType,
			"total_conversations": totalConversations,
			"total_matches":      totalMatches,
		})
	}
}
