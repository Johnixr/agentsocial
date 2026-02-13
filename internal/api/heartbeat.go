package api

import (
	"database/sql"
	"net/http"
	"time"

	"agentsocial/internal/core"

	"github.com/gin-gonic/gin"
)

// OutboundMessage represents a message being sent during a heartbeat.
type OutboundMessage struct {
	ConversationID string `json:"conversation_id" binding:"required"`
	Message        string `json:"message" binding:"required"`
}

// HeartbeatRequest is the body for POST /api/v1/heartbeat.
type HeartbeatRequest struct {
	Outbound []OutboundMessage `json:"outbound"`
}

// InboundMessage represents a message received during a heartbeat pull.
type InboundMessage struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	FromAgentID    string `json:"from_agent_id"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}

// Notification represents a notification delivered during heartbeat.
type Notification struct {
	Type           string `json:"type"`
	ConversationID string `json:"conversation_id,omitempty"`
	FromAgentID    string `json:"from_agent_id,omitempty"`
	Message        string `json:"message"`
	CreatedAt      string `json:"created_at"`
}

// Heartbeat handles POST /api/v1/heartbeat.
// It sends outbound messages, pulls inbound messages, and deletes them (relay only).
func Heartbeat(database *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		agent, ok := getAgent(c)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		var req HeartbeatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// It's okay to have an empty body; just pull messages.
			req = HeartbeatRequest{}
		}

		now := time.Now().UTC().Format(time.RFC3339)

		// Process outbound messages.
		for _, out := range req.Outbound {
			// Look up the conversation to find the other agent.
			var initiatorAgent, targetAgent, convState string
			err := database.QueryRow(
				"SELECT initiator_agent, target_agent, state FROM conversations WHERE id = ?",
				out.ConversationID,
			).Scan(&initiatorAgent, &targetAgent, &convState)
			if err != nil {
				continue
			}

			// Determine the recipient.
			var toAgentID string
			if agent.ID == initiatorAgent {
				toAgentID = targetAgent
			} else if agent.ID == targetAgent {
				toAgentID = initiatorAgent
			} else {
				continue
			}

			// Auto-accept: if target replies, move conversation to active.
			if convState == "pending_acceptance" && agent.ID == targetAgent {
				_, _ = database.Exec(
					"UPDATE conversations SET state = 'active', updated_at = ? WHERE id = ?",
					now, out.ConversationID,
				)
			}

			// Insert into message queue.
			msgID := core.GenerateMD5(out.ConversationID, agent.ID, now, out.Message)
			_, _ = database.Exec(
				`INSERT INTO message_queue (id, conversation_id, from_agent_id, to_agent_id, content, created_at)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				msgID, out.ConversationID, agent.ID, toAgentID, out.Message, now,
			)
		}

		// Pull all inbound messages for this agent.
		rows, err := database.Query(
			`SELECT id, conversation_id, from_agent_id, content, created_at
			 FROM message_queue
			 WHERE to_agent_id = ?
			 ORDER BY created_at ASC`,
			agent.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "Failed to pull messages",
			})
			return
		}
		defer rows.Close()

		var inbound []InboundMessage
		var messageIDs []string
		for rows.Next() {
			var msg InboundMessage
			if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.FromAgentID, &msg.Content, &msg.CreatedAt); err != nil {
				continue
			}
			inbound = append(inbound, msg)
			messageIDs = append(messageIDs, msg.ID)
		}

		// Delete pulled messages immediately -- relay only, privacy first!
		for _, id := range messageIDs {
			_, _ = database.Exec("DELETE FROM message_queue WHERE id = ?", id)
		}

		// Pull notifications: new conversation requests where this agent is the target.
		notifRows, err := database.Query(
			`SELECT id, initiator_agent, created_at
			 FROM conversations
			 WHERE target_agent = ? AND state = 'pending_acceptance'
			 ORDER BY created_at DESC`,
			agent.ID,
		)
		if err != nil {
			// Non-fatal: return empty notifications.
			notifRows = nil
		}

		var notifications []Notification
		if notifRows != nil {
			defer notifRows.Close()
			for notifRows.Next() {
				var convID, fromAgent, createdAt string
				if err := notifRows.Scan(&convID, &fromAgent, &createdAt); err != nil {
					continue
				}
				notifications = append(notifications, Notification{
					Type:           "conversation_request",
					ConversationID: convID,
					FromAgentID:    fromAgent,
					Message:        "New conversation request",
					CreatedAt:      createdAt,
				})
			}
		}

		if inbound == nil {
			inbound = []InboundMessage{}
		}
		if notifications == nil {
			notifications = []Notification{}
		}

		c.JSON(http.StatusOK, gin.H{
			"inbound":       inbound,
			"notifications": notifications,
		})
	}
}
