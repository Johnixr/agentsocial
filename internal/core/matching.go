package core

import (
	"database/sql"
	"fmt"
	"sort"
)

// MatchResult holds information about a matched task.
type MatchResult struct {
	AgentID     string  `json:"agent_id"`
	TaskID      string  `json:"task_id"`
	DisplayName string  `json:"display_name"`
	PublicBio   string  `json:"public_bio"`
	Mode        string  `json:"mode"`
	Type        string  `json:"type"`
	Title       string  `json:"title"`
	Score       float64 `json:"score"`
}

// FindMatches loads all active task embeddings from the database, computes cosine similarity
// with the query embedding, and returns the top matches above the minimum score.
// heartbeatCutoff filters out agents that haven't been active since the given time (RFC3339).
// Pass empty string to skip the filter.
func FindMatches(db *sql.DB, queryEmbedding []float32, excludeAgentID string, maxResults int, minScore float64, heartbeatCutoff string) ([]MatchResult, error) {
	query := `
		SELECT t.id, t.agent_id, t.mode, t.type, t.title, a.display_name, a.public_bio, te.embedding
		FROM tasks t
		JOIN agents a ON t.agent_id = a.id
		JOIN task_embeddings te ON t.id = te.task_id
		WHERE a.status = 'active'
		  AND t.status = 'active'
		  AND t.agent_id != ?
	`

	args := []interface{}{excludeAgentID}
	if heartbeatCutoff != "" {
		query += `  AND a.last_heartbeat >= ?
	`
		args = append(args, heartbeatCutoff)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var results []MatchResult

	for rows.Next() {
		var (
			taskID       string
			agentID      string
			mode         string
			taskType     string
			title        string
			displayName  string
			publicBio    string
			embeddingRaw []byte
		)

		if err := rows.Scan(&taskID, &agentID, &mode, &taskType, &title, &displayName, &publicBio, &embeddingRaw); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		taskEmbedding := BytesToEmbedding(embeddingRaw)
		score := CosineSimilarity(queryEmbedding, taskEmbedding)

		if score >= minScore {
			results = append(results, MatchResult{
				AgentID:     agentID,
				TaskID:      taskID,
				DisplayName: displayName,
				PublicBio:   publicBio,
				Mode:        mode,
				Type:        taskType,
				Title:       title,
				Score:       score,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Sort by score descending.
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Limit to maxResults.
	if len(results) > maxResults {
		results = results[:maxResults]
	}

	return results, nil
}
