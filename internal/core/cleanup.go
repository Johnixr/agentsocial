package core

import (
	"database/sql"
	"log"
	"time"

	"agentsocial/internal/config"
)

// StartCleanupTicker runs periodic cleanup tasks every hour.
func StartCleanupTicker(db *sql.DB, cfg *config.Config) {
	ticker := time.NewTicker(1 * time.Hour)
	// Run once immediately on startup.
	runCleanup(db, cfg)

	for range ticker.C {
		runCleanup(db, cfg)
	}
}

func runCleanup(db *sql.DB, cfg *config.Config) {
	now := time.Now().UTC()

	hibernated := hibernateInactiveAgents(db, now, cfg.AgentInactiveDays)
	expired := expirePendingConversations(db, now, cfg.ConversationTimeoutDays)
	cleaned := cleanOrphanMessages(db, now, cfg.MessageTTLDays)

	if hibernated > 0 || expired > 0 || cleaned > 0 {
		log.Printf("Cleanup: hibernated %d agents, expired %d conversations, cleaned %d messages",
			hibernated, expired, cleaned)
	}
}

// hibernateInactiveAgents marks agents with no heartbeat in N days as inactive,
// along with their tasks. Returns number of agents hibernated.
func hibernateInactiveAgents(db *sql.DB, now time.Time, inactiveDays int) int64 {
	if inactiveDays <= 0 {
		return 0
	}

	cutoff := now.AddDate(0, 0, -inactiveDays).Format(time.RFC3339)

	// Mark agents inactive if they haven't sent a heartbeat since cutoff.
	// Only affect currently active agents. Also skip agents that never heartbeated
	// but were created recently (give them a grace period equal to inactiveDays).
	result, err := db.Exec(
		`UPDATE agents SET status = 'inactive'
		 WHERE status = 'active'
		   AND ((last_heartbeat IS NOT NULL AND last_heartbeat < ?)
		     OR (last_heartbeat IS NULL AND created_at < ?))`,
		cutoff, cutoff,
	)
	if err != nil {
		log.Printf("Cleanup error (hibernate agents): %v", err)
		return 0
	}

	count, _ := result.RowsAffected()
	if count > 0 {
		// Also hibernate their active tasks.
		_, err = db.Exec(
			`UPDATE tasks SET status = 'inactive', updated_at = ?
			 WHERE agent_id IN (SELECT id FROM agents WHERE status = 'inactive')
			   AND status = 'active'`,
			now.Format(time.RFC3339),
		)
		if err != nil {
			log.Printf("Cleanup error (hibernate tasks): %v", err)
		}
	}

	return count
}

// expirePendingConversations marks pending_acceptance conversations as expired
// if they've been waiting longer than N days. Returns count expired.
func expirePendingConversations(db *sql.DB, now time.Time, timeoutDays int) int64 {
	if timeoutDays <= 0 {
		return 0
	}

	cutoff := now.AddDate(0, 0, -timeoutDays).Format(time.RFC3339)

	result, err := db.Exec(
		`UPDATE conversations SET state = 'expired', updated_at = ?
		 WHERE state = 'pending_acceptance' AND created_at < ?`,
		now.Format(time.RFC3339), cutoff,
	)
	if err != nil {
		log.Printf("Cleanup error (expire conversations): %v", err)
		return 0
	}

	count, _ := result.RowsAffected()
	return count
}

// cleanOrphanMessages deletes messages older than N days that are addressed to
// inactive agents (they'll never pick them up). Returns count deleted.
func cleanOrphanMessages(db *sql.DB, now time.Time, ttlDays int) int64 {
	if ttlDays <= 0 {
		return 0
	}

	cutoff := now.AddDate(0, 0, -ttlDays).Format(time.RFC3339)

	result, err := db.Exec(
		`DELETE FROM message_queue
		 WHERE created_at < ?
		   AND to_agent_id IN (SELECT id FROM agents WHERE status = 'inactive')`,
		cutoff,
	)
	if err != nil {
		log.Printf("Cleanup error (clean messages): %v", err)
		return 0
	}

	count, _ := result.RowsAffected()
	return count
}
