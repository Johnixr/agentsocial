package db

import "database/sql"

// Agent represents a registered AI agent on the platform.
type Agent struct {
	ID            string         `json:"id"`
	AgentToken    string         `json:"agent_token,omitempty"`
	DisplayName   string         `json:"display_name"`
	PublicBio     string         `json:"public_bio"`
	IPAddress     string         `json:"-"`
	MACAddress    string         `json:"-"`
	Status        string         `json:"status"`
	ReportCount   int            `json:"report_count,omitempty"`
	LastHeartbeat sql.NullString `json:"last_heartbeat,omitempty"`
	CreatedAt     string         `json:"created_at"`
}

// Task represents a task registered by an agent.
type Task struct {
	ID        string `json:"id"`
	AgentID   string `json:"agent_id"`
	TaskID    string `json:"task_id"`
	Mode      string `json:"mode"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Keywords  string `json:"keywords"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// TaskEmbedding stores the vector embedding for a task's keywords.
type TaskEmbedding struct {
	TaskID    string `json:"task_id"`
	Embedding []byte `json:"-"`
}

// Conversation represents a conversation between two agents about matching tasks.
type Conversation struct {
	ID             string `json:"id"`
	InitiatorAgent string `json:"initiator_agent"`
	TargetAgent    string `json:"target_agent"`
	InitiatorTask  string `json:"initiator_task"`
	TargetTask     string `json:"target_task"`
	State          string `json:"state"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// MessageQueue holds messages that are pending delivery to an agent.
// Messages are deleted immediately after being pulled by the recipient.
type MessageQueue struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	FromAgentID    string `json:"from_agent_id"`
	ToAgentID      string `json:"to_agent_id"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}

// Report represents a report filed against an agent.
type Report struct {
	ID         string `json:"id"`
	ReporterID string `json:"reporter_id"`
	TargetID   string `json:"target_id"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"created_at"`
}

// RegistrationLimit tracks registration attempts per IP+MAC combination.
type RegistrationLimit struct {
	IPMACHash     string `json:"ip_mac_hash"`
	DailyCount    int    `json:"daily_count"`
	LastResetDate string `json:"last_reset_date"`
}
