package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration values.
type Config struct {
	Port                      string
	BaseURL                   string
	SQLitePath                string
	OpenAIAPIKey              string
	OpenAIEmbeddingModel      string
	OpenAIEmbeddingDimensions int
	RegistrationDailyLimit    int
	ScanMaxResults            int
	ScanMinScore              float64
	ReportBanThreshold        int
	AdminEmail                string
	TokenLength               int
	AgentInactiveDays         int
	ConversationTimeoutDays   int
	MessageTTLDays            int
}

// Load reads configuration from environment variables (and .env file if present).
func Load() *Config {
	// Attempt to load .env file; ignore error if it doesn't exist.
	_ = godotenv.Load()

	cfg := &Config{
		Port:                     getEnv("PORT", "8080"),
		BaseURL:                  getEnv("BASE_URL", "http://localhost:8080"),
		SQLitePath:               getEnv("SQLITE_PATH", "./data/agentsocial.db"),
		OpenAIAPIKey:             getEnv("OPENAI_API_KEY", ""),
		OpenAIEmbeddingModel:     getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-large"),
		OpenAIEmbeddingDimensions: getEnvInt("OPENAI_EMBEDDING_DIMENSIONS", 256),
		RegistrationDailyLimit:   getEnvInt("REGISTRATION_DAILY_LIMIT", 2),
		ScanMaxResults:           getEnvInt("SCAN_MAX_RESULTS", 10),
		ScanMinScore:             getEnvFloat("SCAN_MIN_SCORE", 0.7),
		ReportBanThreshold:       getEnvInt("REPORT_BAN_THRESHOLD", 3),
		AdminEmail:               getEnv("ADMIN_EMAIL", "admin@plaw.social"),
		TokenLength:              getEnvInt("TOKEN_LENGTH", 32),
		AgentInactiveDays:        getEnvInt("AGENT_INACTIVE_DAYS", 30),
		ConversationTimeoutDays:  getEnvInt("CONVERSATION_TIMEOUT_DAYS", 7),
		MessageTTLDays:           getEnvInt("MESSAGE_TTL_DAYS", 7),
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	val := getEnv(key, "")
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvFloat(key string, fallback float64) float64 {
	val := getEnv(key, "")
	if val == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fallback
	}
	return f
}
