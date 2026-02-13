package core

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
)

const openAIEmbeddingsURL = "https://api.openai.com/v1/embeddings"

// EmbeddingClient handles communication with the OpenAI Embeddings API.
type EmbeddingClient struct {
	apiKey     string
	model      string
	dimensions int
	httpClient *http.Client
}

// NewEmbeddingClient creates a new client for generating text embeddings.
func NewEmbeddingClient(apiKey, model string, dimensions int) *EmbeddingClient {
	return &EmbeddingClient{
		apiKey:     apiKey,
		model:      model,
		dimensions: dimensions,
		httpClient: &http.Client{},
	}
}

// embeddingRequest is the request body sent to the OpenAI API.
type embeddingRequest struct {
	Model      string `json:"model"`
	Input      string `json:"input"`
	Dimensions int    `json:"dimensions"`
}

// embeddingResponse is the response from the OpenAI API.
type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// GetEmbedding calls the OpenAI Embeddings API and returns the embedding vector.
func (ec *EmbeddingClient) GetEmbedding(text string) ([]float32, error) {
	if ec.apiKey == "" {
		return nil, fmt.Errorf("OpenAI API key is not configured")
	}

	reqBody := embeddingRequest{
		Model:      ec.model,
		Input:      text,
		Dimensions: ec.dimensions,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, openAIEmbeddingsURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ec.apiKey)

	resp, err := ec.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if embResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", embResp.Error.Message)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	// Convert float64 to float32.
	f64 := embResp.Data[0].Embedding
	embedding := make([]float32, len(f64))
	for i, v := range f64 {
		embedding[i] = float32(v)
	}

	return embedding, nil
}

// EmbeddingToBytes converts a float32 slice to a binary blob for storage.
func EmbeddingToBytes(embedding []float32) []byte {
	buf := new(bytes.Buffer)
	for _, v := range embedding {
		_ = binary.Write(buf, binary.LittleEndian, v)
	}
	return buf.Bytes()
}

// BytesToEmbedding converts a binary blob back to a float32 slice.
func BytesToEmbedding(data []byte) []float32 {
	count := len(data) / 4
	embedding := make([]float32, count)
	reader := bytes.NewReader(data)
	for i := 0; i < count; i++ {
		_ = binary.Read(reader, binary.LittleEndian, &embedding[i])
	}
	return embedding
}

// CosineSimilarity computes the cosine similarity between two vectors.
// Returns a value between -1.0 and 1.0.
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		ai := float64(a[i])
		bi := float64(b[i])
		dotProduct += ai * bi
		normA += ai * ai
		normB += bi * bi
	}

	denominator := math.Sqrt(normA) * math.Sqrt(normB)
	if denominator == 0 {
		return 0.0
	}

	return dotProduct / denominator
}
