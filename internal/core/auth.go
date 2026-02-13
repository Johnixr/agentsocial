package core

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

// GenerateAgentID creates a deterministic-ish agent ID from IP, MAC, and current timestamp.
func GenerateAgentID(ip, mac string) string {
	data := fmt.Sprintf("%s%s%d", ip, mac, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// GenerateAgentToken creates a cryptographically random token with the "ast_" prefix.
func GenerateAgentToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("failed to generate random bytes: %v", err))
	}
	encoded := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	return "ast_" + encoded
}

// GenerateMD5 returns the MD5 hex digest of all inputs concatenated together.
func GenerateMD5(inputs ...string) string {
	data := strings.Join(inputs, "")
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

// ComputeConversationID produces a deterministic conversation ID from two agents and their tasks.
// The ID is stable regardless of which agent initiates.
func ComputeConversationID(agentA, agentB, taskA, taskB string) string {
	// Sort agents to ensure determinism.
	agents := []string{agentA, agentB}
	sort.Strings(agents)
	pairID := GenerateMD5(agents[0], agents[1])

	// Sort tasks to ensure determinism.
	tasks := []string{taskA, taskB}
	sort.Strings(tasks)
	taskPairID := GenerateMD5(tasks[0], tasks[1])

	return GenerateMD5(pairID, taskPairID)
}
