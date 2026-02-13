package core

import (
	"database/sql"
	"fmt"
)

// CheckAndBan increments the report count for the target agent and bans them
// if the count reaches or exceeds the threshold. Returns true if the agent was banned.
func CheckAndBan(db *sql.DB, targetID string, threshold int) (bool, error) {
	// Increment report count.
	_, err := db.Exec(
		"UPDATE agents SET report_count = report_count + 1 WHERE id = ?",
		targetID,
	)
	if err != nil {
		return false, fmt.Errorf("failed to increment report count: %w", err)
	}

	// Check current report count.
	var reportCount int
	err = db.QueryRow("SELECT report_count FROM agents WHERE id = ?", targetID).Scan(&reportCount)
	if err != nil {
		return false, fmt.Errorf("failed to get report count: %w", err)
	}

	if reportCount >= threshold {
		_, err = db.Exec("UPDATE agents SET status = 'banned' WHERE id = ?", targetID)
		if err != nil {
			return false, fmt.Errorf("failed to ban agent: %w", err)
		}
		return true, nil
	}

	return false, nil
}

// IsAgentBanned checks whether an agent is currently banned.
func IsAgentBanned(db *sql.DB, agentID string) (bool, error) {
	var status string
	err := db.QueryRow("SELECT status FROM agents WHERE id = ?", agentID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check agent status: %w", err)
	}
	return status == "banned", nil
}
