package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"agentsocial/internal/config"
	"agentsocial/internal/core"
	dbpkg "agentsocial/internal/db"

	"github.com/gin-gonic/gin"
)

// RegisterRequest is the body for POST /api/v1/agents/register.
type RegisterRequest struct {
	DisplayName string        `json:"display_name" binding:"required"`
	PublicBio   string        `json:"public_bio"`
	IPAddress   string        `json:"ip_address" binding:"required"`
	MACAddress  string        `json:"mac_address" binding:"required"`
	Tasks       []TaskRequest `json:"tasks"`
}

// TaskRequest represents a single task in the registration payload.
type TaskRequest struct {
	TaskID   string   `json:"task_id" binding:"required"`
	Mode     string   `json:"mode" binding:"required"`
	Type     string   `json:"type" binding:"required"`
	Title    string   `json:"title" binding:"required"`
	Keywords []string `json:"keywords"`
}

// RegisterAgent handles POST /api/v1/agents/register.
func RegisterAgent(database *sql.DB, cfg *config.Config, embClient *core.EmbeddingClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request body: " + err.Error(),
			})
			return
		}

		// Validate mode values.
		for _, t := range req.Tasks {
			if t.Mode != "beacon" && t.Mode != "radar" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "invalid_mode",
					"message": "Task mode must be 'beacon' or 'radar'",
				})
				return
			}
		}

		now := time.Now().UTC()
		today := now.Format("2006-01-02")

		// Check registration limit.
		ipMACHash := core.GenerateMD5(req.IPAddress, req.MACAddress)

		var dailyCount int
		var lastResetDate string
		err := database.QueryRow(
			"SELECT daily_count, last_reset_date FROM registration_limits WHERE ip_mac_hash = ?",
			ipMACHash,
		).Scan(&dailyCount, &lastResetDate)

		if err == sql.ErrNoRows {
			// First registration from this IP+MAC.
			_, err = database.Exec(
				"INSERT INTO registration_limits (ip_mac_hash, daily_count, last_reset_date) VALUES (?, 0, ?)",
				ipMACHash, today,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_error",
					"message": "Failed to initialize registration limit",
				})
				return
			}
			dailyCount = 0
			lastResetDate = today
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to check registration limit",
			})
			return
		}

		// Reset daily count if the date has changed.
		if lastResetDate != today {
			dailyCount = 0
			_, _ = database.Exec(
				"UPDATE registration_limits SET daily_count = 0, last_reset_date = ? WHERE ip_mac_hash = ?",
				today, ipMACHash,
			)
		}

		if dailyCount >= cfg.RegistrationDailyLimit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "registration_limit_exceeded",
				"message": "Daily registration limit exceeded. Try again tomorrow.",
			})
			return
		}

		// Generate agent credentials.
		agentID := core.GenerateAgentID(req.IPAddress, req.MACAddress)
		agentToken := core.GenerateAgentToken(cfg.TokenLength)
		createdAt := now.Format(time.RFC3339)

		// Insert agent record.
		_, err = database.Exec(
			`INSERT INTO agents (id, agent_token, display_name, public_bio, ip_address, mac_address, status, report_count, created_at)
			 VALUES (?, ?, ?, ?, ?, ?, 'active', 0, ?)`,
			agentID, agentToken, req.DisplayName, req.PublicBio, req.IPAddress, req.MACAddress, createdAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to create agent: " + err.Error(),
			})
			return
		}

		// Insert tasks and compute embeddings.
		type taskMapping struct {
			TaskID     string `json:"task_id"`
			PlatformID string `json:"platform_id"`
			Title      string `json:"title"`
			Mode       string `json:"mode"`
		}
		var taskMappings []taskMapping

		for _, t := range req.Tasks {
			taskID := core.GenerateMD5(agentID, t.TaskID)
			keywordsJSON, _ := json.Marshal(t.Keywords)

			_, err = database.Exec(
				`INSERT INTO tasks (id, agent_id, task_id, mode, type, title, keywords, status, created_at)
				 VALUES (?, ?, ?, ?, ?, ?, ?, 'active', ?)`,
				taskID, agentID, t.TaskID, t.Mode, t.Type, t.Title, string(keywordsJSON), createdAt,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_error",
					"message": "Failed to create task: " + err.Error(),
				})
				return
			}

			taskMappings = append(taskMappings, taskMapping{
				TaskID:     t.TaskID,
				PlatformID: taskID,
				Title:      t.Title,
				Mode:       t.Mode,
			})

			// Compute embedding from keywords.
			if len(t.Keywords) > 0 {
				keywordsText := strings.Join(t.Keywords, " ")
				embedding, embErr := embClient.GetEmbedding(keywordsText)
				if embErr != nil {
					log.Printf("WARNING: Failed to compute embedding for task %s: %v", t.TaskID, embErr)
					continue
				}
				embBytes := core.EmbeddingToBytes(embedding)
				_, _ = database.Exec(
					"INSERT OR REPLACE INTO task_embeddings (task_id, embedding) VALUES (?, ?)",
					taskID, embBytes,
				)
			}
		}

		// Increment registration count.
		_, _ = database.Exec(
			"UPDATE registration_limits SET daily_count = daily_count + 1 WHERE ip_mac_hash = ?",
			ipMACHash,
		)

		c.JSON(http.StatusCreated, gin.H{
			"agent_id":      agentID,
			"agent_token":   agentToken,
			"registered_at": createdAt,
			"tasks":         taskMappings,
		})
	}
}

// UpdateTaskRequest is the body for PUT /api/v1/agents/tasks/:taskId.
type UpdateTaskRequest struct {
	Title    string   `json:"title"`
	Keywords []string `json:"keywords"`
	Status   string   `json:"status"`
}

// UpdateTask handles PUT /api/v1/agents/tasks/:taskId.
func UpdateTask(database *sql.DB, embClient *core.EmbeddingClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent, ok := getAgent(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		taskID := c.Param("taskId")

		// Verify the task belongs to the authenticated agent.
		// Support lookup by internal ID (md5 hash) or by user-provided task_id.
		var existingTask dbpkg.Task
		err := database.QueryRow(
			"SELECT id, agent_id, task_id, mode, type, title, keywords, status, created_at FROM tasks WHERE (id = ? OR task_id = ?) AND agent_id = ?",
			taskID, taskID, agent.ID,
		).Scan(
			&existingTask.ID, &existingTask.AgentID, &existingTask.TaskID,
			&existingTask.Mode, &existingTask.Type, &existingTask.Title,
			&existingTask.Keywords, &existingTask.Status, &existingTask.CreatedAt,
		)
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

		var req UpdateTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request body: " + err.Error(),
			})
			return
		}

		// Apply updates.
		if req.Title != "" {
			existingTask.Title = req.Title
		}
		if req.Status != "" {
			existingTask.Status = req.Status
		}

		keywordsChanged := false
		if req.Keywords != nil {
			keywordsJSON, _ := json.Marshal(req.Keywords)
			existingTask.Keywords = string(keywordsJSON)
			keywordsChanged = true
		}

		_, err = database.Exec(
			"UPDATE tasks SET title = ?, keywords = ?, status = ? WHERE id = ?",
			existingTask.Title, existingTask.Keywords, existingTask.Status, existingTask.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to update task",
			})
			return
		}

		// Recompute embedding if keywords changed.
		if keywordsChanged && req.Keywords != nil && len(req.Keywords) > 0 {
			keywordsText := strings.Join(req.Keywords, " ")
			embedding, embErr := embClient.GetEmbedding(keywordsText)
			if embErr == nil {
				embBytes := core.EmbeddingToBytes(embedding)
				_, _ = database.Exec(
					"INSERT OR REPLACE INTO task_embeddings (task_id, embedding) VALUES (?, ?)",
					existingTask.ID, embBytes,
				)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"task": existingTask,
		})
	}
}

// GetMe handles GET /api/v1/agents/me.
func GetMe(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent, ok := getAgent(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		// Fetch tasks for this agent.
		rows, err := database.Query(
			"SELECT id, agent_id, task_id, mode, type, title, keywords, status, created_at FROM tasks WHERE agent_id = ?",
			agent.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to fetch tasks",
			})
			return
		}
		defer rows.Close()

		var tasks []dbpkg.Task
		for rows.Next() {
			var t dbpkg.Task
			if err := rows.Scan(&t.ID, &t.AgentID, &t.TaskID, &t.Mode, &t.Type, &t.Title, &t.Keywords, &t.Status, &t.CreatedAt); err != nil {
				continue
			}
			tasks = append(tasks, t)
		}

		if tasks == nil {
			tasks = []dbpkg.Task{}
		}

		hb := ""
		if agent.LastHeartbeat.Valid {
			hb = agent.LastHeartbeat.String
		}

		c.JSON(http.StatusOK, gin.H{
			"agent": gin.H{
				"id":             agent.ID,
				"display_name":   agent.DisplayName,
				"public_bio":     agent.PublicBio,
				"status":         agent.Status,
				"report_count":   agent.ReportCount,
				"last_heartbeat": hb,
				"created_at":     agent.CreatedAt,
			},
			"tasks": tasks,
		})
	}
}
