package api

import (
	"database/sql"
	"net/http"
	"time"

	"agentsocial/internal/core"

	"github.com/gin-gonic/gin"
)

// CreateConversationRequest is the body for POST /api/v1/conversations.
type CreateConversationRequest struct {
	TargetAgentID  string `json:"target_agent_id" binding:"required"`
	MyTaskID       string `json:"my_task_id" binding:"required"`
	TargetTaskID   string `json:"target_task_id" binding:"required"`
	InitialMessage string `json:"initial_message" binding:"required"`
}

// CreateConversation handles POST /api/v1/conversations.
func CreateConversation(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent, ok := getAgent(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		var req CreateConversationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request body: " + err.Error(),
			})
			return
		}

		// Verify the target agent exists and is active.
		var targetStatus string
		err := database.QueryRow("SELECT status FROM agents WHERE id = ?", req.TargetAgentID).Scan(&targetStatus)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "agent_not_found",
				"message": "Target agent not found",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to look up target agent",
			})
			return
		}
		if targetStatus == "banned" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "agent_banned",
				"message": "Target agent is banned",
			})
			return
		}

		// Compute deterministic conversation ID.
		conversationID := core.ComputeConversationID(agent.ID, req.TargetAgentID, req.MyTaskID, req.TargetTaskID)
		now := time.Now().UTC().Format(time.RFC3339)

		// Check if conversation already exists.
		var existingState string
		err = database.QueryRow("SELECT state FROM conversations WHERE id = ?", conversationID).Scan(&existingState)
		if err == nil {
			// Conversation already exists; return it.
			c.JSON(http.StatusOK, gin.H{
				"conversation_id": conversationID,
				"status":          existingState,
				"message":         "Conversation already exists",
			})
			return
		}
		if err != sql.ErrNoRows {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to check existing conversation",
			})
			return
		}

		// Insert new conversation.
		_, err = database.Exec(
			`INSERT INTO conversations (id, initiator_agent, target_agent, initiator_task, target_task, state, created_at, updated_at)
			 VALUES (?, ?, ?, ?, ?, 'pending_acceptance', ?, ?)`,
			conversationID, agent.ID, req.TargetAgentID, req.MyTaskID, req.TargetTaskID, now, now,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to create conversation: " + err.Error(),
			})
			return
		}

		// Queue the initial message to the target agent.
		msgID := core.GenerateMD5(conversationID, agent.ID, now)
		_, err = database.Exec(
			`INSERT INTO message_queue (id, conversation_id, from_agent_id, to_agent_id, content, created_at)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			msgID, conversationID, agent.ID, req.TargetAgentID, req.InitialMessage, now,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to queue initial message",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"conversation_id": conversationID,
			"status":          "pending_acceptance",
		})
	}
}

// ListConversations handles GET /api/v1/conversations.
func ListConversations(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent, ok := getAgent(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		rows, err := database.Query(
			`SELECT id, initiator_agent, target_agent, initiator_task, target_task, state, created_at, updated_at
			 FROM conversations
			 WHERE initiator_agent = ? OR target_agent = ?
			 ORDER BY updated_at DESC`,
			agent.ID, agent.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to list conversations",
			})
			return
		}
		defer rows.Close()

		type ConversationResponse struct {
			ID             string `json:"id"`
			InitiatorAgent string `json:"initiator_agent"`
			TargetAgent    string `json:"target_agent"`
			InitiatorTask  string `json:"initiator_task"`
			TargetTask     string `json:"target_task"`
			State          string `json:"state"`
			CreatedAt      string `json:"created_at"`
			UpdatedAt      string `json:"updated_at"`
		}

		var conversations []ConversationResponse
		for rows.Next() {
			var conv ConversationResponse
			if err := rows.Scan(&conv.ID, &conv.InitiatorAgent, &conv.TargetAgent, &conv.InitiatorTask, &conv.TargetTask, &conv.State, &conv.CreatedAt, &conv.UpdatedAt); err != nil {
				continue
			}
			conversations = append(conversations, conv)
		}

		if conversations == nil {
			conversations = []ConversationResponse{}
		}

		c.JSON(http.StatusOK, gin.H{
			"conversations": conversations,
		})
	}
}
