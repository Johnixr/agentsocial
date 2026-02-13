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

		// Resolve task IDs — accept both internal hash ID and user-provided task_id.
		myTaskInternalID := resolveTaskID(database, req.MyTaskID, agent.ID)
		if myTaskInternalID == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "task_not_found",
				"message": "Your task not found: " + req.MyTaskID,
			})
			return
		}

		targetTaskInternalID := resolveTaskID(database, req.TargetTaskID, req.TargetAgentID)
		if targetTaskInternalID == "" {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "task_not_found",
				"message": "Target task not found: " + req.TargetTaskID,
			})
			return
		}

		// Compute deterministic conversation ID using internal IDs.
		conversationID := core.ComputeConversationID(agent.ID, req.TargetAgentID, myTaskInternalID, targetTaskInternalID)
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
			conversationID, agent.ID, req.TargetAgentID, myTaskInternalID, targetTaskInternalID, now, now,
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

// ConcludeConversationRequest is the body for PUT /api/v1/conversations/:id/conclude.
type ConcludeConversationRequest struct {
	Outcome string `json:"outcome" binding:"required"`
}

// ConcludeConversation handles PUT /api/v1/conversations/:id/conclude.
// Either participant can conclude a conversation with an outcome of "matched" or "no_match".
func ConcludeConversation(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent, ok := getAgent(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		convID := c.Param("id")

		var req ConcludeConversationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_request",
				"message": "Invalid request body: " + err.Error(),
			})
			return
		}

		if req.Outcome != "matched" && req.Outcome != "no_match" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid_outcome",
				"message": "Outcome must be 'matched' or 'no_match'",
			})
			return
		}

		// Verify the conversation exists and the agent is a participant.
		var initiatorAgent, targetAgent, state string
		err := database.QueryRow(
			"SELECT initiator_agent, target_agent, state FROM conversations WHERE id = ?",
			convID,
		).Scan(&initiatorAgent, &targetAgent, &state)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "conversation_not_found",
				"message": "Conversation not found",
			})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to look up conversation",
			})
			return
		}

		if agent.ID != initiatorAgent && agent.ID != targetAgent {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "not_participant",
				"message": "You are not a participant of this conversation",
			})
			return
		}

		// Map outcome to state.
		newState := "concluded_no_match"
		if req.Outcome == "matched" {
			newState = "concluded_matched"
		}

		now := time.Now().UTC().Format(time.RFC3339)
		_, err = database.Exec(
			"UPDATE conversations SET state = ?, updated_at = ? WHERE id = ?",
			newState, now, convID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to conclude conversation",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"conversation_id": convID,
			"state":           newState,
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

// resolveTaskID tries to find a task's internal ID. Accepts either the internal
// hash ID (PK) or the user-provided task_id. Returns empty string if not found.
func resolveTaskID(database *sql.DB, taskID string, agentID string) string {
	// Try as internal ID first (PK lookup).
	var id string
	err := database.QueryRow("SELECT id FROM tasks WHERE id = ?", taskID).Scan(&id)
	if err == nil {
		return id
	}

	// Fallback: try as user-provided task_id scoped to the agent.
	err = database.QueryRow("SELECT id FROM tasks WHERE task_id = ? AND agent_id = ?", taskID, agentID).Scan(&id)
	if err == nil {
		return id
	}

	return ""
}
